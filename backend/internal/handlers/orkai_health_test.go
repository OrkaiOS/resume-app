package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestOrkaiHealth_Running(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	router := gin.New()
	h := NewOrkaiHealthHandler(srv.URL)
	router.GET("/v1/api/health/orkai", h.CheckHealth)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/health/orkai", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}
	if env.Data == nil {
		t.Fatal("data is nil")
	}

	got, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	running, ok := got["running"].(bool)
	if !ok || !running {
		t.Fatalf("running = %v, want true", got["running"])
	}
}

func TestOrkaiHealth_NotRunning(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	router := gin.New()
	h := NewOrkaiHealthHandler(srv.URL)
	router.GET("/v1/api/health/orkai", h.CheckHealth)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/health/orkai", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}
	if env.Data == nil {
		t.Fatal("data is nil")
	}

	got, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	running, ok := got["running"].(bool)
	if !ok || running {
		t.Fatalf("running = %v, want false", got["running"])
	}
}

func TestOrkaiHealth_Unreachable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	url := srv.URL
	srv.Close()

	router := gin.New()
	h := NewOrkaiHealthHandler(url)
	router.GET("/v1/api/health/orkai", h.CheckHealth)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/health/orkai", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}
	if env.Data == nil {
		t.Fatal("data is nil")
	}

	got, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	running, ok := got["running"].(bool)
	if !ok || running {
		t.Fatalf("running = %v, want false", got["running"])
	}
}

func TestOrkaiHealth_Timeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer srv.Close()

	router := gin.New()
	h := NewOrkaiHealthHandler(srv.URL)
	router.GET("/v1/api/health/orkai", h.CheckHealth)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/health/orkai", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not valid json: %v", err)
	}
	if env.Data == nil {
		t.Fatal("data is nil")
	}

	got, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	running, ok := got["running"].(bool)
	if !ok || running {
		t.Fatalf("running = %v, want false (timeout)", got["running"])
	}
}

func TestOrkaiHealth_Non200StatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
	}{
		{"404", http.StatusNotFound},
		{"500", http.StatusInternalServerError},
		{"302", http.StatusFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			router := gin.New()
			h := NewOrkaiHealthHandler(srv.URL)
			router.GET("/v1/api/health/orkai", h.CheckHealth)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/api/health/orkai", nil)
			router.ServeHTTP(rec, req)

			if rec.Code != 200 {
				t.Fatalf("status = %d, want 200", rec.Code)
			}

			var env Envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("body is not valid json: %v", err)
			}
			got, ok := env.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("data is not an object: %T", env.Data)
			}
			running, ok := got["running"].(bool)
			if !ok || running {
				t.Fatalf("running = %v, want false for status %d", got["running"], tt.status)
			}
		})
	}
}

func TestOrkaiHealth_ResponseEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	router := gin.New()
	h := NewOrkaiHealthHandler(srv.URL)
	router.GET("/v1/api/health/orkai", h.CheckHealth)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/health/orkai", nil)
	router.ServeHTTP(rec, req)

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}
	if !strings.Contains(rec.Body.String(), `"running"`) {
		t.Fatalf("response missing running field: %s", rec.Body.String())
	}
}
