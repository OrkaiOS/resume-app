package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const orkaiHealthTimeout = 2 * time.Second

// OrkaiHealthHandler serves GET /v1/api/health/orkai — a versioned API
// endpoint (NOT the built-in /health) that probes the orkai daemon and returns
// {"data":{"running":true/false},"error":null} using the standard API envelope.
type OrkaiHealthHandler struct {
	HealthURL string
	client    *http.Client
}

func NewOrkaiHealthHandler(healthURL string) *OrkaiHealthHandler {
	return &OrkaiHealthHandler{
		HealthURL: healthURL,
		client: &http.Client{
			Timeout: orkaiHealthTimeout,
		},
	}
}

type orkaiHealthResponse struct {
	Running bool `json:"running"`
}

func (h *OrkaiHealthHandler) CheckHealth(c *gin.Context) {
	// Per API Contract Standard, versioned endpoints use the standard
	// Success() envelope. This is not the built-in /health endpoint.
	resp, err := h.client.Get(h.HealthURL)
	if err != nil {
		c.JSON(http.StatusOK, Success(orkaiHealthResponse{Running: false}))
		return
	}
	defer resp.Body.Close()

	running := resp.StatusCode == http.StatusOK
	c.JSON(http.StatusOK, Success(orkaiHealthResponse{Running: running}))
}
