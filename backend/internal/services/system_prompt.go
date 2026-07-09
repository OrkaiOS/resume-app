package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/orkai"
	"github.com/marco/resume-app/internal/store"
)

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision SystemPromptService assembles the agent system prompt at session start from orkai standards + local profile data. Every claim must be traceable to the profile, the job description, the User Insights, or the writing principles. The service gracefully degrades when orkai is unreachable — profile-only prompt still works.
type SystemPromptService struct {
	onboardingStore  store.OnboardingStore
	profileStore     store.ProfileStore
	opportunityStore store.OpportunityStore
	orkaiClient      *orkai.OrkaiClient
	sessionSvc       *SessionService
}

func NewSystemPromptService(
	onboardingStore store.OnboardingStore,
	profileStore store.ProfileStore,
	opportunityStore store.OpportunityStore,
	orkaiClient *orkai.OrkaiClient,
	sessionSvc *SessionService,
) *SystemPromptService {
	return &SystemPromptService{
		onboardingStore:  onboardingStore,
		profileStore:     profileStore,
		opportunityStore: opportunityStore,
		orkaiClient:      orkaiClient,
		sessionSvc:       sessionSvc,
	}
}

const mandatorySourceRule = `
## CRITICAL RULES

1. You must use the provided sources. Do not assume or fabricate information.
   Every claim must be traceable to the profile, the job description, the
   User Insights, or the writing principles.
2. The canonical profile is the single source of truth. When any user-uploaded
   file or older document disagrees with the profile, the profile wins.
3. When the system prompt alone is insufficient, retrieve additional context
   from orkai:
   - Use orkai_get(id) when you know the entity ID (listed with each source below).
   - Use orkai_search(query) only when you do NOT have an ID and need to discover
     relevant standards, skills, or documents by topic.`

func (s *SystemPromptService) Build(ctx context.Context, opportunityID string) string {
	var b strings.Builder

	b.WriteString("You are a professional resume and cover letter writing assistant. ")

	profile, profileErr := s.profileStore.Get(ctx)
	if profileErr == nil {
		b.WriteString("You have access to the user's full professional profile, ")
		b.WriteString("orkai standards for document generation, and the job opportunity details. ")
		s.appendProfile(&b, profile)
	} else {
		b.WriteString("The user's profile is not yet configured. ")
		b.WriteString("Ask them to complete onboarding first. ")
	}

	if opportunityID != "" {
		opp, err := s.opportunityStore.Get(ctx, opportunityID)
		if err == nil {
			s.appendOpportunity(&b, opp)
		}
	}

	s.appendOrkaiSources(&b, ctx)

	b.WriteString(mandatorySourceRule)

	return b.String()
}

func (s *SystemPromptService) appendProfile(b *strings.Builder, profile models.Profile) {
	b.WriteString("\n\n## USER PROFILE\n\n")
	if profile.FullName != "" {
		b.WriteString("Name: " + profile.FullName + "\n")
	}
	if profile.Email != "" {
		b.WriteString("Email: " + profile.Email + "\n")
	}
	if profile.Phone != "" {
		b.WriteString("Phone: " + profile.Phone + "\n")
	}
	if profile.Location != "" {
		b.WriteString("Location: " + profile.Location + "\n")
	}
	if profile.LinkedInURL != "" {
		b.WriteString("LinkedIn: " + profile.LinkedInURL + "\n")
	}
	if profile.GitHubURL != "" {
		b.WriteString("GitHub: " + profile.GitHubURL + "\n")
	}
	if profile.WebsiteURL != "" {
		b.WriteString("Website: " + profile.WebsiteURL + "\n")
	}
	if profile.ProfessionalSummary != "" {
		b.WriteString("\nProfessional Summary:\n" + profile.ProfessionalSummary + "\n")
	}

	if len(profile.WorkExperience) > 0 {
		b.WriteString("\nWork Experience:\n")
		for _, w := range profile.WorkExperience {
			b.WriteString(fmt.Sprintf("- %s at %s (%s to %s)\n", w.JobTitle, w.Company, w.StartDate, endDateOrPresent(w.EndDate)))
			if w.Description != "" {
				b.WriteString("  " + w.Description + "\n")
			}
		}
	}

	if len(profile.Education) > 0 {
		b.WriteString("\nEducation:\n")
		for _, e := range profile.Education {
			b.WriteString(fmt.Sprintf("- %s from %s (%s to %s)\n", e.Degree, e.Institution, e.StartDate, endDateOrPresent(e.EndDate)))
		}
	}

	if len(profile.Skills) > 0 {
		b.WriteString("\nSkills:\n")
		for _, cat := range profile.Skills {
			b.WriteString(fmt.Sprintf("- %s: %s\n", cat.Name, strings.Join(cat.Skills, ", ")))
		}
	}

	if len(profile.Projects) > 0 {
		b.WriteString("\nProjects:\n")
		for _, p := range profile.Projects {
			techStr := strings.Join(p.Technologies, ", ")
			b.WriteString(fmt.Sprintf("- %s: %s [%s]\n", p.Name, p.Description, techStr))
		}
	}

	if len(profile.Certifications) > 0 {
		b.WriteString("\nCertifications:\n")
		for _, cert := range profile.Certifications {
			b.WriteString(fmt.Sprintf("- %s (%s)\n", cert.Name, cert.IssuingOrg))
		}
	}

	if len(profile.Languages) > 0 {
		b.WriteString("\nLanguages:\n")
		for _, lang := range profile.Languages {
			b.WriteString(fmt.Sprintf("- %s: %s\n", lang.Name, lang.Proficiency))
		}
	}
}

func (s *SystemPromptService) appendOpportunity(b *strings.Builder, opp models.Opportunity) {
	b.WriteString("\n\n## JOB OPPORTUNITY\n\n")
	b.WriteString(fmt.Sprintf("Company: %s\n", opp.Company))
	b.WriteString(fmt.Sprintf("Role: %s\n", opp.Role))
	if opp.Description != "" {
		b.WriteString(fmt.Sprintf("Job Description:\n%s\n", opp.Description))
	}
}

func (s *SystemPromptService) appendOrkaiSources(b *strings.Builder, ctx context.Context) {
	state, err := s.onboardingStore.Get(ctx)
	if err != nil {
		return
	}

	if state.CanonicalProfileStandardID != "" {
		entity, err := s.orkaiClient.GetEntity(ctx, state.CanonicalProfileStandardID)
		if err == nil && entity.Text != "" {
			b.WriteString("\n\n## CANONICAL PROFILE (from orkai — authoritative source)\n\n")
			b.WriteString(entity.Text)
			b.WriteString(fmt.Sprintf("\n(orkai ID: %s)\n", state.CanonicalProfileStandardID))
		}
	}

	if state.CoverLetterPrinciplesStandardID != "" {
		entity, err := s.orkaiClient.GetEntity(ctx, state.CoverLetterPrinciplesStandardID)
		if err == nil && entity.Text != "" {
			b.WriteString("\n\n## COVER LETTER WRITING PRINCIPLES (from orkai)\n\n")
			b.WriteString(entity.Text)
			b.WriteString(fmt.Sprintf("\n(orkai ID: %s)\n", state.CoverLetterPrinciplesStandardID))
		}
	}

	if state.PDFGenerationSkillID != "" {
		entity, err := s.orkaiClient.GetEntity(ctx, state.PDFGenerationSkillID)
		if err == nil && entity.Text != "" {
			b.WriteString("\n\n## PDF GENERATION SKILL (from orkai)\n\n")
			b.WriteString(entity.Text)
			b.WriteString(fmt.Sprintf("\n(orkai ID: %s)\n", state.PDFGenerationSkillID))
		}
	}

	// Load User Insights standard if it exists (FR-031, FR-032).
	if s.sessionSvc != nil {
		insightsText, err := s.sessionSvc.GetUserInsightsText(ctx)
		if err == nil && insightsText != "" {
			b.WriteString("\n\n## USER INSIGHTS (from orkai)\n\n")
			b.WriteString(insightsText)
			b.WriteString("\n")
		}
	}
}

func endDateOrPresent(date string) string {
	if date == "" {
		return "Present"
	}
	return date
}
