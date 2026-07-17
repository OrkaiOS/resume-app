package services

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Acme Corp", "Acme-Corp"},
		{"with spaces", "  Leading   Spaces  ", "Leading-Spaces"},
		{"special chars", "Foo!@#Bar$%^Baz", "FooBarBaz"},
		{"empty becomes untitled", "", "Untitled"},
		{"only special", "!@#$%^", "Untitled"},
		{"hyphens preserved", "a-b c_d", "a-b-c_d"},
		{"numbers ok", "Company 123", "Company-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize(tt.input)
			if got != tt.want {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPDFService_OutputDirCreated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "pdfs")

	svc := NewPDFService(outDir)

	_, err := os.Stat(outDir)
	if err == nil {
		t.Fatal("output dir should not exist before first Generate call")
	}

	_ = svc
}

func TestPDFService_MissingPandoc(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("pandoc"); err == nil {
		t.Skip("pandoc is installed, skipping missing-binary test")
	}

	svc := NewPDFService(t.TempDir())
	_, err := svc.Generate(context.Background(), "# Test", "body{}", "resume", "Acme", "Engineer")
	if err == nil {
		t.Fatal("expected error for missing pandoc, got nil")
	}
	if !strings.Contains(err.Error(), "pandoc") {
		t.Errorf("expected error to mention pandoc, got: %v", err)
	}
}

func TestPDFService_MissingWeasyprint(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed, skipping weasyprint missing test")
	}
	if _, err := exec.LookPath("weasyprint"); err == nil {
		t.Skip("weasyprint is installed, skipping missing-binary test")
	}

	svc := NewPDFService(t.TempDir())
	_, err := svc.Generate(context.Background(), "# Test", "body{}", "resume", "Acme", "Engineer")
	if err == nil {
		t.Fatal("expected error for missing weasyprint, got nil")
	}
	if !strings.Contains(err.Error(), "weasyprint") {
		t.Errorf("expected error to mention weasyprint, got: %v", err)
	}
}

func TestPDFService_Generate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	if _, err := exec.LookPath("weasyprint"); err != nil {
		t.Skip("weasyprint not installed")
	}

	svc := NewPDFService(t.TempDir())

	markdown := `# John Doe
## Experience
- Built things at Acme Corp`

	result, err := svc.Generate(context.Background(), markdown, resumeCSS, "resume", "Acme Corp", "Software Engineer")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	stat, err := os.Stat(result.Path)
	if err != nil {
		t.Fatalf("output file not found at %q: %v", result.Path, err)
	}
	if stat.Size() == 0 {
		t.Error("output PDF is empty")
	}
	if result.Filename != "Acme-Corp-Software-Engineer-Resume.pdf" {
		t.Errorf("filename = %q, want %q", result.Filename, "Acme-Corp-Software-Engineer-Resume.pdf")
	}

	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("cannot read output PDF: %v", err)
	}
	if !strings.HasPrefix(string(content), "%PDF") {
		t.Error("output does not start with PDF header")
	}
}

func TestPDFService_Generate_CoverLetter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	if _, err := exec.LookPath("weasyprint"); err != nil {
		t.Skip("weasyprint not installed")
	}

	svc := NewPDFService(t.TempDir())

	markdown := "Dear Hiring Manager,\n\nI am writing to express my interest..."

	result, err := svc.Generate(context.Background(), markdown, coverLetterCSS, "cover-letter", "Acme Corp", "Product Manager")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(result.Filename, "CoverLetter") {
		t.Errorf("filename %q should contain CoverLetter", result.Filename)
	}

	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("cannot read output PDF: %v", err)
	}
	if !strings.HasPrefix(string(content), "%PDF") {
		t.Error("output does not start with PDF header")
	}
}

func TestCheckBinary(t *testing.T) {
	t.Parallel()

	err := checkBinary("this-binary-does-not-exist-xyzzy")
	if err == nil {
		t.Error("expected error for missing binary, got nil")
	}
	if !strings.Contains(err.Error(), "not found in PATH") {
		t.Errorf("expected 'not found in PATH' in error, got: %v", err)
	}
}
