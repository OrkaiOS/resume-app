package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type profileService interface {
	Get(ctx context.Context) (models.Profile, error)
	Upsert(ctx context.Context, p models.Profile) (models.Profile, error)
}

type ProfileHandler struct {
	svc profileService
}

func NewProfileHandler(svc profileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

type profileResponse struct {
	ID                  string                    `json:"id"`
	FullName            string                    `json:"fullName"`
	Email               string                    `json:"email"`
	Phone               string                    `json:"phone"`
	Location            string                    `json:"location"`
	LinkedInURL         string                    `json:"linkedinUrl"`
	WebsiteURL          string                    `json:"websiteUrl"`
	GitHubURL           string                    `json:"githubUrl"`
	ProfessionalSummary string                    `json:"professionalSummary"`
	WorkExperience      models.WorkExperienceList `json:"workExperience"`
	Education           models.EducationList      `json:"education"`
	Skills              models.SkillCategoryList  `json:"skills"`
	Projects            models.ProjectList        `json:"projects"`
	Certifications      models.CertificationList  `json:"certifications"`
	Languages           models.LanguageList       `json:"languages"`
	CreatedAt           string                    `json:"createdAt"`
	UpdatedAt           string                    `json:"updatedAt"`
}

type profileUpsertRequest struct {
	FullName            string                    `json:"fullName"`
	Email               string                    `json:"email"`
	Phone               string                    `json:"phone"`
	Location            string                    `json:"location"`
	LinkedInURL         string                    `json:"linkedinUrl"`
	WebsiteURL          string                    `json:"websiteUrl"`
	GitHubURL           string                    `json:"githubUrl"`
	ProfessionalSummary string                    `json:"professionalSummary"`
	WorkExperience      models.WorkExperienceList `json:"workExperience"`
	Education           models.EducationList      `json:"education"`
	Skills              models.SkillCategoryList  `json:"skills"`
	Projects            models.ProjectList        `json:"projects"`
	Certifications      models.CertificationList  `json:"certifications"`
	Languages           models.LanguageList       `json:"languages"`
}

func profileToResponse(p models.Profile) profileResponse {
	return profileResponse{
		ID:                  p.ID,
		FullName:            p.FullName,
		Email:               p.Email,
		Phone:               p.Phone,
		Location:            p.Location,
		LinkedInURL:         p.LinkedInURL,
		WebsiteURL:          p.WebsiteURL,
		GitHubURL:           p.GitHubURL,
		ProfessionalSummary: p.ProfessionalSummary,
		WorkExperience:      p.WorkExperience,
		Education:           p.Education,
		Skills:              p.Skills,
		Projects:            p.Projects,
		Certifications:      p.Certifications,
		Languages:           p.Languages,
		CreatedAt:           p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ProfileHandler) Get(c *gin.Context) {
	p, err := h.svc.Get(c)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusOK, Success(nil))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(profileToResponse(p)))
}

func (h *ProfileHandler) Upsert(c *gin.Context) {
	var req profileUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

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

	created, err := h.svc.Upsert(c, p)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(profileToResponse(created)))
}
