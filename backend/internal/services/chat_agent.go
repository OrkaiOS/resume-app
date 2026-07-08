package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/marco/resume-app/internal/llm"
)

// ChatAgentService owns the agentic tool-calling loop. It drives the
// LLM + tool registry, streams text tokens and tool events to the
// caller, and feeds tool results back into the conversation. The HTTP
// handler is responsible only for SSE transport; this service owns the
// orchestration logic.
// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision The agentic loop lives in the service layer, not the handler. The handler is a thin transport adapter that parses HTTP and writes SSE events; this service owns the LLM round-trips, tool execution, and conversation message assembly. Max 5 tool-calling iterations per turn prevents infinite loops.
type ChatAgentService struct {
	client        llm.Client
	promptBuilder PromptBuilder
	tools         llm.ToolRegistry
	maxIterations int
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

// NewChatAgentService builds a ChatAgentService with the standard
// maxIterations of 5.
func NewChatAgentService(client llm.Client, promptBuilder PromptBuilder, tools llm.ToolRegistry) *ChatAgentService {
	return &ChatAgentService{
		client:        client,
		promptBuilder: promptBuilder,
		tools:         tools,
		maxIterations: 5,
	}
}

// Run drives the agentic loop. It calls onEvent for each token, tool
// call, tool result, and a final done event. Returns an error only on
// LLM or unrecoverable failures; tool execution errors are surfaced as
// AgentEventToolResult with Error set, not as a returned error.
func (s *ChatAgentService) Run(ctx context.Context, opportunityID string, messages []llm.Message, onEvent func(AgentEvent) error) error {
	systemPrompt := s.promptBuilder.Build(ctx, opportunityID)

	var tools []llm.ToolDefinition
	if s.tools != nil {
		tools = s.tools.Definitions()
	}

	for i := 0; i < s.maxIterations; i++ {
		var fullText strings.Builder
		var toolCalls []llm.ToolCall

		err := s.client.StreamWithTools(ctx, systemPrompt, messages, tools, func(ev llm.StreamEvent) error {
			switch ev.Type {
			case llm.StreamEventText:
				fullText.WriteString(ev.Token)
				return onEvent(AgentEvent{Type: AgentEventText, Token: ev.Token})
			case llm.StreamEventToolCalls:
				toolCalls = ev.ToolCalls
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("services.ChatAgentService.Run: %w", err)
		}

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

			result := AgentToolResult{ID: tc.ID}
			output, execErr := s.tools.Execute(ctx, tc)
			if execErr != nil {
				result.Error = execErr.Error()
			} else {
				result.Output = output
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
		return nil
	})
}
