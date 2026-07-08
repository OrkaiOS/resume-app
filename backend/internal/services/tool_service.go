package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/marco/resume-app/internal/llm"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/orkai"
)

// ShellResult is the output of a single shell command execution.
type ShellResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

// ShellService executes shell commands in an isolated temporary
// directory. Each call gets its own temp dir to satisfy NFR-19
// (sandboxed shell execution). The default timeout is 30 seconds.
type ShellService struct {
	defaultTimeout time.Duration
}

// NewShellService creates a ShellService with a 30s default timeout.
func NewShellService() *ShellService {
	return &ShellService{defaultTimeout: 30 * time.Second}
}

// Execute runs the given script with the specified language interpreter.
// language must be "bash" or "python"; anything else falls back to bash.
// The script is written to a temp file and executed; the temp dir is
// always cleaned up.
func (s *ShellService) Execute(ctx context.Context, command, language string) (ShellResult, error) {
	if command == "" {
		return ShellResult{}, fmt.Errorf("services.ShellService.Execute: empty command")
	}

	ctx, cancel := context.WithTimeout(ctx, s.defaultTimeout)
	defer cancel()

	dir, err := os.MkdirTemp("", "resume-app-shell-*")
	if err != nil {
		return ShellResult{}, fmt.Errorf("services.ShellService.Execute: mkdir: %w", err)
	}
	defer os.RemoveAll(dir)

	var scriptPath, interpreter string
	switch strings.ToLower(language) {
	case "python", "python3":
		scriptPath = filepath.Join(dir, "script.py")
		interpreter = "python3"
	default:
		scriptPath = filepath.Join(dir, "script.sh")
		interpreter = "bash"
	}

	if err := os.WriteFile(scriptPath, []byte(command), 0o644); err != nil {
		return ShellResult{}, fmt.Errorf("services.ShellService.Execute: write script: %w", err)
	}

	cmd := exec.CommandContext(ctx, interpreter, scriptPath)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			return ShellResult{
				Stdout:   stdout.String(),
				Stderr:   fmt.Sprintf("execution timed out after %s: %s", s.defaultTimeout, stderr.String()),
				ExitCode: -1,
			}, nil
		} else {
			return ShellResult{}, fmt.Errorf("services.ShellService.Execute: run: %w", err)
		}
	}

	return ShellResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// OrkaiSearchService searches the orkai workspace for documents.
type OrkaiSearchService struct {
	client *orkai.OrkaiClient
}

// NewOrkaiSearchService creates an OrkaiSearchService.
func NewOrkaiSearchService(client *orkai.OrkaiClient) *OrkaiSearchService {
	return &OrkaiSearchService{client: client}
}

// Search calls the orkai MCP search_document tool and returns the
// formatted results string.
func (s *OrkaiSearchService) Search(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("services.OrkaiSearchService.Search: empty query")
	}
	return s.client.SearchDocuments(ctx, query)
}

// ToolRegistry implements llm.ToolRegistry by wiring the shell, orkai
// search, profile, and artifact tools. Each Execute returns a JSON
// string the LLM sees as the tool result.
// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision One registry owns all agent tools (shell, orkai_search, profile, artifacts). Per-call temp dirs for shell (per-session workspace is FR-034 scope). Errors from tools are returned as JSON {error:...} so the agent can react, not as Go errors that would abort the chat.
type ToolRegistry struct {
	shell     *ShellService
	search    *OrkaiSearchService
	profile   *ProfileService
	artifacts *ArtifactService
	defs      []llm.ToolDefinition
}

// NewToolRegistry builds a ToolRegistry wiring all four tool services.
func NewToolRegistry(shell *ShellService, orkaiClient *orkai.OrkaiClient, profile *ProfileService, artifacts *ArtifactService) *ToolRegistry {
	return &ToolRegistry{
		shell:     shell,
		search:    NewOrkaiSearchService(orkaiClient),
		profile:   profile,
		artifacts: artifacts,
		defs: []llm.ToolDefinition{
			{
				Name:        "shell",
				Description: "Execute a bash or python script in a sandboxed temporary directory. Use this to run shell commands, generate files, or test scripts. The working directory is isolated per call.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "command": {"type": "string", "description": "The full script content to execute"},
    "language": {"type": "string", "enum": ["bash", "python"], "description": "Interpreter to use (default bash)"}
  },
  "required": ["command"]
}`),
			},
			{
				Name:        "orkai_search",
				Description: "Search the user's orkai workspace for standards, skills, and documents. Use this when the system prompt is insufficient and you need additional context about cover letter principles, PDF generation, or past accepted documents.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Natural language search query"}
  },
  "required": ["query"]
}`),
			},
			{
				Name:        "get_profile",
				Description: "Return the user's structured professional profile (name, contact, work history, skills, etc.). Read-only.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
			{
				Name:        "list_artifacts",
				Description: "List reusable scripts saved as artifacts. Check this before creating a new script for the same purpose.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
			{
				Name:        "save_artifact",
				Description: "Save a script as a reusable artifact with a description.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "name": {"type": "string", "description": "Filename or descriptive label"},
    "type": {"type": "string", "enum": ["python", "bash"], "description": "Script type"},
    "description": {"type": "string", "description": "What the script does"},
    "scriptContent": {"type": "string", "description": "The script source code"}
  },
  "required": ["name", "type"]
}`),
			},
			{
				Name:        "get_artifact",
				Description: "Retrieve the content of a specific artifact by ID.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "id": {"type": "string", "description": "The artifact ID"}
  },
  "required": ["id"]
}`),
			},
		},
	}
}

// Definitions returns the tool definitions advertised to the LLM.
func (r *ToolRegistry) Definitions() []llm.ToolDefinition {
	return r.defs
}

// Execute dispatches a tool call to the matching service and returns the
// JSON-encoded result string the LLM sees. Tool execution errors are
// returned as JSON {"error": "..."} strings, not Go errors, so the
// agent can react to them without aborting the chat.
func (r *ToolRegistry) Execute(ctx context.Context, call llm.ToolCall) (string, error) {
	switch call.Name {
	case "shell":
		return r.execShell(ctx, call.Arguments)
	case "orkai_search":
		return r.execSearch(ctx, call.Arguments)
	case "get_profile":
		return r.execProfile(ctx)
	case "list_artifacts":
		return r.execListArtifacts(ctx)
	case "save_artifact":
		return r.execSaveArtifact(ctx, call.Arguments)
	case "get_artifact":
		return r.execGetArtifact(ctx, call.Arguments)
	default:
		return encodeJSON(map[string]string{"error": "unknown tool: " + call.Name})
	}
}

func (r *ToolRegistry) execShell(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Command  string `json:"command"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execShell: parse args: %w", err)
	}
	result, err := r.shell.Execute(ctx, args.Command, args.Language)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(result)
}

func (r *ToolRegistry) execSearch(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execSearch: parse args: %w", err)
	}
	result, err := r.search.Search(ctx, args.Query)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(map[string]string{"results": result})
}

func (r *ToolRegistry) execProfile(ctx context.Context) (string, error) {
	profile, err := r.profile.Get(ctx)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(profile)
}

func (r *ToolRegistry) execListArtifacts(ctx context.Context) (string, error) {
	items, err := r.artifacts.List(ctx)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(items)
}

func (r *ToolRegistry) execSaveArtifact(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Name          string `json:"name"`
		Type          string `json:"type"`
		Description   string `json:"description"`
		ScriptContent string `json:"scriptContent"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execSaveArtifact: parse args: %w", err)
	}
	created, err := r.artifacts.Create(ctx, models.Artifact{
		Name:          args.Name,
		Type:          args.Type,
		Description:   args.Description,
		ScriptContent: args.ScriptContent,
	})
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(created)
}

func (r *ToolRegistry) execGetArtifact(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execGetArtifact: parse args: %w", err)
	}
	item, err := r.artifacts.Get(ctx, args.ID)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(item)
}

func encodeJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("services.ToolRegistry.encodeJSON: %w", err)
	}
	return string(b), nil
}
