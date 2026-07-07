package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
)

const maxUploadSize = 10 << 20

type profileParser interface {
	ParsePDF(r io.Reader) (*models.Profile, error)
	ParseMarkdown(r io.Reader) (*models.Profile, error)
}

type profileLLMParser interface {
	Parse(ctx context.Context, text string, provider, model, apiKey string) (*models.Profile, error)
}

type llmConfigGetter interface {
	GetStatus(ctx context.Context) (models.OnboardingState, error)
}

type ProfileUploadHandler struct {
	parser     profileParser
	llmParser  profileLLMParser
	profiles   profileService
	onboarding llmConfigGetter
}

func NewProfileUploadHandler(parser profileParser, llmParser profileLLMParser, profiles profileService, onboarding llmConfigGetter) *ProfileUploadHandler {
	return &ProfileUploadHandler{
		parser:     parser,
		llmParser:  llmParser,
		profiles:   profiles,
		onboarding: onboarding,
	}
}

func (h *ProfileUploadHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "file is required: "+err.Error())
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))

	var profile *models.Profile
	switch ext {
	case ".pdf":
		profile, err = h.parser.ParsePDF(file)
	case ".md", ".markdown":
		profile, err = h.parser.ParseMarkdown(file)
	default:
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "unsupported file type: "+ext+" (use .pdf or .md)")
		return
	}

	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	if ext == ".pdf" {
		rawText := profile.ProfessionalSummary
		if rawText != "" {
			llmProfile, llmErr := h.parseWithLLM(c, rawText)
			if llmErr != nil {
				log.Printf("handlers.ProfileUploadHandler.Upload: LLM parse failed, using raw text: %v", llmErr)
			} else if llmProfile != nil {
				profile = llmProfile
			}
		}
	}

	normalizeProfile(profile)

	if profile != nil && profile.FullName != "" {
		if _, err := h.profiles.Upsert(c, *profile); err != nil {
			status, code := mapError(err)
			respondError(c, status, code, err.Error())
			return
		}
	}

	c.JSON(http.StatusOK, Success(profileToResponse(*profile)))
}

func (h *ProfileUploadHandler) parseWithLLM(ctx context.Context, text string) (*models.Profile, error) {
	state, err := h.onboarding.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("handlers.ProfileUploadHandler.parseWithLLM: get onboarding state: %w", err)
	}
	if state.LLMProvider == "" || state.LLMModel == "" {
		return nil, fmt.Errorf("handlers.ProfileUploadHandler.parseWithLLM: LLM not configured")
	}
	return h.llmParser.Parse(ctx, text, state.LLMProvider, state.LLMModel, state.LLMAPIKey)
}

func normalizeProfile(p *models.Profile) {
	if p == nil {
		return
	}
	if p.WorkExperience == nil {
		p.WorkExperience = models.WorkExperienceList{}
	}
	if p.Education == nil {
		p.Education = models.EducationList{}
	}
	if p.Skills == nil {
		p.Skills = models.SkillCategoryList{}
	}
	if p.Projects == nil {
		p.Projects = models.ProjectList{}
	}
	if p.Certifications == nil {
		p.Certifications = models.CertificationList{}
	}
	if p.Languages == nil {
		p.Languages = models.LanguageList{}
	}
}
