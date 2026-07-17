package services

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed templates/resume.css
var resumeCSS string

//go:embed templates/cover-letter.css
var coverLetterCSS string

var sanitizePattern = regexp.MustCompile(`[^a-zA-Z0-9\-_ ]+`)

type PDFService struct {
	outputDir string
}

type PDFResult struct {
	Path     string
	Filename string
}

func NewPDFService(outputDir string) *PDFService {
	return &PDFService{outputDir: outputDir}
}

func ResumeCSS() string      { return resumeCSS }
func CoverLetterCSS() string { return coverLetterCSS }

func (s *PDFService) Generate(ctx context.Context, markdown, css, docType, company, role string) (*PDFResult, error) {
	if err := checkBinary("pandoc"); err != nil {
		return nil, err
	}
	if err := checkBinary("weasyprint"); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("pdf_service.Generate: cannot create output dir %q: %w", s.outputDir, err)
	}

	workDir, err := os.MkdirTemp("", "pdf-gen-*")
	if err != nil {
		return nil, fmt.Errorf("pdf_service.Generate: cannot create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	mdPath := filepath.Join(workDir, "input.md")
	if err := os.WriteFile(mdPath, []byte(markdown), 0644); err != nil {
		return nil, fmt.Errorf("pdf_service.Generate: cannot write markdown: %w", err)
	}

	htmlPath := filepath.Join(workDir, "intermediate.html")
	pandocCmd := exec.CommandContext(ctx, "pandoc", mdPath, "-o", htmlPath)
	if out, err := pandocCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdf_service.Generate: pandoc failed: %w\n%s", err, string(out))
	}

	cssPath := filepath.Join(workDir, "style.css")
	if err := os.WriteFile(cssPath, []byte(css), 0644); err != nil {
		return nil, fmt.Errorf("pdf_service.Generate: cannot write CSS: %w", err)
	}

	sanitizedCompany := sanitize(company)
	sanitizedRole := sanitize(role)
	docLabel := "Resume"
	if docType == "cover-letter" {
		docLabel = "CoverLetter"
	}
	filename := fmt.Sprintf("%s-%s-%s.pdf", sanitizedCompany, sanitizedRole, docLabel)
	outPath := filepath.Join(s.outputDir, filename)

	weasyCmd := exec.CommandContext(ctx, "weasyprint", htmlPath, outPath, "-s", cssPath)
	if out, err := weasyCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdf_service.Generate: weasyprint failed: %w\n%s", err, string(out))
	}

	return &PDFResult{Path: outPath, Filename: filename}, nil
}

func checkBinary(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("pdf_service.Generate: %s binary not found in PATH (is it installed?)", name)
	}
	return nil
}

func sanitize(s string) string {
	cleaned := sanitizePattern.ReplaceAllString(s, "")
	cleaned = strings.Join(strings.Fields(cleaned), "-")
	if cleaned == "" {
		return "Untitled"
	}
	return cleaned
}
