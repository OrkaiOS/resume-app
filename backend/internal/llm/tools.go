package llm

import (
	"context"
	"encoding/json"
)

// ToolDefinition describes a tool the LLM can invoke.
// Parameters is a JSON Schema describing the tool's arguments.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ToolCall is a single tool invocation request from the LLM.
// Arguments is the raw JSON arguments string returned by the LLM.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// StreamEventType enumerates the kinds of events emitted during a
// StreamWithTools call.
type StreamEventType string

const (
	// StreamEventText is emitted for each token chunk of the assistant's
	// text response.
	StreamEventText StreamEventType = "text"
	// StreamEventToolCalls is emitted when the LLM finishes a response
	// that requested one or more tool invocations. The assistant will
	// not produce further text until the tool results are fed back.
	StreamEventToolCalls StreamEventType = "tool_calls"
)

// StreamEvent is a single event in a StreamWithTools stream.
// For Type == StreamEventText, Token is set.
// For Type == StreamEventToolCalls, ToolCalls is set.
type StreamEvent struct {
	Type      StreamEventType `json:"type"`
	Token     string          `json:"token,omitempty"`
	ToolCalls []ToolCall      `json:"tool_calls,omitempty"`
}

// ToolRegistry resolves tool definitions and executes tool calls.
// Implementations are owned by the services layer and injected into
// the ChatHandler.
type ToolRegistry interface {
	// Definitions returns the tool definitions advertised to the LLM.
	Definitions() []ToolDefinition
	// Execute runs a single tool call and returns the result as a JSON
	// string (the content the LLM should see as the tool result).
	Execute(ctx context.Context, call ToolCall) (string, error)
}
