package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/llm"
	"github.com/marco/resume-app/internal/services"
)

type fakeStreamLLM struct {
	tokens []string
}

func (f *fakeStreamLLM) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "", nil
}

func (f *fakeStreamLLM) Stream(ctx context.Context, systemPrompt string, messages []llm.Message, onToken func(string) error) error {
	for _, token := range f.tokens {
		if err := onToken(token); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStreamLLM) StreamWithTools(ctx context.Context, systemPrompt string, messages []llm.Message, tools []llm.ToolDefinition, onEvent func(llm.StreamEvent) error) error {
	for _, token := range f.tokens {
		if err := onEvent(llm.StreamEvent{Type: llm.StreamEventText, Token: token}); err != nil {
			return err
		}
	}
	return nil
}

type errorStreamLLM struct{}

func (e *errorStreamLLM) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "", nil
}

func (e *errorStreamLLM) Stream(ctx context.Context, systemPrompt string, messages []llm.Message, onToken func(string) error) error {
	return context.Canceled
}

func (e *errorStreamLLM) StreamWithTools(ctx context.Context, systemPrompt string, messages []llm.Message, tools []llm.ToolDefinition, onEvent func(llm.StreamEvent) error) error {
	return context.Canceled
}

// agenticStreamLLM simulates an LLM that returns a tool call on the
// first iteration and final text on the second.
type agenticStreamLLM struct {
	iterations  int
	toolCall    llm.ToolCall
	finalTokens []string
}

func (a *agenticStreamLLM) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "", nil
}

func (a *agenticStreamLLM) Stream(ctx context.Context, systemPrompt string, messages []llm.Message, onToken func(string) error) error {
	return nil
}

func (a *agenticStreamLLM) StreamWithTools(ctx context.Context, systemPrompt string, messages []llm.Message, tools []llm.ToolDefinition, onEvent func(llm.StreamEvent) error) error {
	a.iterations++
	if a.iterations == 1 {
		return onEvent(llm.StreamEvent{Type: llm.StreamEventToolCalls, ToolCalls: []llm.ToolCall{a.toolCall}})
	}
	for _, t := range a.finalTokens {
		if err := onEvent(llm.StreamEvent{Type: llm.StreamEventText, Token: t}); err != nil {
			return err
		}
	}
	return nil
}

type fakePromptBuilder struct{}

func (f *fakePromptBuilder) Build(ctx context.Context, opportunityID string) string {
	return "You are a helpful AI assistant."
}

type fakeToolRegistry struct {
	defs   []llm.ToolDefinition
	result string
	err    error
	calls  []llm.ToolCall
}

func (r *fakeToolRegistry) Definitions() []llm.ToolDefinition { return r.defs }

func (r *fakeToolRegistry) Execute(ctx context.Context, call llm.ToolCall) (string, error) {
	r.calls = append(r.calls, call)
	if r.err != nil {
		return "", r.err
	}
	return r.result, nil
}

func TestChatHandlerStream(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("streams tokens as SSE events", func(t *testing.T) {
		t.Parallel()

		fake := &fakeStreamLLM{tokens: []string{"Hel", "lo", "!"}}
		agent := services.NewChatAgentService(fake, &fakePromptBuilder{}, nil)
		handler := NewChatHandler(agent)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/chat", strings.NewReader(`{"messages":[{"role":"user","content":"Hi"}]}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Stream(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, `data: {"token":"Hel"}`) {
			t.Errorf("expected first token event, got %q", body)
		}
		if !strings.Contains(body, `data: {"token":"lo"}`) {
			t.Errorf("expected second token event, got %q", body)
		}
		if !strings.Contains(body, `data: {"token":"!"}`) {
			t.Errorf("expected third token event, got %q", body)
		}
		if !strings.Contains(body, `"done":true`) {
			t.Errorf("expected done event, got %q", body)
		}
	})

	t.Run("returns 400 on empty body", func(t *testing.T) {
		t.Parallel()

		fake := &fakeStreamLLM{tokens: []string{}}
		agent := services.NewChatAgentService(fake, &fakePromptBuilder{}, nil)
		handler := NewChatHandler(agent)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/chat", strings.NewReader(""))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Stream(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("reports error in stream", func(t *testing.T) {
		t.Parallel()

		agent := services.NewChatAgentService(&errorStreamLLM{}, &fakePromptBuilder{}, nil)
		handler := NewChatHandler(agent)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/chat", strings.NewReader(`{"messages":[{"role":"user","content":"Hi"}]}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Stream(c)

		if !strings.Contains(w.Body.String(), `"error":`) {
			t.Errorf("expected error event, got %q", w.Body.String())
		}
	})

	t.Run("executes tool calls and feeds results back", func(t *testing.T) {
		t.Parallel()

		agentLLM := &agenticStreamLLM{
			toolCall:    llm.ToolCall{ID: "call_1", Name: "get_profile", Arguments: "{}"},
			finalTokens: []string{"Profile", " loaded"},
		}
		registry := &fakeToolRegistry{result: `{"fullName":"Marco"}`}
		agent := services.NewChatAgentService(agentLLM, &fakePromptBuilder{}, registry)
		handler := NewChatHandler(agent)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/chat", strings.NewReader(`{"messages":[{"role":"user","content":"What's my profile?"}]}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Stream(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `"toolCall":{"id":"call_1","name":"get_profile"`) {
			t.Errorf("expected toolCall event, got %q", body)
		}
		if !strings.Contains(body, `"toolResult":{"id":"call_1","output":"{\"fullName\":\"Marco\"}"}`) {
			t.Errorf("expected toolResult event, got %q", body)
		}
		if !strings.Contains(body, `data: {"token":"Profile"}`) {
			t.Errorf("expected final text token, got %q", body)
		}
		if !strings.Contains(body, `"done":true`) {
			t.Errorf("expected done event, got %q", body)
		}
		if len(registry.calls) != 1 {
			t.Errorf("expected 1 tool call, got %d", len(registry.calls))
		}
		if agentLLM.iterations != 2 {
			t.Errorf("expected 2 LLM iterations, got %d", agentLLM.iterations)
		}
	})
}
