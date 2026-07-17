package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type resumeService interface {
	GetByOpportunity(ctx context.Context, opportunityID string) (models.Resume, error)
	Upsert(ctx context.Context, r models.Resume) (models.Resume, error)
}

type ResumeHandler struct {
	svc resumeService
}

func NewResumeHandler(svc resumeService) *ResumeHandler {
	return &ResumeHandler{svc: svc}
}

type resumeResponse struct {
	ID              string `json:"id"`
	OpportunityID   string `json:"opportunityId"`
	MarkdownContent string `json:"markdownContent"`
	PDFPath         string `json:"pdfPath"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type resumeUpsertRequest struct {
	MarkdownContent string `json:"markdownContent"`
	PDFPath         string `json:"pdfPath"`
}

func resumeToResponse(r models.Resume) resumeResponse {
	return resumeResponse{
		ID:              r.ID,
		OpportunityID:   r.OpportunityID,
		MarkdownContent: r.MarkdownContent,
		PDFPath:         r.PDFPath,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ResumeHandler) GetByOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	r, err := h.svc.GetByOpportunity(c, opportunityID)
	if err != nil {
		if isNotFound(err) {
			c.JSON(http.StatusOK, Success(nil))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(resumeToResponse(r)))
}

func (h *ResumeHandler) Upsert(c *gin.Context) {
	opportunityID := c.Param("id")

	var req resumeUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	existing, err := h.svc.GetByOpportunity(c, opportunityID)
	if err != nil && !isNotFound(err) {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	r := models.Resume{
		ID:              existing.ID,
		OpportunityID:   opportunityID,
		MarkdownContent: req.MarkdownContent,
		PDFPath:         req.PDFPath,
		Status:          existing.Status,
	}

	upserted, err := h.svc.Upsert(c, r)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(resumeToResponse(upserted)))
}

func (h *ResumeHandler) DownloadPDF(c *gin.Context) {
	opportunityID := c.Param("id")

	r, err := h.svc.GetByOpportunity(c, opportunityID)
	if err != nil {
		if isNotFound(err) {
			respondError(c, http.StatusNotFound, ErrCodeNotFound, "no resume found for this opportunity")
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}
	if r.PDFPath == "" {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "no PDF generated for this resume yet")
		return
	}
	filename := filepath.Base(r.PDFPath)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
	c.File(r.PDFPath)
}

func isNotFound(err error) bool {
	return errors.Is(err, store.ErrNotFound)
}
