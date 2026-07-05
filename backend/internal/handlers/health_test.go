package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealth_ReturnsOK(t *testing.T) {
	t.Parallel()

	router := gin.New()
	h := NewHealthHandler()
	router.GET("/health", h.Health)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status body = %q, want \"ok\"", body["status"])
	}
}

func TestMetrics_ReturnsPrometheusFormat(t *testing.T) {
	t.Parallel()

	router := gin.New()
	h := NewMetricsHandler()
	router.GET("/metrics", h.Metrics)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain; version=0.0.4; charset=utf-8" && ct != "text/plain; version=0.0.4; charset=utf-8; escaping=values" {
		// Tolerate minor Content-Type variations across promhttp versions.
		if !contains(ct, "text/plain") {
			t.Fatalf("Content-Type = %q, want Prometheus text/plain", ct)
		}
	}
	// promhttp with the default registry always emits at least one # HELP line.
	if !contains(rec.Body.String(), "# HELP") {
		t.Fatalf("body does not look like Prometheus format (missing # HELP):\n%s", rec.Body.String())
	}
}

func TestEnvelope_SuccessAndFailure(t *testing.T) {
	t.Parallel()

	s := Success("payload")
	if s.Data != "payload" || s.Error != nil {
		t.Fatalf("Success = %+v, want Data=payload Error=nil", s)
	}

	f := Failure("VALIDATION_ERROR", "missing field")
	if f.Data != nil || f.Error == nil || f.Error.Code != "VALIDATION_ERROR" || f.Error.Message != "missing field" {
		t.Fatalf("Failure = %+v, want error envelope", f)
	}
	if f.Error.Details == nil {
		t.Fatalf("Failure.Details = nil, want non-nil (empty object per API Contract Standard)")
	}

	d := FailureWithDetails("VALIDATION_ERROR", "bad input", map[string]string{"field": "email", "issue": "invalid format"})
	if d.Error == nil || d.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("FailureWithDetails = %+v, want error envelope", d)
	}
	details, ok := d.Error.Details.(map[string]string)
	if !ok || details["field"] != "email" {
		t.Fatalf("FailureWithDetails.Details = %+v, want field-level map", d.Error.Details)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
