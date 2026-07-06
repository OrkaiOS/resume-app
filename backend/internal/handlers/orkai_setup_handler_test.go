package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/services"
	"github.com/marco/resume-app/internal/store"
)

type mockOrkaiSetupSvc struct {
	sessionID string
	ch        chan services.SetupStep
	err       error
}

func (m *mockOrkaiSetupSvc) RunSetup(ctx context.Context, profile models.Profile) (string, <-chan services.SetupStep, error) {
	if m.err != nil {
		return "", nil, m.err
	}
	return m.sessionID, m.ch, nil
}

type mockProfileSvc struct {
	profile models.Profile
	err     error
}

func (m *mockProfileSvc) Get(ctx context.Context) (models.Profile, error) {
	return m.profile, m.err
}

func emitSteps(ch chan services.SetupStep) {
	ch <- services.SetupStep{Name: "Create personal category", Status: "success"}
	ch <- services.SetupStep{Name: "Create Canonical Profile standard", Status: "success"}
	ch <- services.SetupStep{Name: "Create Cover Letter Principles standard", Status: "success"}
	ch <- services.SetupStep{Name: "Create PDF Pipeline standard", Status: "success"}
	ch <- services.SetupStep{Name: "Create PDF Generation skill", Status: "success"}
	ch <- services.SetupStep{Name: "Link entities", Status: "success"}
	ch <- services.SetupStep{Name: "Verify MCP token", Status: "success"}
	close(ch)
}

func TestOrkaiSetup_StartSetup_Success(t *testing.T) {
	t.Parallel()

	ch := make(chan services.SetupStep, 7)
	svc := &mockOrkaiSetupSvc{sessionID: "test-session-1", ch: ch}
	profile := &mockProfileSvc{
		profile: models.Profile{FullName: "Jane Doe", Email: "jane@example.com"},
	}

	handler := NewOrkaiSetupHandler(svc, profile)

	go emitSteps(ch)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/orkai-setup", nil)

	router := gin.New()
	router.POST("/v1/api/onboarding/orkai-setup", handler.StartSetup)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	if data["sessionId"] != "test-session-1" {
		t.Errorf("sessionId = %v, want test-session-1", data["sessionId"])
	}
}

func TestOrkaiSetup_GetStatus_Success(t *testing.T) {
	t.Parallel()

	ch := make(chan services.SetupStep, 7)
	svc := &mockOrkaiSetupSvc{sessionID: "test-session-2", ch: ch}
	profile := &mockProfileSvc{
		profile: models.Profile{FullName: "John Smith", Email: "john@example.com"},
	}

	handler := NewOrkaiSetupHandler(svc, profile)

	go emitSteps(ch)

	startRec := httptest.NewRecorder()
	startReq := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/orkai-setup", nil)
	router := gin.New()
	router.POST("/v1/api/onboarding/orkai-setup", handler.StartSetup)
	router.GET("/v1/api/onboarding/orkai-setup/status", handler.GetStatus)
	router.ServeHTTP(startRec, startReq)

	statusRec := httptest.NewRecorder()
	statusReq := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/orkai-setup/status?sessionId=test-session-2", nil)
	router.ServeHTTP(statusRec, statusReq)

	if statusRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", statusRec.Code, statusRec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(statusRec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}

	steps, ok := data["steps"].([]interface{})
	if !ok {
		t.Fatalf("steps is not an array: %T", data["steps"])
	}
	if len(steps) != 7 {
		t.Errorf("steps length = %d, want 7", len(steps))
	}
}

func TestOrkaiSetup_GetStatus_MissingSessionID(t *testing.T) {
	t.Parallel()

	svc := &mockOrkaiSetupSvc{}
	profile := &mockProfileSvc{}
	handler := NewOrkaiSetupHandler(svc, profile)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/orkai-setup/status", nil)

	router := gin.New()
	router.GET("/v1/api/onboarding/orkai-setup/status", handler.GetStatus)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil || env.Error.Code != ErrCodeValidation {
		t.Fatalf("expected validation error, got: %+v", env.Error)
	}
}

func TestOrkaiSetup_GetStatus_UnknownSession(t *testing.T) {
	t.Parallel()

	svc := &mockOrkaiSetupSvc{}
	profile := &mockProfileSvc{}
	handler := NewOrkaiSetupHandler(svc, profile)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/orkai-setup/status?sessionId=nonexistent", nil)

	router := gin.New()
	router.GET("/v1/api/onboarding/orkai-setup/status", handler.GetStatus)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil || env.Error.Code != ErrCodeNotFound {
		t.Fatalf("expected not_found error, got: %+v", env.Error)
	}
}

func TestOrkaiSetup_StartSetup_ProfileNotFound(t *testing.T) {
	t.Parallel()

	svc := &mockOrkaiSetupSvc{}
	profile := &mockProfileSvc{err: store.ErrNotFound}

	handler := NewOrkaiSetupHandler(svc, profile)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/orkai-setup", nil)

	router := gin.New()
	router.POST("/v1/api/onboarding/orkai-setup", handler.StartSetup)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil || env.Error.Code != ErrCodeNotFound {
		t.Fatalf("expected not_found error, got: %+v", env.Error)
	}
}

func TestOrkaiSetup_StartSetup_ServiceError(t *testing.T) {
	t.Parallel()

	svc := &mockOrkaiSetupSvc{
		err: context.DeadlineExceeded,
	}
	profile := &mockProfileSvc{
		profile: models.Profile{FullName: "Test User"},
	}

	handler := NewOrkaiSetupHandler(svc, profile)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/orkai-setup", nil)

	router := gin.New()
	router.POST("/v1/api/onboarding/orkai-setup", handler.StartSetup)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil || env.Error.Code != ErrCodeInternal {
		t.Fatalf("expected internal error, got: %+v", env.Error)
	}
}
