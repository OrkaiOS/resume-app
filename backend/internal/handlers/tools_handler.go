package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/services"
)

// ToolsHandler exposes the agent tools as REST endpoints for contract
// documentation and manual debugging. The agent itself calls the
// services.ToolRegistry directly (not these HTTP endpoints) during
// chat. The handler depends on the single ToolsService interface
// rather than multiple store-like providers, keeping the handler layer
// free of business-logic dependencies.
// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
type ToolsHandler struct {
	svc toolsService
}

// toolsService is the service interface the handler depends on.
// Implemented by services.ToolsService.
type toolsService interface {
	Shell(ctx context.Context, command, language string) (services.ShellResult, error)
	OrkaiSearch(ctx context.Context, query string) (string, error)
	Profile(ctx context.Context) (any, error)
}

func NewToolsHandler(svc toolsService) *ToolsHandler {
	return &ToolsHandler{svc: svc}
}

type shellRequest struct {
	Command  string `json:"command" binding:"required"`
	Language string `json:"language"`
}

func (h *ToolsHandler) Shell(c *gin.Context) {
	var req shellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	result, err := h.svc.Shell(c.Request.Context(), req.Command, req.Language)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(result))
}

type orkaiSearchRequest struct {
	Query string `json:"query" binding:"required"`
}

func (h *ToolsHandler) OrkaiSearch(c *gin.Context) {
	var req orkaiSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	result, err := h.svc.OrkaiSearch(c.Request.Context(), req.Query)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeInternal, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(map[string]string{"results": result}))
}

func (h *ToolsHandler) Profile(c *gin.Context) {
	profile, err := h.svc.Profile(c.Request.Context())
	if err != nil {
		if errors.Is(err, services.ErrProfileNotFound) {
			c.JSON(http.StatusOK, Success(nil))
			return
		}
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(profile))
}
