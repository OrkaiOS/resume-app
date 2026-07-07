package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/llm"
)

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:ref(id=61ce6f49-1307-4a1e-8ecf-9c49ce906520)
// @orkai:decision SSE streaming with text/event-stream for FR-030 Chat Interface. Each token chunk is emitted as a JSON event. System prompt is assembled server-side by SystemPromptService from orkai standards + profile + opportunity context per FR-031.
type SystemPromptBuilder interface {
	Build(ctx context.Context, opportunityID string) string
}

type ChatHandler struct {
	client        llm.Client
	promptBuilder SystemPromptBuilder
}

func NewChatHandler(client llm.Client, promptBuilder SystemPromptBuilder) *ChatHandler {
	return &ChatHandler{client: client, promptBuilder: promptBuilder}
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
	Token string `json:"token,omitempty"`
	Done  bool   `json:"done,omitempty"`
	Error string `json:"error,omitempty"`
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

	systemPrompt := h.promptBuilder.Build(c.Request.Context(), req.OpportunityID)

	ctx := c.Request.Context()
	err := h.client.Stream(ctx, systemPrompt, messages, func(token string) error {
		event := chatStreamEvent{Token: token}
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
		c.Writer.Flush()
		return nil
	})
	if err != nil {
		event := chatStreamEvent{Error: err.Error()}
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
		c.Writer.Flush()
		return
	}

	event := chatStreamEvent{Done: true}
	data, _ := json.Marshal(event)
	fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
	c.Writer.Flush()
}
