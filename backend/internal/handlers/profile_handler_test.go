package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type fakeProfileService struct {
	getFn    func(ctx context.Context) (models.Profile, error)
	upsertFn func(ctx context.Context, p models.Profile) (models.Profile, error)
}

func (f *fakeProfileService) Get(ctx context.Context) (models.Profile, error) {
	return f.getFn(ctx)
}

func (f *fakeProfileService) Upsert(ctx context.Context, p models.Profile) (models.Profile, error) {
	return f.upsertFn(ctx, p)
}

func TestProfile_Get_ReturnsProfile(t *testing.T) {
	t.Parallel()

	svc := &fakeProfileService{
		getFn: func(ctx context.Context) (models.Profile, error) {
			return models.Profile{
				ID:       "p1",
				FullName: "Jane Doe",
				Email:    "jane@example.com",
			}, nil
		},
	}

	router := gin.New()
	h := NewProfileHandler(svc)
	router.GET("/v1/api/profile", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/profile", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	if data["fullName"] != "Jane Doe" {
		t.Errorf("fullName = %v, want Jane Doe", data["fullName"])
	}
}

func TestProfile_Get_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeProfileService{
		getFn: func(ctx context.Context) (models.Profile, error) {
			return models.Profile{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewProfileHandler(svc)
	router.GET("/v1/api/profile", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/profile", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Data != nil {
		t.Fatalf("data = %v, want nil", env.Data)
	}
}

func TestProfile_Upsert_CreatesProfile(t *testing.T) {
	t.Parallel()

	svc := &fakeProfileService{
		upsertFn: func(ctx context.Context, p models.Profile) (models.Profile, error) {
			p.ID = "new-id"
			return p, nil
		},
	}

	router := gin.New()
	h := NewProfileHandler(svc)
	router.PUT("/v1/api/profile", h.Upsert)

	body := `{"fullName":"John Smith","email":"john@example.com"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	if data["fullName"] != "John Smith" {
		t.Errorf("fullName = %v, want John Smith", data["fullName"])
	}
}

func TestProfile_Upsert_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := &fakeProfileService{}
	router := gin.New()
	h := NewProfileHandler(svc)
	router.PUT("/v1/api/profile", h.Upsert)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/profile", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestProfile_ResponseEnvelope(t *testing.T) {
	t.Parallel()

	svc := &fakeProfileService{
		getFn: func(ctx context.Context) (models.Profile, error) {
			return models.Profile{
				ID:       "p1",
				FullName: "Test User",
			}, nil
		},
	}

	router := gin.New()
	h := NewProfileHandler(svc)
	router.GET("/v1/api/profile", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/profile", nil)
	router.ServeHTTP(rec, req)

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}
	if !strings.Contains(rec.Body.String(), `"fullName"`) {
		t.Fatalf("response missing fullName: %s", rec.Body.String())
	}
}
