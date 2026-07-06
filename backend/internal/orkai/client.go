package orkai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultRequestTimeout = 30 * time.Second

type OrkaiClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewOrkaiClient(baseURL, token string) *OrkaiClient {
	return &OrkaiClient{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}
}

type createResponse struct {
	ID string `json:"id"`
}

func (c *OrkaiClient) CreateCategory(ctx context.Context, name string) (string, error) {
	body := map[string]any{
		"action":      "create",
		"name":        name,
		"description": "User's personal resume-app workspace",
	}

	resp, err := c.mcpCall(ctx, "categories", body)
	if err != nil {
		return "", err
	}

	var r createResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateCategory: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateCategory: response missing id")
	}

	return r.ID, nil
}

func (c *OrkaiClient) CreateStandard(ctx context.Context, name, text, categoryID string) (string, error) {
	body := map[string]any{
		"action":       "create",
		"name":         name,
		"text":         text,
		"category_ids": []string{categoryID},
	}

	resp, err := c.mcpCall(ctx, "standards", body)
	if err != nil {
		return "", err
	}

	var r createResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateStandard: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateStandard: response missing id")
	}

	return r.ID, nil
}

func (c *OrkaiClient) CreateSkill(ctx context.Context, name, text, categoryID string) (string, error) {
	body := map[string]any{
		"action":       "create",
		"name":         name,
		"text":         text,
		"category_ids": []string{categoryID},
	}

	resp, err := c.mcpCall(ctx, "skills", body)
	if err != nil {
		return "", err
	}

	var r createResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateSkill: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateSkill: response missing id")
	}

	return r.ID, nil
}

func (c *OrkaiClient) LinkEntities(ctx context.Context, sourceID, targetID string) error {
	body := map[string]any{
		"action": "update",
		"id":     sourceID,
		"relations": []map[string]string{
			{"type": "references", "targetId": targetID},
		},
	}

	_, err := c.mcpCall(ctx, "entity", body)
	if err != nil {
		return err
	}

	return nil
}

func (c *OrkaiClient) mcpCall(ctx context.Context, tool string, body map[string]any) ([]byte, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("orkai.mcpCall: marshal body: %w", err)
	}

	url := c.baseURL + "/" + tool
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("orkai.mcpCall: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("orkai.mcpCall: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("orkai.mcpCall: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("orkai.mcpCall: %s returned %d: %s", tool, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func DetectMCPToken() (string, error) {
	if t := os.Getenv("ORKAI_MCP_TOKEN"); t != "" {
		return t, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("orkai.DetectMCPToken: cannot determine home directory: %w", err)
	}

	paths := []string{
		filepath.Join(home, ".cursor", "mcp.json"),
		filepath.Join(home, ".config", "opencode", "mcp.json"),
		filepath.Join(home, ".config", "cline", "mcp.json"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		var cfg map[string]any
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		if token := extractToken(cfg); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("orkai.DetectMCPToken: no ORKAI_MCP_TOKEN found in env or config files (%v)", paths)
}

func extractToken(cfg map[string]any) string {
	if t, ok := cfg["ORKAI_MCP_TOKEN"].(string); ok && t != "" {
		return t
	}

	servers, _ := cfg["mcpServers"].(map[string]any)
	for _, srv := range servers {
		srvMap, _ := srv.(map[string]any)
		env, _ := srvMap["env"].(map[string]any)
		if env != nil {
			if t, ok := env["ORKAI_MCP_TOKEN"].(string); ok && t != "" {
				return t
			}
		}
		headers, _ := srvMap["headers"].(map[string]any)
		if headers != nil {
			if t, ok := headers["Authorization"].(string); ok && len(t) > 7 {
				return t[7:] // strip "Bearer " prefix
			}
		}
	}

	return ""
}
