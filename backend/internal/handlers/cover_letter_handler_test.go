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

type fakeCoverLetterService struct {
	getByOpportunityFn func(ctx context.Context, opportunityID string) (models.CoverLetter, error)
	upsertFn           func(ctx context.Context, cl models.CoverLetter) (models.CoverLetter, error)
}

func (f *fakeCoverLetterService) GetByOpportunity(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
	return f.getByOpportunityFn(ctx, opportunityID)
}

func (f *fakeCoverLetterService) Upsert(ctx context.Context, cl models.CoverLetter) (models.CoverLetter, error) {
	return f.upsertFn(ctx, cl)
}

func TestCoverLetter_GetByOpportunity_ReturnsCoverLetter(t *testing.T) {
	t.Parallel()

	svc := &fakeCoverLetterService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
			return models.CoverLetter{
				ID:              "cl1",
				OpportunityID:   opportunityID,
				MarkdownContent: "Dear Hiring Manager,\nI am excited to apply...",
				Status:          "draft",
			}, nil
		},
	}

	router := gin.New()
	h := NewCoverLetterHandler(svc)
	router.GET("/v1/api/opportunities/:id/cover-letter", h.GetByOpportunity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/cover-letter", nil)
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
	if data["id"] != "cl1" {
		t.Errorf("id = %v, want cl1", data["id"])
	}
	if data["status"] != "draft" {
		t.Errorf("status = %v, want draft", data["status"])
	}
}

func TestCoverLetter_GetByOpportunity_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeCoverLetterService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
			return models.CoverLetter{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewCoverLetterHandler(svc)
	router.GET("/v1/api/opportunities/:id/cover-letter", h.GetByOpportunity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/cover-letter", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestCoverLetter_Upsert_CreatesCoverLetter(t *testing.T) {
	t.Parallel()

	svc := &fakeCoverLetterService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
			return models.CoverLetter{}, store.ErrNotFound
		},
		upsertFn: func(ctx context.Context, cl models.CoverLetter) (models.CoverLetter, error) {
			cl.ID = "new-cl-id"
			return cl, nil
		},
	}

	router := gin.New()
	h := NewCoverLetterHandler(svc)
	router.PUT("/v1/api/opportunities/:id/cover-letter", h.Upsert)

	body := `{"markdownContent":"Dear Hiring Manager..."}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/cover-letter", strings.NewReader(body))
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
}

func TestCoverLetter_Upsert_UpdatesExisting(t *testing.T) {
	t.Parallel()

	svc := &fakeCoverLetterService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
			return models.CoverLetter{
				ID:              "cl1",
				OpportunityID:   opportunityID,
				MarkdownContent: "old cover letter",
				Status:          "draft",
			}, nil
		},
		upsertFn: func(ctx context.Context, cl models.CoverLetter) (models.CoverLetter, error) {
			return cl, nil
		},
	}

	router := gin.New()
	h := NewCoverLetterHandler(svc)
	router.PUT("/v1/api/opportunities/:id/cover-letter", h.Upsert)

	body := `{"markdownContent":"updated cover letter"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/cover-letter", strings.NewReader(body))
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
	if data["markdownContent"] != "updated cover letter" {
		t.Errorf("markdownContent = %v, want 'updated cover letter'", data["markdownContent"])
	}
}

func TestCoverLetter_Upsert_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := &fakeCoverLetterService{}
	router := gin.New()
	h := NewCoverLetterHandler(svc)
	router.PUT("/v1/api/opportunities/:id/cover-letter", h.Upsert)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/cover-letter", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCoverLetter_ResponseEnvelope(t *testing.T) {
	t.Parallel()

	svc := &fakeCoverLetterService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
			return models.CoverLetter{
				ID:              "cl1",
				OpportunityID:   opportunityID,
				MarkdownContent: "# Cover Letter",
				Status:          "draft",
			}, nil
		},
	}

	router := gin.New()
	h := NewCoverLetterHandler(svc)
	router.GET("/v1/api/opportunities/:id/cover-letter", h.GetByOpportunity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/cover-letter", nil)
	router.ServeHTTP(rec, req)

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}
	if !strings.Contains(rec.Body.String(), `"markdownContent"`) {
		t.Fatalf("response missing markdownContent: %s", rec.Body.String())
	}
}
