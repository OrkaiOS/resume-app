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

type coverLetterService interface {
	GetByOpportunity(ctx context.Context, opportunityID string) (models.CoverLetter, error)
	Upsert(ctx context.Context, cl models.CoverLetter) (models.CoverLetter, error)
}

type CoverLetterHandler struct {
	svc coverLetterService
}

func NewCoverLetterHandler(svc coverLetterService) *CoverLetterHandler {
	return &CoverLetterHandler{svc: svc}
}

type coverLetterResponse struct {
	ID              string `json:"id"`
	OpportunityID   string `json:"opportunityId"`
	MarkdownContent string `json:"markdownContent"`
	PDFPath         string `json:"pdfPath"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type coverLetterUpsertRequest struct {
	MarkdownContent string `json:"markdownContent"`
	PDFPath         string `json:"pdfPath"`
}

func coverLetterToResponse(cl models.CoverLetter) coverLetterResponse {
	return coverLetterResponse{
		ID:              cl.ID,
		OpportunityID:   cl.OpportunityID,
		MarkdownContent: cl.MarkdownContent,
		PDFPath:         cl.PDFPath,
		Status:          cl.Status,
		CreatedAt:       cl.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       cl.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *CoverLetterHandler) GetByOpportunity(c *gin.Context) {
	opportunityID := c.Param("id")

	cl, err := h.svc.GetByOpportunity(c, opportunityID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusOK, Success(nil))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(coverLetterToResponse(cl)))
}

func (h *CoverLetterHandler) Upsert(c *gin.Context) {
	opportunityID := c.Param("id")

	var req coverLetterUpsertRequest
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

	cl := models.CoverLetter{
		ID:              existing.ID,
		OpportunityID:   opportunityID,
		MarkdownContent: req.MarkdownContent,
		PDFPath:         req.PDFPath,
		Status:          existing.Status,
	}

	upserted, err := h.svc.Upsert(c, cl)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(coverLetterToResponse(upserted)))
}

func (h *CoverLetterHandler) DownloadPDF(c *gin.Context) {
	opportunityID := c.Param("id")

	cl, err := h.svc.GetByOpportunity(c, opportunityID)
	if err != nil {
		if isNotFound(err) {
			respondError(c, http.StatusNotFound, ErrCodeNotFound, "no cover letter found for this opportunity")
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}
	if cl.PDFPath == "" {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "no PDF generated for this cover letter yet")
		return
	}
	filename := filepath.Base(cl.PDFPath)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
	c.File(cl.PDFPath)
}
