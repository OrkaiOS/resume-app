package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/handlers"
)

func init() {
	gin.SetMode(gin.TestMode)
	hasFrontendFS = false
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != `{"status":"ok"}` {
		t.Errorf("body = %q, want {\"status\":\"ok\"}", body)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	t.Parallel()

	router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "# HELP") {
		t.Errorf("body does not look like Prometheus format (missing # HELP):\n%s", body)
	}
}

func TestV1APIGroupReturns404(t *testing.T) {
	t.Parallel()

	router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/api", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestNoSPAInDevMode(t *testing.T) {
	t.Parallel()

	router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for / in dev mode (no SPA)", w.Code)
	}
}

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()

	router := gin.New()

	healthHandler := handlers.NewHealthHandler()
	metricsHandler := handlers.NewMetricsHandler()

	router.GET("/health", healthHandler.Health)
	router.GET("/metrics", metricsHandler.Metrics)

	router.Group("/v1/api")

	return router
}
