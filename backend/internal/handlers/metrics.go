package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler serves the built-in /metrics endpoint in Prometheus format.
type MetricsHandler struct {
	handler http.Handler
}

// NewMetricsHandler constructs a MetricsHandler backed by the default
// Prometheus registry (which the middleware registers collectors on).
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{handler: promhttp.Handler()}
}

// Metrics exposes Prometheus-format metrics. No auth.
func (h *MetricsHandler) Metrics(c *gin.Context) {
	h.handler.ServeHTTP(c.Writer, c.Request)
}
