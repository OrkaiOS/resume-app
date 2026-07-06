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

type fakeOpportunityService struct {
	listFn        func(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error)
	getFn         func(ctx context.Context, id string) (models.Opportunity, error)
	createFn      func(ctx context.Context, o models.Opportunity) (models.Opportunity, error)
	updateFn      func(ctx context.Context, o models.Opportunity) (models.Opportunity, error)
	deleteFn      func(ctx context.Context, id string) error
	setArchivedFn func(ctx context.Context, id string, archived bool) (models.Opportunity, error)
}

func (f *fakeOpportunityService) List(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error) {
	return f.listFn(ctx, cursor, limit)
}

func (f *fakeOpportunityService) Get(ctx context.Context, id string) (models.Opportunity, error) {
	return f.getFn(ctx, id)
}

func (f *fakeOpportunityService) Create(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
	return f.createFn(ctx, o)
}

func (f *fakeOpportunityService) Update(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
	return f.updateFn(ctx, o)
}

func (f *fakeOpportunityService) Delete(ctx context.Context, id string) error {
	return f.deleteFn(ctx, id)
}

func (f *fakeOpportunityService) SetArchived(ctx context.Context, id string, archived bool) (models.Opportunity, error) {
	return f.setArchivedFn(ctx, id, archived)
}

func TestOpportunity_List_ReturnsItems(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		listFn: func(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error) {
			return []models.Opportunity{
				{ID: "o1", Company: "Acme", Role: "Engineer", Status: "active"},
				{ID: "o2", Company: "Globex", Role: "Manager", Status: "active"},
			}, "o2", nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.GET("/v1/api/opportunities", h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities", nil)
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
	items, ok := data["items"].([]interface{})
	if !ok {
		t.Fatalf("items is not an array: %T", data["items"])
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	if data["nextCursor"] != "o2" {
		t.Errorf("nextCursor = %v, want o2", data["nextCursor"])
	}
}

func TestOpportunity_List_WithCursorAndLimit(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		listFn: func(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error) {
			if cursor != "o5" {
				t.Errorf("cursor = %q, want o5", cursor)
			}
			if limit != 5 {
				t.Errorf("limit = %d, want 5", limit)
			}
			return []models.Opportunity{}, "", nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.GET("/v1/api/opportunities", h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities?cursor=o5&limit=5", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestOpportunity_List_InvalidLimit(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{}
	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.GET("/v1/api/opportunities", h.List)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities?limit=abc", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestOpportunity_Get_ReturnsOpportunity(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		getFn: func(ctx context.Context, id string) (models.Opportunity, error) {
			return models.Opportunity{
				ID:      id,
				Company: "Acme",
				Role:    "Engineer",
				Status:  "active",
			}, nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.GET("/v1/api/opportunities/:id", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	if data["company"] != "Acme" {
		t.Errorf("company = %v, want Acme", data["company"])
	}
}

func TestOpportunity_Get_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		getFn: func(ctx context.Context, id string) (models.Opportunity, error) {
			return models.Opportunity{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.GET("/v1/api/opportunities/:id", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/nonexistent", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestOpportunity_Create_ReturnsCreated(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		createFn: func(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
			o.ID = "new-id"
			return o, nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.POST("/v1/api/opportunities", h.Create)

	body := `{"company":"Acme","role":"Engineer"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/opportunities", strings.NewReader(body))
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
}

func TestOpportunity_Create_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{}
	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.POST("/v1/api/opportunities", h.Create)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/opportunities", strings.NewReader("bad"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestOpportunity_Update_ReturnsUpdated(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		updateFn: func(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
			return o, nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.PUT("/v1/api/opportunities/:id", h.Update)

	body := `{"company":"Updated","role":"Senior"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestOpportunity_Update_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		updateFn: func(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
			return models.Opportunity{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.PUT("/v1/api/opportunities/:id", h.Update)

	body := `{"company":"X","role":"Y"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestOpportunity_Delete_ReturnsOK(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		deleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.DELETE("/v1/api/opportunities/:id", h.Delete)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/api/opportunities/o1", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestOpportunity_Delete_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		deleteFn: func(ctx context.Context, id string) error {
			return store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.DELETE("/v1/api/opportunities/:id", h.Delete)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/api/opportunities/nonexistent", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestOpportunity_SetArchived_Archives(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		setArchivedFn: func(ctx context.Context, id string, archived bool) (models.Opportunity, error) {
			return models.Opportunity{
				ID:     id,
				Status: "archived",
			}, nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.PUT("/v1/api/opportunities/:id/archive", h.SetArchived)

	body := `{"archived":true}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/archive", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	if data["status"] != "archived" {
		t.Errorf("status = %v, want archived", data["status"])
	}
}

func TestOpportunity_SetArchived_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		setArchivedFn: func(ctx context.Context, id string, archived bool) (models.Opportunity, error) {
			return models.Opportunity{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.PUT("/v1/api/opportunities/:id/archive", h.SetArchived)

	body := `{"archived":true}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/nonexistent/archive", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestOpportunity_ResponseEnvelope(t *testing.T) {
	t.Parallel()

	svc := &fakeOpportunityService{
		getFn: func(ctx context.Context, id string) (models.Opportunity, error) {
			return models.Opportunity{
				ID:      id,
				Company: "Acme",
				Role:    "Engineer",
				Status:  "active",
			}, nil
		},
	}

	router := gin.New()
	h := NewOpportunityHandler(svc)
	router.GET("/v1/api/opportunities/:id", h.Get)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1", nil)
	router.ServeHTTP(rec, req)

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}
	if !strings.Contains(rec.Body.String(), `"company"`) {
		t.Fatalf("response missing company: %s", rec.Body.String())
	}
}
