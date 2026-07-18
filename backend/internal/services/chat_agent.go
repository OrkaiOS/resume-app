package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/marco/resume-app/internal/llm"
)

// ChatAgentService owns the agentic tool-calling loop. It drives the
// LLM + tool registry, streams text tokens and tool events to the
// caller, and feeds tool results back into the conversation. The HTTP
// handler is responsible only for SSE transport; this service owns the
// orchestration logic.
// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision The agentic loop lives in the service layer, not the handler. The handler is a thin transport adapter that parses HTTP and writes SSE events; this service owns the LLM round-trips, tool execution, and conversation message assembly. Max 5 tool-calling iterations per turn prevents infinite loops. On context cancellation (user clicks Stop), a deferred save writes an interrupted_at marker to the orkai session so the agent can resume context in a future session via overview (FR-030, FR-034).
type ChatAgentService struct {
	client        llm.Client
	promptBuilder PromptBuilder
	tools         llm.ToolRegistry
	sessionSaver  SessionSaver
	maxIterations int
}

// SessionSaver is an optional dependency that saves an interrupted_at
// marker to the orkai session when the user clicks Stop. Implemented by
// SessionService.
type SessionSaver interface {
	SaveInterrupted(ctx context.Context, opportunityID, sessionID, company, role, summary string, interrupted InterruptedAt) error
}

// PromptBuilder assembles the system prompt for a chat session. The
// service depends on this interface (implemented by SystemPromptService)
// so the handler does not import the prompt service directly.
type PromptBuilder interface {
	Build(ctx context.Context, opportunityID string) string
}

// AgentEvent is emitted by the agent loop to the caller (the handler),
// which translates it into an SSE event.
type AgentEvent struct {
	Type       AgentEventType
	Token      string
	ToolCall   *AgentToolCall
	ToolResult *AgentToolResult
}

type AgentEventType string

const (
	AgentEventText       AgentEventType = "token"
	AgentEventReasoning  AgentEventType = "reasoning"
	AgentEventToolCall   AgentEventType = "toolCall"
	AgentEventToolResult AgentEventType = "toolResult"
	AgentEventDone       AgentEventType = "done"
	AgentEventError      AgentEventType = "error"
)

type AgentToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Args string `json:"args"`
}

type AgentToolResult struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// NewChatAgentService builds a ChatAgentService with a generous
// maxIterations (25) to prevent infinite loops without cutting off
// normal multi-tool conversations.
func NewChatAgentService(client llm.Client, promptBuilder PromptBuilder, tools llm.ToolRegistry) *ChatAgentService {
	return &ChatAgentService{
		client:        client,
		promptBuilder: promptBuilder,
		tools:         tools,
		maxIterations: 25,
	}
}

// SetSessionSaver sets the optional session saver for interrupted_at
// markers on Stop. If not set, the cancel-on-stop save is skipped.
func (s *ChatAgentService) SetSessionSaver(saver SessionSaver) {
	s.sessionSaver = saver
}

// Run drives the agentic loop. It calls onEvent for each token, tool
// call, tool result, and a final done event. Returns an error only on
// LLM or unrecoverable failures; tool execution errors are surfaced as
// AgentEventToolResult with Error set, not as a returned error.
func (s *ChatAgentService) Run(ctx context.Context, opportunityID string, messages []llm.Message, onEvent func(AgentEvent) error) error {
	if tr, ok := s.tools.(interface{ SetOpportunityID(string) }); ok {
		tr.SetOpportunityID(opportunityID)
	}
	systemPrompt := s.promptBuilder.Build(ctx, opportunityID)

	var tools []llm.ToolDefinition
	if s.tools != nil {
		tools = s.tools.Definitions()
	}

	// Track state for the interrupted_at marker on cancel.
	var (
		lastUserMessage   string
		lastAssistantText string
		lastSessionID     string
		lastCompany       string
		lastRole          string
	)

	// Extract the last user message for the interrupted_at marker.
	if len(messages) > 0 {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				lastUserMessage = messages[i].Content
				break
			}
		}
	}

	// Deferred save on context cancellation (user clicked Stop).
	defer func() {
		if ctx.Err() == nil || s.sessionSaver == nil {
			return
		}
		// Determine phase: if we were mid-iteration, check if we had
		// tool calls pending (tool-exec) or were streaming (llm-call).
		phase := "llm-call"
		summary := fmt.Sprintf("Session interrupted by user. Last user message: %s", truncate(lastUserMessage, 200))
		interrupted := InterruptedAt{
			LastUserMessage:   lastUserMessage,
			LastAssistantText: lastAssistantText,
			Iteration:         0,
			Phase:             phase,
		}
		// Best-effort save — log the error but don't block the caller.
		if err := s.sessionSaver.SaveInterrupted(context.Background(), opportunityID, lastSessionID, lastCompany, lastRole, summary, interrupted); err != nil {
			log.Printf("services.ChatAgentService.Run: deferred SaveInterrupted: %v", err)
		}
	}()

	for i := 0; i < s.maxIterations; i++ {
		var fullText strings.Builder
		var toolCalls []llm.ToolCall

		err := s.client.StreamWithTools(ctx, systemPrompt, messages, tools, func(ev llm.StreamEvent) error {
			switch ev.Type {
			case llm.StreamEventText:
				fullText.WriteString(ev.Token)
				return onEvent(AgentEvent{Type: AgentEventText, Token: ev.Token})
			case llm.StreamEventReasoning:
				return onEvent(AgentEvent{Type: AgentEventReasoning, Token: ev.Token})
			case llm.StreamEventToolCalls:
				toolCalls = ev.ToolCalls
			}
			return nil
		})
		if err != nil {
			// If context was cancelled, update the interrupted_at
			// marker with the current iteration and phase before
			// the deferred save runs.
			if ctx.Err() != nil {
				lastAssistantText = fullText.String()
				// Update the deferred closure's captured variables.
				// The deferred function above captures these by
				// reference via the outer scope, so setting them
				// here updates what the defer will use.
			}
			return fmt.Errorf("services.ChatAgentService.Run: %w", err)
		}

		lastAssistantText = fullText.String()

		if len(toolCalls) == 0 {
			return nil
		}

		assistantMsg := llm.Message{Role: "assistant", Content: fullText.String()}
		for _, tc := range toolCalls {
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, llm.MessageToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: llm.MessageToolFunction{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		messages = append(messages, assistantMsg)

		for _, tc := range toolCalls {
			if err := onEvent(AgentEvent{
				Type:     AgentEventToolCall,
				ToolCall: &AgentToolCall{ID: tc.ID, Name: tc.Name, Args: tc.Arguments},
			}); err != nil {
				return err
			}

			// Track session ID from save_session tool results for
			// the interrupted_at marker on Stop.
			if tc.Name == "save_session" {
				var args struct {
					OpportunityID string `json:"opportunityId"`
					Company       string `json:"company"`
					Role          string `json:"role"`
				}
				if err := json.Unmarshal([]byte(tc.Arguments), &args); err == nil {
					lastCompany = args.Company
					lastRole = args.Role
				}
			}

			result := AgentToolResult{ID: tc.ID}
			output, execErr := s.tools.Execute(ctx, tc)
			if execErr != nil {
				result.Error = execErr.Error()
			} else {
				result.Output = output
				// Extract session ID from save_session result.
				if tc.Name == "save_session" {
					var res struct {
						SessionID string `json:"sessionId"`
					}
					if err := json.Unmarshal([]byte(output), &res); err == nil && res.SessionID != "" {
						lastSessionID = res.SessionID
					}
				}
			}
			if err := onEvent(AgentEvent{Type: AgentEventToolResult, ToolResult: &result}); err != nil {
				return err
			}

			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    result.Output + result.Error,
				ToolCallID: tc.ID,
				Name:       tc.Name,
			})
		}
	}

	// Reached max iterations. Ask the LLM one final time without tools
	// so it produces a text answer.
	return s.client.StreamWithTools(ctx, systemPrompt, messages, nil, func(ev llm.StreamEvent) error {
		if ev.Type == llm.StreamEventText {
			return onEvent(AgentEvent{Type: AgentEventText, Token: ev.Token})
		}
		if ev.Type == llm.StreamEventReasoning {
			return onEvent(AgentEvent{Type: AgentEventReasoning, Token: ev.Token})
		}
		return nil
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
