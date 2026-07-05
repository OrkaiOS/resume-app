package handlers

import "github.com/gin-gonic/gin"

// HealthHandler serves the built-in /health endpoint.
type HealthHandler struct{}

// NewHealthHandler constructs a HealthHandler. It has no dependencies.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns 200 with {"status":"ok"}. No auth.
//
// /health is intentionally NOT wrapped in the standard envelope: it is a
// load-balancer probe with a fixed contract.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
