package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
)

type opportunityService interface {
	List(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error)
	Get(ctx context.Context, id string) (models.Opportunity, error)
	Create(ctx context.Context, o models.Opportunity) (models.Opportunity, error)
	Update(ctx context.Context, o models.Opportunity) (models.Opportunity, error)
	Delete(ctx context.Context, id string) error
	SetArchived(ctx context.Context, id string, archived bool) (models.Opportunity, error)
}

type OpportunityHandler struct {
	svc opportunityService
}

func NewOpportunityHandler(svc opportunityService) *OpportunityHandler {
	return &OpportunityHandler{svc: svc}
}

type opportunityResponse struct {
	ID          string `json:"id"`
	Company     string `json:"company"`
	Role        string `json:"role"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type opportunityListResponse struct {
	Items      []opportunityResponse `json:"items"`
	NextCursor string                `json:"nextCursor"`
}

type opportunityCreateRequest struct {
	Company     string `json:"company" binding:"required"`
	Role        string `json:"role" binding:"required"`
	Description string `json:"description"`
}

type opportunityUpdateRequest struct {
	Company     string `json:"company" binding:"required"`
	Role        string `json:"role" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type opportunityArchiveRequest struct {
	Archived bool `json:"archived"`
}

func opportunityToResponse(o models.Opportunity) opportunityResponse {
	return opportunityResponse{
		ID:          o.ID,
		Company:     o.Company,
		Role:        o.Role,
		Description: o.Description,
		Status:      o.Status,
		CreatedAt:   o.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   o.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *OpportunityHandler) List(c *gin.Context) {
	cursor := c.Query("cursor")
	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			respondError(c, http.StatusBadRequest, ErrCodeValidation, "limit must be a positive integer")
			return
		}
	}

	items, nextCursor, err := h.svc.List(c, cursor, limit)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	resp := make([]opportunityResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, opportunityToResponse(item))
	}

	c.JSON(http.StatusOK, Success(opportunityListResponse{
		Items:      resp,
		NextCursor: nextCursor,
	}))
}

func (h *OpportunityHandler) Get(c *gin.Context) {
	id := c.Param("id")

	o, err := h.svc.Get(c, id)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(opportunityToResponse(o)))
}

func (h *OpportunityHandler) Create(c *gin.Context) {
	var req opportunityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	o := models.Opportunity{
		Company:     req.Company,
		Role:        req.Role,
		Description: req.Description,
	}

	created, err := h.svc.Create(c, o)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusCreated, Success(opportunityToResponse(created)))
}

func (h *OpportunityHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req opportunityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	o := models.Opportunity{
		ID:          id,
		Company:     req.Company,
		Role:        req.Role,
		Description: req.Description,
		Status:      req.Status,
	}

	updated, err := h.svc.Update(c, o)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(opportunityToResponse(updated)))
}

func (h *OpportunityHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(c, id); err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(nil))
}

func (h *OpportunityHandler) SetArchived(c *gin.Context) {
	id := c.Param("id")

	var req opportunityArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "invalid request body: "+err.Error())
		return
	}

	updated, err := h.svc.SetArchived(c, id, req.Archived)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	c.JSON(http.StatusOK, Success(opportunityToResponse(updated)))
}
