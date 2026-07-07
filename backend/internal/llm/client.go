package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

func NewClient(provider, model, apiKey string) Client {
	switch strings.ToLower(provider) {
	case "anthropic":
		return &anthropicClient{model: model, apiKey: apiKey, http: &http.Client{Timeout: 120 * time.Second}}
	default:
		baseURL := "https://api.openai.com/v1"
		if provider == "ollama" {
			baseURL = "http://localhost:11434/v1"
		}
		return &openaiClient{model: model, apiKey: apiKey, baseURL: baseURL, http: &http.Client{Timeout: 120 * time.Second}}
	}
}

type openaiClient struct {
	model   string
	apiKey  string
	baseURL string
	http    *http.Client
}

type openaiChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
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
