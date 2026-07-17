package orkai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const defaultRequestTimeout = 10 * time.Second

type rpcReply struct {
	result json.RawMessage
	err    error
}

type OrkaiClient struct {
	sseURL    string
	token     string
	sseClient *http.Client
	rpcClient *http.Client
	requestID atomic.Int32
	msgURL    string
	pending   map[int32]chan rpcReply
	done      chan struct{}
	mu        sync.Mutex
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int32       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int32           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpToolResult struct {
	Content []mcpContentItem `json:"content"`
}

type createResponse struct {
	ID string `json:"id"`
}

func NewOrkaiClient(sseURL, token string) *OrkaiClient {
	return &OrkaiClient{
		sseURL:    sseURL,
		token:     token,
		sseClient: &http.Client{Timeout: 0},
		rpcClient: &http.Client{Timeout: defaultRequestTimeout},
		pending:   make(map[int32]chan rpcReply),
		done:      make(chan struct{}),
	}
}

func (c *OrkaiClient) connect() error {
	c.mu.Lock()
	if c.msgURL != "" {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	req, err := http.NewRequest(http.MethodGet, c.sseURL, nil)
	if err != nil {
		return fmt.Errorf("orkai.connect: new request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.sseClient.Do(req)
	if err != nil {
		return fmt.Errorf("orkai.connect: GET %s: %w", c.sseURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("orkai.connect: GET %s returned %d", c.sseURL, resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	msgURL, err := readSSEEndpoint(reader)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("orkai.connect: %w", err)
	}

	c.mu.Lock()
	c.msgURL = msgURL
	c.mu.Unlock()

	go c.readSSE(resp.Body, reader)
	return nil
}

func readSSEEndpoint(reader *bufio.Reader) (string, error) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read sse: %w", err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
				return data, nil
			}
		}
	}
}

func (c *OrkaiClient) readSSE(body io.ReadCloser, reader *bufio.Reader) {
	defer body.Close()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("orkai.readSSE: connection closed: %v", err)
			return
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			c.dispatchResponse(data)
		}
	}
}

func (c *OrkaiClient) dispatchResponse(data string) {
	var rpcResp jsonRPCResponse
	if err := json.Unmarshal([]byte(data), &rpcResp); err != nil {
		return
	}

	c.mu.Lock()
	ch, ok := c.pending[rpcResp.ID]
	c.mu.Unlock()

	if !ok {
		return
	}

	if rpcResp.Error != nil {
		select {
		case ch <- rpcReply{err: fmt.Errorf("orkai rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)}:
		default:
		}
		return
	}

	select {
	case ch <- rpcReply{result: rpcResp.Result}:
	default:
	}
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision ListCategoryNames returns all orkai category names via categories list. Used for uniqueness validation before creating a project workspace.
func (c *OrkaiClient) ListCategoryNames(ctx context.Context) ([]string, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	result, err := c.callTool(ctx, "categories", map[string]interface{}{
		"action": "list",
	})
	if err != nil {
		return nil, fmt.Errorf("orkai.ListCategoryNames: %w", err)
	}

	var items []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(result, &items); err != nil {
		var raw string
		if err2 := json.Unmarshal(result, &raw); err2 == nil {
			if err3 := json.Unmarshal([]byte(raw), &items); err3 != nil {
				return nil, fmt.Errorf("orkai.ListCategoryNames: parse response: %w", err3)
			}
		} else {
			return nil, fmt.Errorf("orkai.ListCategoryNames: parse response: %w", err)
		}
	}

	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names, nil
}

func (c *OrkaiClient) CreateCategory(ctx context.Context, name string) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}

	result, err := c.callTool(ctx, "categories", map[string]interface{}{
		"action":      "create",
		"name":        name,
		"description": "resume-app workspace",
	})
	if err != nil {
		return "", fmt.Errorf("orkai.CreateCategory: %w", err)
	}

	var r createResponse
	if err := json.Unmarshal(result, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateCategory: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateCategory: response missing id")
	}

	return r.ID, nil
}

func (c *OrkaiClient) CreateStandard(ctx context.Context, name, text, categoryID string) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}

	result, err := c.callTool(ctx, "standards", map[string]interface{}{
		"action":       "create",
		"name":         name,
		"text":         text,
		"category_ids": []string{categoryID},
	})
	if err != nil {
		return "", fmt.Errorf("orkai.CreateStandard: %w", err)
	}

	var r createResponse
	if err := json.Unmarshal(result, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateStandard: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateStandard: response missing id")
	}

	return r.ID, nil
}

func (c *OrkaiClient) CreateSkill(ctx context.Context, name, text, categoryID string) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}

	result, err := c.callTool(ctx, "skills", map[string]interface{}{
		"action":       "create",
		"name":         name,
		"text":         text,
		"category_ids": []string{categoryID},
	})
	if err != nil {
		return "", fmt.Errorf("orkai.CreateSkill: %w", err)
	}

	var r createResponse
	if err := json.Unmarshal(result, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateSkill: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateSkill: response missing id")
	}

	return r.ID, nil
}

type EntityResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Text string `json:"text"`
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision GetEntity fetches an orkai entity by ID at runtime using the MCP entity tool. Used by SystemPromptService to read profile standards and skills for chat session context.
func (c *OrkaiClient) GetEntity(ctx context.Context, id string) (EntityResponse, error) {
	if err := c.connect(); err != nil {
		return EntityResponse{}, err
	}

	result, err := c.callTool(ctx, "entity", map[string]interface{}{
		"action": "get",
		"id":     id,
	})
	if err != nil {
		return EntityResponse{}, fmt.Errorf("orkai.GetEntity: %w", err)
	}

	var r EntityResponse
	if err := json.Unmarshal(result, &r); err != nil {
		return EntityResponse{}, fmt.Errorf("orkai.GetEntity: parse response: %w", err)
	}
	if r.ID == "" {
		return EntityResponse{}, fmt.Errorf("orkai.GetEntity: response missing id")
	}

	return r, nil
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision SearchDocuments calls the orkai MCP search_document tool and returns the formatted results string. Used by the agent's orkai_search tool to discover additional context beyond the system prompt.
func (c *OrkaiClient) SearchDocuments(ctx context.Context, query, categoryID string) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}

	result, err := c.callTool(ctx, "search_document", map[string]interface{}{
		"query":        query,
		"category_ids": []string{categoryID},
	})
	if err != nil {
		return "", fmt.Errorf("orkai.SearchDocuments: %w", err)
	}

	// search_document returns a text content item with the results.
	// Unwrap the MCP content envelope if present, otherwise return raw.
	var raw string
	if err := json.Unmarshal(result, &raw); err == nil {
		return raw, nil
	}
	// If it's not a plain string, return the raw JSON — the agent can
	// read either.
	return string(result), nil
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision Overview calls the orkai MCP overview tool scoped to a category. Returns a summary of recent sessions, available standards, and skills — the agent's discovery mechanism at session start (FR-034).
func (c *OrkaiClient) Overview(ctx context.Context, categoryID, projectName string) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}

	args := map[string]interface{}{}
	if categoryID != "" {
		args["category_id"] = categoryID
	}
	if projectName != "" {
		args["project_name"] = projectName
	}

	result, err := c.callTool(ctx, "overview", args)
	if err != nil {
		return "", fmt.Errorf("orkai.Overview: %w", err)
	}

	var raw string
	if err := json.Unmarshal(result, &raw); err == nil {
		return raw, nil
	}
	return string(result), nil
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision CreateSession creates an orkai session entity in the personal category with opportunity metadata. Returns the new session ID for subsequent update_session calls (FR-039).
func (c *OrkaiClient) CreateSession(ctx context.Context, name, text, categoryID string, metadata map[string]string) (string, error) {
	if err := c.connect(); err != nil {
		return "", err
	}

	args := map[string]interface{}{
		"action":       "create",
		"name":         name,
		"text":         text,
		"category_ids": []string{categoryID},
	}
	if metadata != nil {
		args["metadata"] = metadata
	}

	result, err := c.callTool(ctx, "session", args)
	if err != nil {
		return "", fmt.Errorf("orkai.CreateSession: %w", err)
	}

	var r createResponse
	if err := json.Unmarshal(result, &r); err != nil {
		return "", fmt.Errorf("orkai.CreateSession: parse response: %w", err)
	}
	if r.ID == "" {
		return "", fmt.Errorf("orkai.CreateSession: response missing id")
	}
	return r.ID, nil
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision UpdateSession updates an existing orkai session entity by ID. Used by the agent at subsequent checkpoints in the same conversation arc (FR-039).
func (c *OrkaiClient) UpdateSession(ctx context.Context, id, text string) error {
	if err := c.connect(); err != nil {
		return err
	}

	_, err := c.callTool(ctx, "session", map[string]interface{}{
		"action": "update",
		"id":     id,
		"text":   text,
	})
	if err != nil {
		return fmt.Errorf("orkai.UpdateSession: %w", err)
	}
	return nil
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision SearchStandards searches the orkai workspace for standards by semantic query. Used by save_user_insight to find the existing User Insights standard by name before updating it.
func (c *OrkaiClient) SearchStandards(ctx context.Context, categoryID, query string) ([]EntityResponse, error) {
	if err := c.connect(); err != nil {
		return nil, err
	}

	result, err := c.callTool(ctx, "standards", map[string]interface{}{
		"action":       "search",
		"query":        query,
		"category_ids": []string{categoryID},
	})
	if err != nil {
		return nil, fmt.Errorf("orkai.SearchStandards: %w", err)
	}

	// The standards search returns a JSON array of entity objects.
	var items []EntityResponse
	if err := json.Unmarshal(result, &items); err != nil {
		// Try unwrapping a string first (some MCP responses wrap in a string).
		var raw string
		if err2 := json.Unmarshal(result, &raw); err2 == nil {
			if err3 := json.Unmarshal([]byte(raw), &items); err3 != nil {
				return nil, fmt.Errorf("orkai.SearchStandards: parse response: %w", err3)
			}
		} else {
			return nil, fmt.Errorf("orkai.SearchStandards: parse response: %w", err)
		}
	}
	return items, nil
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision UpdateStandard updates an existing orkai standard entity by ID. Used by save_user_insight to append new insights to the existing User Insights standard (FR-032).
func (c *OrkaiClient) UpdateStandard(ctx context.Context, id, text string) error {
	if err := c.connect(); err != nil {
		return err
	}

	_, err := c.callTool(ctx, "standards", map[string]interface{}{
		"action": "update",
		"id":     id,
		"text":   text,
	})
	if err != nil {
		return fmt.Errorf("orkai.UpdateStandard: %w", err)
	}
	return nil
}

func (c *OrkaiClient) LinkEntities(ctx context.Context, sourceID, targetID string) error {
	if err := c.connect(); err != nil {
		return err
	}

	_, err := c.callTool(ctx, "entity", map[string]interface{}{
		"action": "update",
		"id":     sourceID,
		"relations": []map[string]string{
			{"type": "references", "targetId": targetID},
		},
	})
	if err != nil {
		return fmt.Errorf("orkai.LinkEntities: %w", err)
	}

	return nil
}

func (c *OrkaiClient) callTool(ctx context.Context, toolName string, args map[string]interface{}) (json.RawMessage, error) {
	resp, err := c.rpcCall(ctx, "tools/call", map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	})
	if err != nil {
		return nil, err
	}

	var result mcpToolResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("orkai.callTool: parse tool result: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("orkai.callTool: empty tool result")
	}

	return json.RawMessage(result.Content[0].Text), nil
}

func (c *OrkaiClient) rpcCall(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id := c.requestID.Add(1)

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("orkai.rpcCall: marshal: %w", err)
	}

	c.mu.Lock()
	msgURL := c.msgURL
	c.mu.Unlock()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, msgURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("orkai.rpcCall: new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	respCh := make(chan rpcReply, 1)
	c.mu.Lock()
	c.pending[id] = respCh
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	resp, err := c.rpcClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("orkai.rpcCall: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("orkai.rpcCall: %s returned %d: %s", method, resp.StatusCode, string(respBody))
	}

	select {
	case reply := <-respCh:
		if reply.err != nil {
			return nil, fmt.Errorf("orkai.rpcCall: %w", reply.err)
		}
		return reply.result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
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

		var cfg map[string]interface{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		if token := extractToken(cfg); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("orkai.DetectMCPToken: no ORKAI_MCP_TOKEN found in env or config files (%v)", paths)
}

func extractToken(cfg map[string]interface{}) string {
	if t, ok := cfg["ORKAI_MCP_TOKEN"].(string); ok && t != "" {
		return t
	}

	servers, _ := cfg["mcpServers"].(map[string]interface{})
	for _, srv := range servers {
		srvMap, _ := srv.(map[string]interface{})
		env, _ := srvMap["env"].(map[string]interface{})
		if env != nil {
			if t, ok := env["ORKAI_MCP_TOKEN"].(string); ok && t != "" {
				return t
			}
		}
		headers, _ := srvMap["headers"].(map[string]interface{})
		if headers != nil {
			if t, ok := headers["Authorization"].(string); ok && t != "" {
				return strings.TrimPrefix(t, "Bearer ")
			}
		}
	}

	return ""
}
