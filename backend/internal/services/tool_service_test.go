package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/marco/resume-app/internal/llm"
)

func TestShellServiceExecuteBash(t *testing.T) {
	t.Parallel()
	svc := NewShellService()
	result, err := svc.Execute(context.Background(), "echo hello", "bash")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Stdout == "" {
		t.Fatalf("expected stdout, got empty")
	}
	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("expected stdout to contain 'hello', got %q", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestShellServiceExecutePython(t *testing.T) {
	t.Parallel()
	svc := NewShellService()
	result, err := svc.Execute(context.Background(), "print('py-hello')", "python")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result.Stdout, "py-hello") {
		t.Errorf("expected stdout to contain 'py-hello', got %q", result.Stdout)
	}
}

func TestShellServiceEmptyCommand(t *testing.T) {
	t.Parallel()
	svc := NewShellService()
	_, err := svc.Execute(context.Background(), "", "bash")
	if err == nil {
		t.Fatal("expected error on empty command, got nil")
	}
}

func TestShellServiceNonZeroExitCode(t *testing.T) {
	t.Parallel()
	svc := NewShellService()
	result, err := svc.Execute(context.Background(), "exit 7", "bash")
	if err != nil {
		t.Fatalf("expected no Go error on non-zero exit, got %v", err)
	}
	if result.ExitCode != 7 {
		t.Errorf("expected exit code 7, got %d", result.ExitCode)
	}
}

func TestToolRegistryExecuteShell(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)

	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "1",
		Name:      "shell",
		Arguments: `{"command":"echo tool","language":"bash"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var result ShellResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON shell result, parse error: %v", err)
	}
	if !strings.Contains(result.Stdout, "tool") {
		t.Errorf("expected stdout to contain 'tool', got %q", result.Stdout)
	}
}

func TestToolRegistryUnknownTool(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "2",
		Name:      "unknown_tool",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("expected no error on unknown tool, got %v", err)
	}
	if !strings.Contains(out, "unknown tool") {
		t.Errorf("expected 'unknown tool' in output, got %q", out)
	}
}

func TestToolRegistryDefinitionsCount(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	defs := registry.Definitions()
	if len(defs) != 14 {
		t.Errorf("expected 14 tool definitions, got %d", len(defs))
	}
}

func TestToolRegistryExecOverviewNoOrkai(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "3",
		Name:      "overview",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("expected error in output when orkai not configured, got %q", out)
	}
}

func TestToolRegistryExecSaveSessionNoService(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "4",
		Name:      "save_session",
		Arguments: `{"summary":"test summary"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "session service not configured") {
		t.Errorf("expected 'session service not configured' error, got %q", out)
	}
}

func TestToolRegistryExecUpdateSessionNoService(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "5",
		Name:      "update_session",
		Arguments: `{"sessionId":"abc","summary":"updated"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "session service not configured") {
		t.Errorf("expected 'session service not configured' error, got %q", out)
	}
}

func TestToolRegistryExecSaveUserInsightNoService(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "6",
		Name:      "save_user_insight",
		Arguments: `{"insight":"test insight"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "session service not configured") {
		t.Errorf("expected 'session service not configured' error, got %q", out)
	}
}

func TestToolRegistryExecSaveSessionMissingSummary(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "7",
		Name:      "save_session",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("expected error in output, got %q", out)
	}
}

func TestToolRegistryGeneratePdfNoService(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "8",
		Name:      "generate_pdf",
		Arguments: `{"markdown":"# Test","documentType":"resume","opportunityId":"1"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "PDF service not configured") {
		t.Errorf("expected 'PDF service not configured' error, got %q", out)
	}
}

func TestToolRegistryGeneratePdfValidation(t *testing.T) {
	t.Parallel()
	pdfSvc := NewPDFService(t.TempDir())
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, pdfSvc, nil, nil, nil)

	tests := []struct {
		name    string
		args    string
		wantErr string
	}{
		{"missing opportunity service", `{"markdown":"# Test","documentType":"resume","opportunityId":"1"}`, "opportunity service not configured"},
		{"invalid document type", `{"markdown":"# Test","documentType":"invalid","opportunityId":"1"}`, ""},
		{"missing markdown", `{"markdown":"","documentType":"resume","opportunityId":"1"}`, ""},
		{"missing opportunityId", `{"markdown":"# Test","documentType":"resume","opportunityId":""}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := registry.Execute(context.Background(), llm.ToolCall{
				ID:        "9",
				Name:      "generate_pdf",
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" && !strings.Contains(out, tt.wantErr) {
				t.Errorf("expected output to contain %q, got %q", tt.wantErr, out)
			}
			if tt.wantErr == "" && !strings.Contains(out, "error") {
				t.Errorf("expected error in output, got %q", out)
			}
		})
	}
}

func TestToolRegistryGeneratePdfToolInDefinitions(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	defs := registry.Definitions()
	found := false
	for _, d := range defs {
		if d.Name == "generate_pdf" {
			found = true
			if !strings.Contains(d.Description, "pandoc") || !strings.Contains(d.Description, "WeasyPrint") {
				t.Errorf("generate_pdf description should mention pandoc and WeasyPrint, got %q", d.Description)
			}
			break
		}
	}
	if !found {
		t.Error("generate_pdf tool not found in definitions")
	}
}

func TestToolRegistryListDocumentsNoService(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "10",
		Name:      "list_documents",
		Arguments: `{"opportunityId":"1"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "document services not configured") {
		t.Errorf("expected 'document services not configured', got %q", out)
	}
}

func TestToolRegistryDeletePdfNoService(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "11",
		Name:      "delete_pdf",
		Arguments: `{"opportunityId":"1","documentType":"resume"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "document services not configured") {
		t.Errorf("expected 'document services not configured', got %q", out)
	}
}

func TestToolRegistryDeletePdfMissingArgs(t *testing.T) {
	t.Parallel()
	pdfSvc := NewPDFService(t.TempDir())
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)

	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "12",
		Name:      "delete_pdf",
		Arguments: `{"documentType":"resume"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "opportunityId is required") {
		t.Errorf("expected 'opportunityId is required', got %q", out)
	}

	_, _ = pdfSvc, registry
}

func TestToolRegistryDeletePdfInvalidType(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "13",
		Name:      "delete_pdf",
		Arguments: `{"opportunityId":"1","documentType":"invalid"}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "must be 'resume'") {
		t.Errorf("expected validation error, got %q", out)
	}
}

func TestToolRegistryListDocumentsMissingArgs(t *testing.T) {
	t.Parallel()
	registry := NewToolRegistry(NewShellService(), nil, nil, nil, nil, nil, nil, nil, nil, nil)
	out, err := registry.Execute(context.Background(), llm.ToolCall{
		ID:        "14",
		Name:      "list_documents",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(out, "opportunityId is required") {
		t.Errorf("expected 'opportunityId is required', got %q", out)
	}
}
