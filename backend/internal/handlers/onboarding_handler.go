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
	ProfileStandardID     string `json:"profileStandardId" binding:"required"`
	CoverLetterStandardID string `json:"coverLetterStandardId" binding:"required"`
	PDFPipelineStandardID string `json:"pdfPipelineStandardId" binding:"required"`
	PDFGenerationSkillID  string `json:"pdfGenerationSkillId" binding:"required"`
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

	c.JSON(http.StatusOK, Success(toStatusResponse(state)))
}

func (h *OnboardingHandler) SaveProfile(c *gin.Context) {
	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	state, err := h.svc.SaveProfile(c, req.ProfileStandardID, req.CoverLetterStandardID, req.PDFPipelineStandardID, req.PDFGenerationSkillID)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(toStatusResponse(state)))
}

func (h *OnboardingHandler) GetStatus(c *gin.Context) {
	state, err := h.svc.GetStatus(c)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusOK, Success(onboardingStatusResponse{
				Onboarded: false,
				Steps: onboardingSteps{
					LLMConfig:  false,
					Profile:    false,
					OrkaiSetup: false,
				},
			}))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(toStatusResponse(state)))
}

func toStatusResponse(state models.OnboardingState) onboardingStatusResponse {
	llmDone := state.LLMProvider != "" && state.LLMModel != ""
	profileDone := state.CanonicalProfileStandardID != "" &&
		state.CoverLetterPrinciplesStandardID != "" &&
		state.PDFPipelineStandardID != "" &&
		state.PDFGenerationSkillID != ""
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
