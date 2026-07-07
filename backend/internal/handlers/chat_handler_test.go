package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/llm"
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

type errorStreamLLM struct{}

func (e *errorStreamLLM) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "", nil
}

func (e *errorStreamLLM) Stream(ctx context.Context, systemPrompt string, messages []llm.Message, onToken func(string) error) error {
	return context.Canceled
}

type fakePromptBuilder struct{}

func (f *fakePromptBuilder) Build(ctx context.Context, opportunityID string) string {
	return "You are a helpful AI assistant."
}

func TestChatHandlerStream(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("streams tokens as SSE events", func(t *testing.T) {
		t.Parallel()

		fake := &fakeStreamLLM{tokens: []string{"Hel", "lo", "!"}}
		handler := NewChatHandler(fake, &fakePromptBuilder{})

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
		handler := NewChatHandler(fake, &fakePromptBuilder{})

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

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/chat", strings.NewReader(`{"messages":[{"role":"user","content":"Hi"}]}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := &ChatHandler{client: &errorStreamLLM{}, promptBuilder: &fakePromptBuilder{}}
		handler.Stream(c)

		if !strings.Contains(w.Body.String(), `"error":`) {
			t.Errorf("expected error event, got %q", w.Body.String())
		}
	})
}
