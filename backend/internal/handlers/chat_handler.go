package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/llm"
	"github.com/marco/resume-app/internal/services"
)

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:ref(id=61ce6f49-1307-4a1e-8ecf-9c49ce906520)
// @orkai:decision FR-030/FR-032 chat handler is a thin SSE transport adapter. It parses the HTTP request, delegates the agentic tool-calling loop to services.ChatAgentService, and translates AgentEvent values into SSE data lines. All orchestration (LLM round-trips, tool execution, message assembly) lives in the service layer; the handler owns only HTTP/SSE transport.
type ChatHandler struct {
	agent *services.ChatAgentService
}

func NewChatHandler(agent *services.ChatAgentService) *ChatHandler {
	return &ChatHandler{agent: agent}
}

type chatRequest struct {
	Messages      []chatMessage `json:"messages" binding:"required"`
	OpportunityID string        `json:"opportunityId"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatStreamEvent struct {
	Token      string          `json:"token,omitempty"`
	Reasoning  string          `json:"reasoning,omitempty"`
	Done       bool            `json:"done,omitempty"`
	Error      string          `json:"error,omitempty"`
	ToolCall   *chatToolCall   `json:"toolCall,omitempty"`
	ToolResult *chatToolResult `json:"toolResult,omitempty"`
}

type chatToolCall struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Args string `json:"args"`
}

type chatToolResult struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func (h *ChatHandler) Stream(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Failure(ErrCodeValidation, "invalid_request"))
		return
	}

	messages := make([]llm.Message, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = llm.Message{Role: m.Role, Content: m.Content}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	ctx := c.Request.Context()
	err := h.agent.Run(ctx, req.OpportunityID, messages, func(ev services.AgentEvent) error {
		h.writeEvent(c, toChatStreamEvent(ev))
		return nil
	})
	if err != nil {
		h.writeEvent(c, chatStreamEvent{Error: err.Error()})
		return
	}

	h.writeEvent(c, chatStreamEvent{Done: true})
}

func toChatStreamEvent(ev services.AgentEvent) chatStreamEvent {
	switch ev.Type {
	case services.AgentEventText:
		return chatStreamEvent{Token: ev.Token}
	case services.AgentEventReasoning:
		return chatStreamEvent{Reasoning: ev.Token}
	case services.AgentEventToolCall:
		return chatStreamEvent{ToolCall: &chatToolCall{
			ID: ev.ToolCall.ID, Name: ev.ToolCall.Name, Args: ev.ToolCall.Args,
		}}
	case services.AgentEventToolResult:
		return chatStreamEvent{ToolResult: &chatToolResult{
			ID: ev.ToolResult.ID, Output: ev.ToolResult.Output, Error: ev.ToolResult.Error,
		}}
	default:
		return chatStreamEvent{}
	}
}

func (h *ChatHandler) writeEvent(c *gin.Context, event chatStreamEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
	c.Writer.Flush()
}
