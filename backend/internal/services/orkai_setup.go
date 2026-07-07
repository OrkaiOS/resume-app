package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/orkai"
	"github.com/marco/resume-app/internal/store"
)

type OrkaiClient interface {
	CreateCategory(ctx context.Context, name string) (string, error)
	CreateStandard(ctx context.Context, name, text, categoryID string) (string, error)
	CreateSkill(ctx context.Context, name, text, categoryID string) (string, error)
	LinkEntities(ctx context.Context, sourceID, targetID string) error
}

type SetupStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type OrkaiSetupService struct {
	client OrkaiClient
	store  store.OnboardingStore
	mu     sync.Mutex
	detect func() (string, error)
}

func NewOrkaiSetupService(client OrkaiClient, s store.OnboardingStore) *OrkaiSetupService {
	return &OrkaiSetupService{
		client: client,
		store:  s,
		detect: orkai.DetectMCPToken,
	}
}

func NewOrkaiSetupServiceWithDetect(client OrkaiClient, s store.OnboardingStore, detect func() (string, error)) *OrkaiSetupService {
	return &OrkaiSetupService{
		client: client,
		store:  s,
		detect: detect,
	}
}

var stepNames = []string{
	"Create personal category",
	"Create Canonical Profile standard",
	"Create Cover Letter Principles standard",
	"Create PDF Pipeline standard",
	"Create PDF Generation skill",
	"Link entities",
	"Verify MCP token",
}

const coverLetterPrinciplesTemplate = `# Cover Letter Writing Principles

## Structure
- Three-paragraph format: introduction, body, closing.
- Never mention the referrer's name in the body.
- Never make the current employer the subject of a paragraph.

## Tone
- Professional, confident, and concise.
- Avoid buzzwords and filler phrases.

## Pre-submission Checklist
- Company name and role title are correct.
- No internal product or framework names are exposed.
- Word count is between 250 and 400 words.
- PDF output is exactly one page.
`

const pdfPipelineTemplate = `# Resume & Cover Letter PDF Pipeline

## Tooling
- PDF generation uses WeasyPrint (macOS: brew install weasyprint).
- Source documents are Markdown, styled with CSS.

## CSS Settings (2-page resume, 1-page cover letter)
- Page size: A4
- Margins: 0.85in all sides
- Body font: 10.5pt
- Heading scale: h1 16pt, h2 13pt, h3 11pt
- page-break-after: avoid on h2 and h3

## Makefile Reference
` + "```makefile" + `
resume.pdf: resume.md resume.css
	weasyprint resume.md resume.pdf -s resume.css

cover-letter.pdf: cover-letter.md cover-letter.css
	weasyprint cover-letter.md cover-letter.pdf -s cover-letter.css
` + "```" + `

## Verification
- Verify page count with: pdfinfo file.pdf | grep Pages
- Ignore WeasyPrint Cairo/GLib warnings.
`

const pdfGenerationSkillTemplate = `# Resume and Cover Letter PDF Generation

## Steps
1. Write the resume or cover letter in Markdown.
2. Apply the CSS from the PDF Pipeline standard.
3. Run WeasyPrint to generate the PDF.
4. Verify the page count matches the target.
5. Optionally generate a PNG preview with: pdftoppm -png -r 150 file.pdf preview

## Notes
- Always use the canonical profile data from the Profile standard.
- Never hardcode personal information in the CSS or Makefile.
`

func (s *OrkaiSetupService) RunSetup(ctx context.Context, profile models.Profile) (string, <-chan SetupStep, error) {
	sessionID := uuid.New().String()
	ch := make(chan SetupStep, 7)

	go func() {
		defer close(ch)

		profileBody := fmt.Sprintf(
			"# %s — Canonical Profile\n\nWhen any source document disagrees with this standard, this standard wins.\n\n%s",
			profile.FullName,
			buildProfileText(profile),
		)

		profileStdName := fmt.Sprintf("%s — Canonical Profile for Resume & Cover Letter Generation", profile.FullName)
		coverLetterStdName := fmt.Sprintf("Cover Letter Writing Principles — Personal Workspace (%s)", profile.FullName)
		pdfPipelineStdName := "Resume & Cover Letter PDF Pipeline — macOS Tooling & Senior CSS Tuning"
		pdfSkillName := "Resume and Cover Letter PDF Generation"

		categoryID, err := s.client.CreateCategory(ctx, "personal")
		ch <- stepResult(0, err)
		if err != nil {
			return
		}

		profileStdID, err := s.client.CreateStandard(ctx, profileStdName, profileBody, categoryID)
		ch <- stepResult(1, err)
		if err != nil {
			return
		}

		coverLetterStdID, err := s.client.CreateStandard(ctx, coverLetterStdName, coverLetterPrinciplesTemplate, categoryID)
		ch <- stepResult(2, err)
		if err != nil {
			return
		}

		pdfPipelineStdID, err := s.client.CreateStandard(ctx, pdfPipelineStdName, pdfPipelineTemplate, categoryID)
		ch <- stepResult(3, err)
		if err != nil {
			return
		}

		pdfSkillID, err := s.client.CreateSkill(ctx, pdfSkillName, pdfGenerationSkillTemplate, categoryID)
		ch <- stepResult(4, err)
		if err != nil {
			return
		}

		linkErr := s.client.LinkEntities(ctx, pdfPipelineStdID, pdfSkillID)
		if linkErr == nil {
			linkErr = s.client.LinkEntities(ctx, coverLetterStdID, profileStdID)
		}
		ch <- stepResult(5, linkErr)
		if linkErr != nil {
			return
		}

		_, tokenErr := s.detect()
		if tokenErr != nil {
			ch <- SetupStep{Name: stepNames[6], Status: "failed", Error: tokenErr.Error()}
			return
		}
		ch <- SetupStep{Name: stepNames[6], Status: "success"}

		if err := s.store.UpsertOrkaiIDs(ctx, categoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID); err != nil {
			log.Printf("services.OrkaiSetupService.RunSetup: UpsertOrkaiIDs: %v", err)
		}
	}()

	return sessionID, ch, nil
}

func stepResult(idx int, err error) SetupStep {
	step := SetupStep{Name: stepNames[idx]}
	if err != nil {
		step.Status = "failed"
		step.Error = err.Error()
	} else {
		step.Status = "success"
	}
	return step
}

func buildProfileText(profile models.Profile) string {
	var s string

	if profile.Email != "" || profile.Phone != "" || profile.Location != "" {
		s += "## Contact\n"
		if profile.Email != "" {
			s += fmt.Sprintf("- Email: %s\n", profile.Email)
		}
		if profile.Phone != "" {
			s += fmt.Sprintf("- Phone: %s\n", profile.Phone)
		}
		if profile.Location != "" {
			s += fmt.Sprintf("- Location: %s\n", profile.Location)
		}
		if profile.LinkedInURL != "" {
			s += fmt.Sprintf("- LinkedIn: %s\n", profile.LinkedInURL)
		}
		if profile.WebsiteURL != "" {
			s += fmt.Sprintf("- Website: %s\n", profile.WebsiteURL)
		}
		if profile.GitHubURL != "" {
			s += fmt.Sprintf("- GitHub: %s\n", profile.GitHubURL)
		}
		s += "\n"
	}

	if profile.ProfessionalSummary != "" {
		s += "## Summary\n" + profile.ProfessionalSummary + "\n\n"
	}

	if len(profile.WorkExperience) > 0 {
		s += "## Experience\n"
		for _, exp := range profile.WorkExperience {
			s += fmt.Sprintf("### %s at %s\n", exp.JobTitle, exp.Company)
			if exp.Location != "" {
				s += fmt.Sprintf("- Location: %s\n", exp.Location)
			}
			s += fmt.Sprintf("- %s to %s\n", exp.StartDate, exp.EndDate)
			if exp.Description != "" {
				s += exp.Description + "\n"
			}
			s += "\n"
		}
	}

	if len(profile.Education) > 0 {
		s += "## Education\n"
		for _, edu := range profile.Education {
			s += fmt.Sprintf("### %s — %s\n", edu.Degree, edu.Institution)
			if edu.Location != "" {
				s += fmt.Sprintf("- Location: %s\n", edu.Location)
			}
			s += fmt.Sprintf("- %s to %s\n", edu.StartDate, edu.EndDate)
			if edu.GPA != "" {
				s += fmt.Sprintf("- GPA: %s\n", edu.GPA)
			}
			s += "\n"
		}
	}

	if len(profile.Skills) > 0 {
		s += "## Skills\n"
		for _, sk := range profile.Skills {
			s += fmt.Sprintf("- %s: %s\n", sk.Name, joinAny(sk.Skills))
		}
		s += "\n"
	}

	if len(profile.Projects) > 0 {
		s += "## Projects\n"
		for _, p := range profile.Projects {
			s += fmt.Sprintf("### %s\n", p.Name)
			if p.Role != "" {
				s += fmt.Sprintf("- Role: %s\n", p.Role)
			}
			if p.Description != "" {
				s += p.Description + "\n"
			}
			if len(p.Technologies) > 0 {
				s += fmt.Sprintf("- Technologies: %s\n", joinAny(p.Technologies))
			}
			if p.URL != "" {
				s += fmt.Sprintf("- URL: %s\n", p.URL)
			}
			s += "\n"
		}
	}

	if len(profile.Certifications) > 0 {
		s += "## Certifications\n"
		for _, cert := range profile.Certifications {
			s += fmt.Sprintf("- %s (%s)\n", cert.Name, cert.IssuingOrg)
		}
		s += "\n"
	}

	if len(profile.Languages) > 0 {
		s += "## Languages\n"
		for _, lang := range profile.Languages {
			s += fmt.Sprintf("- %s: %s\n", lang.Name, lang.Proficiency)
		}
		s += "\n"
	}

	return s
}

func joinAny(items []string) string {
	var out string
	for i, item := range items {
		if i > 0 {
			out += ", "
		}
		out += item
	}
	return out
}
