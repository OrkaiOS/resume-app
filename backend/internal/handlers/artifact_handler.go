package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
)

type artifactService interface {
	List(ctx context.Context) ([]models.Artifact, error)
	Get(ctx context.Context, id string) (models.Artifact, error)
	Create(ctx context.Context, a models.Artifact) (models.Artifact, error)
	Delete(ctx context.Context, id string) error
}

type ArtifactHandler struct {
	svc artifactService
}

func NewArtifactHandler(svc artifactService) *ArtifactHandler {
	return &ArtifactHandler{svc: svc}
}

type artifactResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	ScriptContent string `json:"scriptContent"`
	UsageCount    int    `json:"usageCount"`
	CreatedAt     string `json:"createdAt"`
	LastUsedAt    string `json:"lastUsedAt"`
}

type artifactCreateRequest struct {
	Name          string `json:"name" binding:"required"`
	Type          string `json:"type" binding:"required"`
	Description   string `json:"description"`
	ScriptContent string `json:"scriptContent"`
}

func artifactToResponse(a models.Artifact) artifactResponse {
	return artifactResponse{
		ID:            a.ID,
		Name:          a.Name,
		Type:          a.Type,
		Description:   a.Description,
		ScriptContent: a.ScriptContent,
		UsageCount:    a.UsageCount,
		CreatedAt:     a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		LastUsedAt:    a.LastUsedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ArtifactHandler) List(c *gin.Context) {
	items, err := h.svc.List(c)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	resp := make([]artifactResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, artifactToResponse(item))
	}

	c.JSON(http.StatusOK, Success(resp))
}

func (h *ArtifactHandler) Create(c *gin.Context) {
	var req artifactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	a := models.Artifact{
		Name:          req.Name,
		Type:          req.Type,
		Description:   req.Description,
		ScriptContent: req.ScriptContent,
	}

	created, err := h.svc.Create(c, a)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusCreated, Success(artifactToResponse(created)))
}

func (h *ArtifactHandler) Get(c *gin.Context) {
	id := c.Param("id")

	a, err := h.svc.Get(c, id)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(artifactToResponse(a)))
}

func (h *ArtifactHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(c, id); err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(nil))
}
