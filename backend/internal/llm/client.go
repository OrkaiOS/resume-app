package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	// ToolCalls is set on assistant messages that requested tool
	// invocations. Each entry has ID and a Function with Name + Arguments.
	ToolCalls []MessageToolCall `json:"tool_calls,omitempty"`
	// ToolCallID is set on tool result messages (role=="tool") to link
	// the result back to the originating assistant tool call.
	ToolCallID string `json:"tool_call_id,omitempty"`
	// Name is set on tool result messages to identify the tool that was
	// executed.
	Name string `json:"name,omitempty"`
}

// MessageToolCall is a tool call carried by an assistant Message.
type MessageToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function MessageToolFunction `json:"function"`
}

// MessageToolFunction holds the tool name and raw arguments string.
type MessageToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Client interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	Stream(ctx context.Context, systemPrompt string, messages []Message, onToken func(string) error) error
	// StreamWithTools sends messages + tool definitions to the LLM and
	// streams the response. The onEvent callback receives text tokens
	// and a final tool_calls event if the LLM requested tool
	// invocations. The caller is responsible for executing the tool
	// calls and feeding results back into a follow-up call.
	StreamWithTools(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDefinition, onEvent func(StreamEvent) error) error
}

func NewClient(provider, model, apiKey string) Client {
	switch strings.ToLower(provider) {
	case "anthropic":
		return &anthropicClient{model: model, apiKey: apiKey, http: &http.Client{Timeout: 120 * time.Second}}
	default:
		baseURL := "https://api.openai.com/v1"
		isOllama := false
		if provider == "ollama" {
			baseURL = "http://localhost:11434/v1"
			isOllama = true
		}
		return &openaiClient{model: model, apiKey: apiKey, baseURL: baseURL, isOllama: isOllama, http: &http.Client{Timeout: 120 * time.Second}}
	}
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
type openaiClient struct {
	model    string
	apiKey   string
	baseURL  string
	isOllama bool
	http     *http.Client
}

type openaiChatRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Stream   bool         `json:"stream,omitempty"`
	Tools    []openaiTool `json:"tools,omitempty"`
}

// openaiTool is the OpenAI/Ollama wire format for a tool definition:
// {"type":"function","function":{"name":...,"description":...,"parameters":...}}.
type openaiTool struct {
	Type     string         `json:"type"`
	Function ToolDefinition `json:"function"`
}

type openaiStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			Reasoning string `json:"reasoning"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision SSE streaming via OpenAI-compatible API (works for Ollama too). Each SSE data line is parsed for token deltas; onToken is called per chunk.
func (c *openaiClient) Stream(ctx context.Context, systemPrompt string, messages []Message, onToken func(string) error) error {
	allMessages := make([]Message, 0, len(messages)+1)
	allMessages = append(allMessages, Message{Role: "system", Content: systemPrompt})
	allMessages = append(allMessages, messages...)

	body := openaiChatRequest{
		Model:    c.model,
		Messages: allMessages,
		Stream:   true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("llm.openaiClient.Stream: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("llm.openaiClient.Stream: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("llm.openaiClient.Stream: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm.openaiClient.Stream: status %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk openaiStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := onToken(chunk.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

// wrapOpenAITools converts the app's flat ToolDefinition list to the
// OpenAI/Ollama wire format with the type:"function" wrapper.
func wrapOpenAITools(tools []ToolDefinition) []openaiTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]openaiTool, len(tools))
	for i, t := range tools {
		out[i] = openaiTool{Type: "function", Function: t}
	}
	return out
}

// thinkParser processes Ollama content chunks for <think>...</think> blocks,
// emitting reasoning events for text inside think blocks and text events for
// text outside them. Tags split across chunk boundaries are handled by
// retaining a safety buffer.
type thinkParser struct {
	buf     strings.Builder
	inThink bool
}

const tagOpen = "<think>"
const tagClose = "</think>"

// runeBoundary returns the greatest byte index <= n that is a valid
// UTF-8 rune start, ensuring that s[:result] does not split a multi-byte
// character.
func runeBoundary(s string, n int) int {
	for n > 0 && n < len(s) && !utf8.RuneStart(s[n]) {
		n--
	}
	return n
}

// flush emits any residual buffer with no safety margin. At
// end-of-stream there is no next chunk that could complete a split
// </think> or  tags, so the safety margin applied by feed() is wrong
// here — it would retain up to len(tagOpen)-1 = 6 bytes and drop
// the tail of every Ollama response (e.g. the final period).
func (p *thinkParser) flush(onText, onReasoning func(string) error) error {
	text := p.buf.String()
	if text == "" {
		return nil
	}
	p.buf.Reset()
	if p.inThink {
		return onReasoning(text)
	}
	return onText(text)
}

func (p *thinkParser) feed(chunk string, onText, onReasoning func(string) error) error {
	p.buf.WriteString(chunk)
	text := p.buf.String()

	for {
		if p.inThink {
			idx := strings.Index(text, tagClose)
			if idx < 0 {
				safeLen := runeBoundary(text, len(text)-len(tagClose)+1)
				if safeLen > 0 {
					if err := onReasoning(text[:safeLen]); err != nil {
						return err
					}
				}
				start := runeBoundary(text, len(text)-len(tagClose)+1)
				if start < 0 {
					start = 0
				}
				p.buf.Reset()
				p.buf.WriteString(text[start:])
				return nil
			}
			if idx > 0 {
				if err := onReasoning(text[:idx]); err != nil {
					return err
				}
			}
			text = text[idx+len(tagClose):]
			p.inThink = false
		} else {
			idx := strings.Index(text, tagOpen)
			if idx < 0 {
				safeLen := runeBoundary(text, len(text)-len(tagOpen)+1)
				if safeLen > 0 {
					if err := onText(text[:safeLen]); err != nil {
						return err
					}
				}
				start := runeBoundary(text, len(text)-len(tagOpen)+1)
				if start < 0 {
					start = 0
				}
				p.buf.Reset()
				p.buf.WriteString(text[start:])
				return nil
			}
			if idx > 0 {
				if err := onText(text[:idx]); err != nil {
					return err
				}
			}
			text = text[idx+len(tagOpen):]
			p.inThink = true
		}
	}
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision StreamWithTools implements the agentic tool-calling loop primitive. It sends tool definitions to the LLM (wrapped in the OpenAI {type:"function",function:{...}} wire format), streams text tokens to the caller, and accumulates tool_calls by index across SSE chunks. When finish_reason=="tool_calls", the final event carries the complete ToolCalls list. The caller (ChatHandler) executes the tools and feeds results back via a follow-up StreamWithTools call. For Ollama, content deltas are parsed for <think>...</think> blocks that encode reasoning; for OpenAI, the reasoning field in the delta carries o-series reasoning summaries.
func (c *openaiClient) StreamWithTools(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDefinition, onEvent func(StreamEvent) error) error {
	allMessages := make([]Message, 0, len(messages)+1)
	allMessages = append(allMessages, Message{Role: "system", Content: systemPrompt})
	allMessages = append(allMessages, messages...)

	body := openaiChatRequest{
		Model:    c.model,
		Messages: allMessages,
		Stream:   true,
		Tools:    wrapOpenAITools(tools),
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("llm.openaiClient.StreamWithTools: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("llm.openaiClient.StreamWithTools: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("llm.openaiClient.StreamWithTools: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm.openaiClient.StreamWithTools: status %d: %s", resp.StatusCode, string(respBody))
	}

	// Accumulate tool calls by index across SSE chunks. Each index
	// gets a stable ID and function name from the first chunk that
	// introduces it; subsequent chunks append argument fragments.
	type accTool struct {
		ID        string
		Name      string
		Arguments string
	}
	acc := make(map[int]*accTool)
	var accOrder []int
	var hasToolCalls bool
	var tp thinkParser

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk openaiStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]

		if choice.Delta.Reasoning != "" {
			if err := onEvent(StreamEvent{Type: StreamEventReasoning, Token: choice.Delta.Reasoning}); err != nil {
				return err
			}
		}

		if choice.Delta.Content != "" {
			if c.isOllama {
				if err := tp.feed(choice.Delta.Content,
					func(s string) error { return onEvent(StreamEvent{Type: StreamEventText, Token: s}) },
					func(s string) error { return onEvent(StreamEvent{Type: StreamEventReasoning, Token: s}) },
				); err != nil {
					return err
				}
			} else {
				if err := onEvent(StreamEvent{Type: StreamEventText, Token: choice.Delta.Content}); err != nil {
					return err
				}
			}
		}

		for _, tc := range choice.Delta.ToolCalls {
			hasToolCalls = true
			entry, ok := acc[tc.Index]
			if !ok {
				entry = &accTool{}
				acc[tc.Index] = entry
				accOrder = append(accOrder, tc.Index)
			}
			if tc.ID != "" {
				entry.ID = tc.ID
			}
			if tc.Function.Name != "" {
				entry.Name = tc.Function.Name
			}
			if tc.Function.Arguments != "" {
				entry.Arguments += tc.Function.Arguments
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("llm.openaiClient.StreamWithTools: scan: %w", err)
	}

	// Flush any residual text buffered in the think parser. The callbacks
	// MUST forward to onEvent — using no-ops here drops the last ~6 chars
	// (the thinkParser safety buffer for split  tags) of every
	// Ollama response.
	if err := tp.flush(
		func(s string) error { return onEvent(StreamEvent{Type: StreamEventText, Token: s}) },
		func(s string) error { return onEvent(StreamEvent{Type: StreamEventReasoning, Token: s}) },
	); err != nil {
		return err
	}

	if hasToolCalls {
		toolCalls := make([]ToolCall, 0, len(accOrder))
		for _, idx := range accOrder {
			entry := acc[idx]
			toolCalls = append(toolCalls, ToolCall{
				ID:        entry.ID,
				Name:      entry.Name,
				Arguments: entry.Arguments,
			})
		}
		return onEvent(StreamEvent{Type: StreamEventToolCalls, ToolCalls: toolCalls})
	}
	return nil
}

func (c *openaiClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	body := openaiChatRequest{
		Model:    c.model,
		Messages: messages,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("llm.openaiClient.Chat: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("llm.openaiClient.Chat: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm.openaiClient.Chat: do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm.openaiClient.Chat: read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llm.openaiClient.Chat: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openaiChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("llm.openaiClient.Chat: unmarshal: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("llm.openaiClient.Chat: api error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm.openaiClient.Chat: no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}

type anthropicClient struct {
	model  string
	apiKey string
	http   *http.Client
}

func (c *anthropicClient) Stream(ctx context.Context, systemPrompt string, messages []Message, onToken func(string) error) error {
	return fmt.Errorf("llm.anthropicClient.Stream: not implemented")
}

func (c *anthropicClient) StreamWithTools(ctx context.Context, systemPrompt string, messages []Message, tools []ToolDefinition, onEvent func(StreamEvent) error) error {
	return fmt.Errorf("llm.anthropicClient.StreamWithTools: not implemented")
}

type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []Message `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *anthropicClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body := anthropicRequest{
		Model:     c.model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: userPrompt},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("llm.anthropicClient.Chat: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("llm.anthropicClient.Chat: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm.anthropicClient.Chat: do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm.anthropicClient.Chat: read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llm.anthropicClient.Chat: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("llm.anthropicClient.Chat: unmarshal: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("llm.anthropicClient.Chat: api error: %s", result.Error.Message)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("llm.anthropicClient.Chat: no content returned")
	}

	return result.Content[0].Text, nil
}
