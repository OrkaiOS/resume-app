package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecovery_Returns500Envelope(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Recovery())
	router.GET("/boom", func(c *gin.Context) {
		panic("kaboom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var env struct {
		Data any `json:"data"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not valid json: %v (body=%s)", err, rec.Body.String())
	}
	if env.Data != nil {
		t.Errorf("Data = %v, want nil", env.Data)
	}
	if env.Error == nil || env.Error.Code != "INTERNAL" {
		t.Errorf("Error = %+v, want code INTERNAL", env.Error)
	}
}

func TestMetrics_HandlerIncrementsCounters(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	m := NewMetrics()
	router := gin.New()
	router.Use(m.Handler())
	router.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	// Collect the metrics and assert the request counter was incremented.
	families, err := gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	total, ok := families["http_requests_total"]
	if !ok {
		t.Fatalf("http_requests_total not registered; got %v", familyNames(families))
	}
	var sawPing bool
	for _, m := range total.Metric {
		for _, l := range m.Label {
			if l.GetName() == "route" && l.GetValue() == "/ping" {
				sawPing = true
			}
		}
	}
	if !sawPing {
		t.Fatalf("http_requests_total has no /ping label; metrics=%v", total.Metric)
	}
}
