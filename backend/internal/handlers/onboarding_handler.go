package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type onboardingService interface {
	GetStatus(ctx context.Context) (models.OnboardingState, error)
	SaveLLMConfig(ctx context.Context, provider, model, apiKey string) (models.OnboardingState, error)
	SaveProfile(ctx context.Context, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) (models.OnboardingState, error)
	SaveProfileData(ctx context.Context, p models.Profile) error
	HasProfile(ctx context.Context) (bool, error)
}

type OnboardingHandler struct {
	svc onboardingService
}

func NewOnboardingHandler(svc onboardingService) *OnboardingHandler {
	return &OnboardingHandler{svc: svc}
}

type llmConfigRequest struct {
	Provider string `json:"provider" binding:"required"`
	Model    string `json:"model" binding:"required"`
	APIKey   string `json:"apiKey"`
}

type profileRequest struct {
	FullName              string                    `json:"fullName"`
	Email                 string                    `json:"email"`
	Phone                 string                    `json:"phone"`
	Location              string                    `json:"location"`
	LinkedInURL           string                    `json:"linkedinUrl"`
	WebsiteURL            string                    `json:"websiteUrl"`
	GitHubURL             string                    `json:"githubUrl"`
	ProfessionalSummary   string                    `json:"professionalSummary"`
	WorkExperience        models.WorkExperienceList `json:"workExperience"`
	Education             models.EducationList      `json:"education"`
	Skills                models.SkillCategoryList  `json:"skills"`
	Projects              models.ProjectList        `json:"projects"`
	Certifications        models.CertificationList  `json:"certifications"`
	Languages             models.LanguageList       `json:"languages"`
	ProfileStandardID     string                    `json:"profileStandardId"`
	CoverLetterStandardID string                    `json:"coverLetterStandardId"`
	PDFPipelineStandardID string                    `json:"pdfPipelineStandardId"`
	PDFGenerationSkillID  string                    `json:"pdfGenerationSkillId"`
}

type onboardingStatusResponse struct {
	Onboarded bool            `json:"onboarded"`
	Steps     onboardingSteps `json:"steps"`
}

type onboardingSteps struct {
	LLMConfig  bool `json:"llmConfig"`
	Profile    bool `json:"profile"`
	OrkaiSetup bool `json:"orkaiSetup"`
}

func (h *OnboardingHandler) SaveLLMConfig(c *gin.Context) {
	var req llmConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	state, err := h.svc.SaveLLMConfig(c, req.Provider, req.Model, req.APIKey)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	hasProfile, err := h.svc.HasProfile(c)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}
	c.JSON(http.StatusOK, Success(toStatusResponse(state, hasProfile)))
}

func (h *OnboardingHandler) SaveProfile(c *gin.Context) {
	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	if req.FullName != "" {
		p := models.Profile{
			FullName:            req.FullName,
			Email:               req.Email,
			Phone:               req.Phone,
			Location:            req.Location,
			LinkedInURL:         req.LinkedInURL,
			WebsiteURL:          req.WebsiteURL,
			GitHubURL:           req.GitHubURL,
			ProfessionalSummary: req.ProfessionalSummary,
			WorkExperience:      req.WorkExperience,
			Education:           req.Education,
			Skills:              req.Skills,
			Projects:            req.Projects,
			Certifications:      req.Certifications,
			Languages:           req.Languages,
		}
		if err := h.svc.SaveProfileData(c, p); err != nil {
			status, code := mapError(err)
			respondError(c, status, code, err.Error())
			return
		}
	}

	hasProfile, hpErr := h.svc.HasProfile(c)
	if hpErr != nil {
		status, code := mapError(hpErr)
		respondError(c, status, code, hpErr.Error())
		return
	}

	if req.ProfileStandardID != "" || req.CoverLetterStandardID != "" ||
		req.PDFPipelineStandardID != "" || req.PDFGenerationSkillID != "" {
		state, err := h.svc.SaveProfile(c, req.ProfileStandardID, req.CoverLetterStandardID, req.PDFPipelineStandardID, req.PDFGenerationSkillID)
		if err != nil {
			status, code := mapError(err)
			respondError(c, status, code, err.Error())
			return
		}
		c.JSON(http.StatusOK, Success(toStatusResponse(state, hasProfile)))
		return
	}

	state, err := h.svc.GetStatus(c)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusOK, Success(toStatusResponse(models.OnboardingState{}, hasProfile)))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}
	c.JSON(http.StatusOK, Success(toStatusResponse(state, hasProfile)))
}

func (h *OnboardingHandler) GetStatus(c *gin.Context) {
	state, err := h.svc.GetStatus(c)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			hasProfile, hpErr := h.svc.HasProfile(c)
			if hpErr != nil {
				status, code := mapError(hpErr)
				respondError(c, status, code, hpErr.Error())
				return
			}
			c.JSON(http.StatusOK, Success(onboardingStatusResponse{
				Onboarded: false,
				Steps: onboardingSteps{
					LLMConfig:  false,
					Profile:    hasProfile,
					OrkaiSetup: false,
				},
			}))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	hasProfile, hpErr := h.svc.HasProfile(c)
	if hpErr != nil {
		status, code := mapError(hpErr)
		respondError(c, status, code, hpErr.Error())
		return
	}
	c.JSON(http.StatusOK, Success(toStatusResponse(state, hasProfile)))
}

func toStatusResponse(state models.OnboardingState, hasProfile bool) onboardingStatusResponse {
	llmDone := state.LLMProvider != "" && state.LLMModel != ""
	profileDone := hasProfile ||
		(state.CanonicalProfileStandardID != "" &&
			state.CoverLetterPrinciplesStandardID != "" &&
			state.PDFPipelineStandardID != "" &&
			state.PDFGenerationSkillID != "")
	orkaiDone := state.OrkaiCategoryID != ""

	return onboardingStatusResponse{
		Onboarded: !state.OnboardedAt.IsZero(),
		Steps: onboardingSteps{
			LLMConfig:  llmDone,
			Profile:    profileDone,
			OrkaiSetup: orkaiDone,
		},
	}
}
