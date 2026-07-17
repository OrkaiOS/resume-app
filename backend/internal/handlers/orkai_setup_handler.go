package handlers

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/services"
)

type orkaiSetupService interface {
	RunSetup(ctx context.Context, profile models.Profile, projectName string) (sessionID string, ch <-chan services.SetupStep, err error)
}

type profileGetter interface {
	Get(ctx context.Context) (models.Profile, error)
}

type OrkaiSetupHandler struct {
	svc     orkaiSetupService
	profile profileGetter
	mu      sync.Mutex
	steps   map[string][]services.SetupStep
}

func NewOrkaiSetupHandler(svc orkaiSetupService, profile profileGetter) *OrkaiSetupHandler {
	return &OrkaiSetupHandler{
		svc:     svc,
		profile: profile,
		steps:   make(map[string][]services.SetupStep),
	}
}

type orkaiSetupRequest struct {
	ProjectName string `json:"projectName" binding:"required"`
}

type orkaiSetupResponse struct {
	SessionID string `json:"sessionId"`
}

type setupStepResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type orkaiSetupStatusResponse struct {
	SessionID string              `json:"sessionId"`
	Steps     []setupStepResponse `json:"steps"`
	Completed bool                `json:"completed"`
}

func (h *OrkaiSetupHandler) StartSetup(c *gin.Context) {
	var req orkaiSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "projectName is required")
		return
	}

	p, err := h.profile.Get(c)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	sessionID, ch, err := h.svc.RunSetup(c, p, req.ProjectName)
	if err != nil {
		status, code := mapError(err)
		respondError(c, status, code, err.Error())
		return
	}

	h.mu.Lock()
	h.steps[sessionID] = make([]services.SetupStep, 8)
	for i, name := range stepNames {
		h.steps[sessionID][i] = services.SetupStep{
			Name:   name,
			Status: "pending",
		}
	}
	h.mu.Unlock()

	go h.consumeProgress(sessionID, ch)

	c.JSON(http.StatusOK, Success(orkaiSetupResponse{SessionID: sessionID}))
}

func (h *OrkaiSetupHandler) GetStatus(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		respondError(c, http.StatusBadRequest, ErrCodeValidation, "sessionId query parameter is required")
		return
	}

	h.mu.Lock()
	steps, ok := h.steps[sessionID]
	h.mu.Unlock()

	if !ok {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "session not found: "+sessionID)
		return
	}

	cp := make([]setupStepResponse, len(steps))
	completed := true
	for i, s := range steps {
		cp[i] = setupStepResponse{Name: s.Name, Status: s.Status, Error: s.Error}
		if s.Status == "pending" || s.Status == "in_progress" {
			completed = false
		}
	}

	c.JSON(http.StatusOK, Success(orkaiSetupStatusResponse{
		SessionID: sessionID,
		Steps:     cp,
		Completed: completed,
	}))
}

func (h *OrkaiSetupHandler) consumeProgress(sessionID string, ch <-chan services.SetupStep) {
	idx := 0
	for step := range ch {
		h.mu.Lock()
		h.steps[sessionID][idx] = services.SetupStep{
			Name:   step.Name,
			Status: step.Status,
			Error:  step.Error,
		}
		h.mu.Unlock()
		idx++
	}
	for i := idx; i < len(stepNames); i++ {
		h.mu.Lock()
		h.steps[sessionID][i] = services.SetupStep{
			Name:   stepNames[i],
			Status: "skipped",
		}
		h.mu.Unlock()
	}
}

var stepNames = []string{
	"Project Name selection + uniqueness validation",
	"Create workspace category",
	"Create Canonical Profile standard",
	"Create Cover Letter Principles standard",
	"Create PDF Pipeline standard",
	"Create PDF Generation skill",
	"Link entities",
	"Verify MCP token",
}
