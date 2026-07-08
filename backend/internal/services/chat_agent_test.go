package services

import (
	"context"
	"testing"

	"github.com/marco/resume-app/internal/llm"
)

type fakeAgentLLM struct {
	iterations  int
	toolCalls   []llm.ToolCall
	finalTokens []string
}

func (f *fakeAgentLLM) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "", nil
}

func (f *fakeAgentLLM) Stream(ctx context.Context, systemPrompt string, messages []llm.Message, onToken func(string) error) error {
	return nil
}

func (f *fakeAgentLLM) StreamWithTools(ctx context.Context, systemPrompt string, messages []llm.Message, tools []llm.ToolDefinition, onEvent func(llm.StreamEvent) error) error {
	f.iterations++
	if f.iterations <= len(f.toolCalls) {
		return onEvent(llm.StreamEvent{Type: llm.StreamEventToolCalls, ToolCalls: []llm.ToolCall{f.toolCalls[f.iterations-1]}})
	}
	for _, t := range f.finalTokens {
		if err := onEvent(llm.StreamEvent{Type: llm.StreamEventText, Token: t}); err != nil {
			return err
		}
	}
	return nil
}

type fakeRegistry struct {
	result string
	calls  []llm.ToolCall
}

func (r *fakeRegistry) Definitions() []llm.ToolDefinition { return nil }

func (r *fakeRegistry) Execute(ctx context.Context, call llm.ToolCall) (string, error) {
	r.calls = append(r.calls, call)
	return r.result, nil
}

func TestChatAgentServiceRunNoTools(t *testing.T) {
	t.Parallel()
	llm_ := &fakeAgentLLM{finalTokens: []string{"hello"}}
	agent := NewChatAgentService(llm_, &fakeBuilder{}, nil)

	var tokens []string
	err := agent.Run(context.Background(), "", []llm.Message{{Role: "user", Content: "hi"}}, func(ev AgentEvent) error {
		if ev.Type == AgentEventText {
			tokens = append(tokens, ev.Token)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tokens) != 1 || tokens[0] != "hello" {
		t.Errorf("expected tokens [hello], got %v", tokens)
	}
	if llm_.iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", llm_.iterations)
	}
}

func TestChatAgentServiceRunWithToolCall(t *testing.T) {
	t.Parallel()
	llm_ := &fakeAgentLLM{
		toolCalls:   []llm.ToolCall{{ID: "c1", Name: "shell", Arguments: `{"command":"ls"}`}},
		finalTokens: []string{"done"},
	}
	registry := &fakeRegistry{result: `{"stdout":"file.txt"}`}
	agent := NewChatAgentService(llm_, &fakeBuilder{}, registry)

	var toolCalls, toolResults int
	var tokens []string
	err := agent.Run(context.Background(), "", []llm.Message{{Role: "user", Content: "list files"}}, func(ev AgentEvent) error {
		switch ev.Type {
		case AgentEventToolCall:
			toolCalls++
		case AgentEventToolResult:
			toolResults++
		case AgentEventText:
			tokens = append(tokens, ev.Token)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if toolCalls != 1 {
		t.Errorf("expected 1 tool call event, got %d", toolCalls)
	}
	if toolResults != 1 {
		t.Errorf("expected 1 tool result event, got %d", toolResults)
	}
	if len(registry.calls) != 1 {
		t.Errorf("expected 1 registry call, got %d", len(registry.calls))
	}
	if llm_.iterations != 2 {
		t.Errorf("expected 2 LLM iterations, got %d", llm_.iterations)
	}
	if len(tokens) != 1 || tokens[0] != "done" {
		t.Errorf("expected final token [done], got %v", tokens)
	}
}

type fakeBuilder struct{}

func (f *fakeBuilder) Build(ctx context.Context, opportunityID string) string { return "sys" }
