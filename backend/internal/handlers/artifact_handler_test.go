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

type fakeArtifactService struct {
	listFn   func(ctx context.Context) ([]models.Artifact, error)
	getFn    func(ctx context.Context, id string) (models.Artifact, error)
	createFn func(ctx context.Context, a models.Artifact) (models.Artifact, error)
	deleteFn func(ctx context.Context, id string) error
}

func (f *fakeArtifactService) List(ctx context.Context) ([]models.Artifact, error) {
	return f.listFn(ctx)
}

func (f *fakeArtifactService) Get(ctx context.Context, id string) (models.Artifact, error) {
	return f.getFn(ctx, id)
}

func (f *fakeArtifactService) Create(ctx context.Context, a models.Artifact) (models.Artifact, error) {
	return f.createFn(ctx, a)
}

func (f *fakeArtifactService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func TestArtifact_List_ReturnsArtifacts(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		listFn: func(ctx context.Context) ([]models.Artifact, error) {
			return []models.Artifact{
				{ID: "a1", Name: "Script 1", Type: "prompt", Description: "desc"},
				{ID: "a2", Name: "Script 2", Type: "template", Description: "desc 2"},
			}, nil
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.GET("/v1/api/tools/artifacts", h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/tools/artifacts", nil)
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

	data, ok := env.Data.([]interface{})
	if !ok {
		t.Fatalf("data is not an array: %T", env.Data)
	}
	if len(data) != 2 {
		t.Fatalf("len(data) = %d, want 2", len(data))
	}
}

func TestArtifact_List_Empty(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		listFn: func(ctx context.Context) ([]models.Artifact, error) {
			return []models.Artifact{}, nil
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.GET("/v1/api/tools/artifacts", h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/tools/artifacts", nil)
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
}

func TestArtifact_Get_ReturnsArtifact(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		getFn: func(ctx context.Context, id string) (models.Artifact, error) {
			return models.Artifact{
				ID:            id,
				Name:          "Test Script",
				Type:          "prompt",
				Description:   "A test artifact",
				ScriptContent: "print('hello')",
			}, nil
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.GET("/v1/api/tools/artifacts/:id", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/tools/artifacts/a1", nil)
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
	if data["id"] != "a1" {
		t.Errorf("id = %v, want a1", data["id"])
	}
	if data["name"] != "Test Script" {
		t.Errorf("name = %v, want Test Script", data["name"])
	}
}

func TestArtifact_Get_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		getFn: func(ctx context.Context, id string) (models.Artifact, error) {
			return models.Artifact{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.GET("/v1/api/tools/artifacts/:id", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/tools/artifacts/missing", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestArtifact_Create_Success(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		createFn: func(ctx context.Context, a models.Artifact) (models.Artifact, error) {
			a.ID = "new-id"
			return a, nil
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.POST("/v1/api/tools/artifacts", h.Create)

	body := `{"name":"My Script","type":"prompt","description":"test","scriptContent":"print(1)"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/tools/artifacts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Fatalf("status = %d, want 201", rec.Code)
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
	if data["id"] != "new-id" {
		t.Errorf("id = %v, want new-id", data["id"])
	}
}

func TestArtifact_Create_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{}
	router := gin.New()
	h := NewArtifactHandler(svc)
	router.POST("/v1/api/tools/artifacts", h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/tools/artifacts", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestArtifact_Delete_Success(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.DELETE("/v1/api/tools/artifacts/:id", h.Delete)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/api/tools/artifacts/a1", nil)
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
}

func TestArtifact_Delete_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeArtifactService{
		deleteFn: func(ctx context.Context, id string) error {
			return store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewArtifactHandler(svc)
	router.DELETE("/v1/api/tools/artifacts/:id", h.Delete)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/api/tools/artifacts/missing", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
