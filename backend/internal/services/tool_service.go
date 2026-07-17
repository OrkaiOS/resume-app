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
	"github.com/marco/resume-app/internal/store"
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
	client          *orkai.OrkaiClient
	onboardingStore store.OnboardingStore
}

// NewOrkaiSearchService creates an OrkaiSearchService.
func NewOrkaiSearchService(client *orkai.OrkaiClient, onboardingStore store.OnboardingStore) *OrkaiSearchService {
	return &OrkaiSearchService{client: client, onboardingStore: onboardingStore}
}

// Search calls the orkai MCP search_document tool and returns the
// formatted results string.
func (s *OrkaiSearchService) Search(ctx context.Context, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("services.OrkaiSearchService.Search: empty query")
	}
	state, err := s.onboardingStore.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("services.OrkaiSearchService.Search: getting category ID: %w", err)
	}
	if state.OrkaiCategoryID == "" {
		return "", fmt.Errorf("services.OrkaiSearchService.Search: orkai category not configured")
	}
	return s.client.SearchDocuments(ctx, query, state.OrkaiCategoryID)
}

// ToolRegistry implements llm.ToolRegistry by wiring the shell, orkai
// search, profile, artifact, session, and user-insight tools. Each
// Execute returns a JSON string the LLM sees as the tool result.
// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision One registry owns all agent tools (shell, orkai_search, orkai_get, profile, artifacts, overview, save_session, update_session, save_user_insight). Per-call temp dirs for shell (per-session workspace is FR-034 scope). Errors from tools are returned as JSON {error:...} so the agent can react, not as Go errors that would abort the chat.
type ToolRegistry struct {
	shell       *ShellService
	search      *OrkaiSearchService
	orkai       *orkai.OrkaiClient
	profile     *ProfileService
	artifacts   *ArtifactService
	session     *SessionService
	pdf         *PDFService
	opportunity *OpportunityService
	resume      *ResumeService
	coverLetter *CoverLetterService
	defs        []llm.ToolDefinition
}

// NewToolRegistry builds a ToolRegistry wiring all tool services.
func NewToolRegistry(shell *ShellService, orkaiClient *orkai.OrkaiClient, onboardingStore store.OnboardingStore, profile *ProfileService, artifacts *ArtifactService, session *SessionService, pdf *PDFService, opportunity *OpportunityService, resume *ResumeService, coverLetter *CoverLetterService) *ToolRegistry {
	return &ToolRegistry{
		shell:       shell,
		search:      NewOrkaiSearchService(orkaiClient, onboardingStore),
		orkai:       orkaiClient,
		profile:     profile,
		artifacts:   artifacts,
		session:     session,
		pdf:         pdf,
		opportunity: opportunity,
		resume:      resume,
		coverLetter: coverLetter,
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
				Description: "Search the user's orkai workspace for standards, skills, and documents by semantic meaning. Use this when you do NOT already know the entity ID.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Natural language search query"}
  },
  "required": ["query"]
}`),
			},
			{
				Name:        "orkai_get",
				Description: "Fetch a specific orkai entity (standard, skill, or document) by its ID. Use this when the entity ID is known or referenced in the system prompt — it is faster and returns the exact entity. Prefer over orkai_search when you have an ID.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "id": {"type": "string", "description": "The orkai entity ID (UUID)"}
  },
  "required": ["id"]
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
			{
				Name:        "overview",
				Description: "Get a summary of the orkai workspace: recent session summaries, available standards, and available skills. Use this at the start of every chat session to discover what sessions exist for this opportunity and what standards/skills you can draw on. This is your discovery mechanism — call it before searching blindly.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
			},
			{
				Name:        "save_session",
				Description: "Save a distilled summary of the current conversation to orkai. Use this at meaningful checkpoints: after resolving a user concern, after producing a draft, after a revision, or after the user shares durable context. The summary should capture what was discussed, decided, and produced — NOT a raw transcript. You will receive a session ID back; use it with update_session for subsequent saves in the same conversation arc.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "summary": {"type": "string", "description": "Distilled summary of the conversation — what was discussed, decided, and produced. NOT a raw transcript."},
    "opportunityId": {"type": "string", "description": "The opportunity ID this session belongs to"},
    "company": {"type": "string", "description": "Company name for the session name"},
    "role": {"type": "string", "description": "Role title for the session name"}
  },
  "required": ["summary"]
}`),
			},
			{
				Name:        "update_session",
				Description: "Update an existing orkai session with new summary text. Use this at subsequent checkpoints in the same conversation arc (after the first save_session). You must provide the session ID returned by a prior save_session call.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "sessionId": {"type": "string", "description": "The session ID returned by a prior save_session call"},
    "summary": {"type": "string", "description": "Updated distilled summary of the conversation"}
  },
  "required": ["sessionId", "summary"]
}`),
			},
			{
				Name:        "save_user_insight",
				Description: "Save a durable user-specific insight that should influence all future resume and cover letter generation. Use this when the user shares something with lasting relevance: tone/style preferences, career narrative, constraints (e.g. age-bias mitigation framing), accessibility needs, naming preferences. The insight is stored in a single 'User Insights' standard that is loaded into the system prompt for every future session — the user never has to repeat it.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "insight": {"type": "string", "description": "The durable user insight to preserve for future sessions. Write it as a clear, actionable guideline the agent can follow."}
  },
  "required": ["insight"]
}`),
			},
			{
				Name:        "generate_pdf",
				Description: "Generate a PDF from markdown content and persist it. Use this when the user approves a resume or cover letter draft. The PDF is rendered via pandoc (Markdown→HTML) + WeasyPrint (HTML+CSS→PDF) using a professional CSS template, saved to disk, and the document record is updated. Returns a download URL you should include in your response.",
				Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "markdown": {"type": "string", "description": "The full markdown content of the document"},
    "documentType": {"type": "string", "enum": ["resume", "cover_letter"], "description": "Document type"},
    "opportunityId": {"type": "string", "description": "The opportunity ID this document belongs to"}
  },
  "required": ["markdown", "documentType", "opportunityId"]
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
	case "orkai_get":
		return r.execGetEntity(ctx, call.Arguments)
	case "get_profile":
		return r.execProfile(ctx)
	case "list_artifacts":
		return r.execListArtifacts(ctx)
	case "save_artifact":
		return r.execSaveArtifact(ctx, call.Arguments)
	case "get_artifact":
		return r.execGetArtifact(ctx, call.Arguments)
	case "overview":
		return r.execOverview(ctx)
	case "save_session":
		return r.execSaveSession(ctx, call.Arguments)
	case "update_session":
		return r.execUpdateSession(ctx, call.Arguments)
	case "save_user_insight":
		return r.execSaveUserInsight(ctx, call.Arguments)
	case "generate_pdf":
		return r.execGeneratePdf(ctx, call.Arguments)
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

func (r *ToolRegistry) execGetEntity(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execGetEntity: parse args: %w", err)
	}
	if args.ID == "" {
		return encodeJSON(map[string]string{"error": "id is required"})
	}
	if r.orkai == nil {
		return encodeJSON(map[string]string{"error": "orkai client not configured"})
	}
	entity, err := r.orkai.GetEntity(ctx, args.ID)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(entity)
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

func (r *ToolRegistry) execOverview(ctx context.Context) (string, error) {
	if r.search == nil || r.orkai == nil {
		return encodeJSON(map[string]string{"error": "orkai client not configured"})
	}
	state, err := r.search.onboardingStore.Get(ctx)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	if state.OrkaiCategoryID == "" {
		return encodeJSON(map[string]string{"error": "orkai category not configured"})
	}
	overview, err := r.orkai.Overview(ctx, state.OrkaiCategoryID, "resume-app")
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(map[string]string{"overview": overview})
}

func (r *ToolRegistry) execSaveSession(ctx context.Context, argsJSON string) (string, error) {
	if r.session == nil {
		return encodeJSON(map[string]string{"error": "session service not configured"})
	}
	var args struct {
		Summary       string `json:"summary"`
		OpportunityID string `json:"opportunityId"`
		Company       string `json:"company"`
		Role          string `json:"role"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execSaveSession: parse args: %w", err)
	}
	if args.Summary == "" {
		return encodeJSON(map[string]string{"error": "summary is required"})
	}
	sessionID, err := r.session.Save(ctx, args.OpportunityID, args.Company, args.Role, args.Summary)
	if err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(map[string]string{"sessionId": sessionID, "message": "Session saved to orkai"})
}

func (r *ToolRegistry) execUpdateSession(ctx context.Context, argsJSON string) (string, error) {
	if r.session == nil {
		return encodeJSON(map[string]string{"error": "session service not configured"})
	}
	var args struct {
		SessionID string `json:"sessionId"`
		Summary   string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execUpdateSession: parse args: %w", err)
	}
	if args.SessionID == "" || args.Summary == "" {
		return encodeJSON(map[string]string{"error": "sessionId and summary are required"})
	}
	if err := r.session.Update(ctx, args.SessionID, args.Summary); err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(map[string]string{"message": "Session updated in orkai"})
}

func (r *ToolRegistry) execSaveUserInsight(ctx context.Context, argsJSON string) (string, error) {
	if r.session == nil {
		return encodeJSON(map[string]string{"error": "session service not configured"})
	}
	var args struct {
		Insight string `json:"insight"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execSaveUserInsight: parse args: %w", err)
	}
	if args.Insight == "" {
		return encodeJSON(map[string]string{"error": "insight is required"})
	}
	if err := r.session.SaveUserInsight(ctx, args.Insight); err != nil {
		return encodeJSON(map[string]string{"error": err.Error()})
	}
	return encodeJSON(map[string]string{"message": "Saved as a user insight for future sessions"})
}

func (r *ToolRegistry) execGeneratePdf(ctx context.Context, argsJSON string) (string, error) {
	if r.pdf == nil {
		return encodeJSON(map[string]string{"error": "PDF service not configured"})
	}
	if r.opportunity == nil {
		return encodeJSON(map[string]string{"error": "opportunity service not configured"})
	}
	if r.resume == nil || r.coverLetter == nil {
		return encodeJSON(map[string]string{"error": "document service not configured"})
	}

	var args struct {
		Markdown      string `json:"markdown"`
		DocumentType  string `json:"documentType"`
		OpportunityID string `json:"opportunityId"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("services.ToolRegistry.execGeneratePdf: parse args: %w", err)
	}
	if args.Markdown == "" {
		return encodeJSON(map[string]string{"error": "markdown is required"})
	}
	if args.DocumentType != "resume" && args.DocumentType != "cover_letter" {
		return encodeJSON(map[string]string{"error": "documentType must be 'resume' or 'cover_letter'"})
	}
	if args.OpportunityID == "" {
		return encodeJSON(map[string]string{"error": "opportunityId is required"})
	}

	opp, err := r.opportunity.Get(ctx, args.OpportunityID)
	if err != nil {
		return encodeJSON(map[string]string{"error": fmt.Sprintf("opportunity not found: %v", err)})
	}

	var css string
	docTypeLabel := "resume"
	pdfResource := "resume"
	if args.DocumentType == "cover_letter" {
		css = coverLetterCSS
		docTypeLabel = "cover-letter"
		pdfResource = "cover-letter"
	} else {
		css = resumeCSS
	}

	result, err := r.pdf.Generate(ctx, args.Markdown, css, docTypeLabel, opp.Company, opp.Role)
	if err != nil {
		return encodeJSON(map[string]string{"error": fmt.Sprintf("PDF generation failed: %v", err)})
	}

	if args.DocumentType == "cover_letter" {
		_, err = r.coverLetter.Upsert(ctx, models.CoverLetter{
			OpportunityID:   args.OpportunityID,
			MarkdownContent: args.Markdown,
			PDFPath:         result.Path,
			Status:          "approved",
		})
	} else {
		_, err = r.resume.Upsert(ctx, models.Resume{
			OpportunityID:   args.OpportunityID,
			MarkdownContent: args.Markdown,
			PDFPath:         result.Path,
			Status:          "approved",
		})
	}
	if err != nil {
		return encodeJSON(map[string]string{"error": fmt.Sprintf("failed to save document record: %v", err)})
	}

	downloadURL := fmt.Sprintf("/v1/api/opportunities/%s/%s/pdf", args.OpportunityID, pdfResource)
	return encodeJSON(map[string]string{
		"downloadUrl": downloadURL,
		"filename":    result.Filename,
		"message":     "PDF generated successfully",
	})
}

func encodeJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("services.ToolRegistry.encodeJSON: %w", err)
	}
	return string(b), nil
}
