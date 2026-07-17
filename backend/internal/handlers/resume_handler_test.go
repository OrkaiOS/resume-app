package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type fakeResumeService struct {
	getByOpportunityFn func(ctx context.Context, opportunityID string) (models.Resume, error)
	upsertFn           func(ctx context.Context, r models.Resume) (models.Resume, error)
}

func (f *fakeResumeService) GetByOpportunity(ctx context.Context, opportunityID string) (models.Resume, error) {
	return f.getByOpportunityFn(ctx, opportunityID)
}

func (f *fakeResumeService) Upsert(ctx context.Context, r models.Resume) (models.Resume, error) {
	return f.upsertFn(ctx, r)
}

func TestResume_GetByOpportunity_ReturnsResume(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{
				ID:              "r1",
				OpportunityID:   opportunityID,
				MarkdownContent: "## Summary\nExperienced engineer",
				Status:          "draft",
			}, nil
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.GET("/v1/api/opportunities/:id/resume", h.GetByOpportunity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/resume", nil)
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
	if data["id"] != "r1" {
		t.Errorf("id = %v, want r1", data["id"])
	}
	if data["status"] != "draft" {
		t.Errorf("status = %v, want draft", data["status"])
	}
}

func TestResume_GetByOpportunity_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.GET("/v1/api/opportunities/:id/resume", h.GetByOpportunity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/resume", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Data != nil {
		t.Fatalf("data = %+v, want nil", env.Data)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}
}

func TestResume_Upsert_CreatesResume(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{}, store.ErrNotFound
		},
		upsertFn: func(ctx context.Context, r models.Resume) (models.Resume, error) {
			r.ID = "new-id"
			return r, nil
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.PUT("/v1/api/opportunities/:id/resume", h.Upsert)

	body := `{"markdownContent":"## Summary\nNew resume"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/resume", strings.NewReader(body))
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

func TestResume_Upsert_UpdatesExisting(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{
				ID:              "r1",
				OpportunityID:   opportunityID,
				MarkdownContent: "old content",
				Status:          "draft",
			}, nil
		},
		upsertFn: func(ctx context.Context, r models.Resume) (models.Resume, error) {
			return r, nil
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.PUT("/v1/api/opportunities/:id/resume", h.Upsert)

	body := `{"markdownContent":"updated content"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/resume", strings.NewReader(body))
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
	if data["markdownContent"] != "updated content" {
		t.Errorf("markdownContent = %v, want 'updated content'", data["markdownContent"])
	}
}

func TestResume_Upsert_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{}
	router := gin.New()
	h := NewResumeHandler(svc)
	router.PUT("/v1/api/opportunities/:id/resume", h.Upsert)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/api/opportunities/o1/resume", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestResume_ResponseEnvelope(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{
				ID:              "r1",
				OpportunityID:   opportunityID,
				MarkdownContent: "# Resume",
				Status:          "draft",
			}, nil
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.GET("/v1/api/opportunities/:id/resume", h.GetByOpportunity)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/resume", nil)
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

func TestResume_DownloadPDF_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "Acme-Engineer-Resume.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-test"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{
				ID:              "r1",
				OpportunityID:   opportunityID,
				PDFPath:         pdfPath,
				MarkdownContent: "# Resume",
				Status:          "approved",
			}, nil
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.GET("/v1/api/opportunities/:id/resume/pdf", h.DownloadPDF)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/resume/pdf", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "inline") {
		t.Errorf("Content-Disposition = %q, want inline", cd)
	}
	if !strings.Contains(cd, "Acme-Engineer-Resume.pdf") {
		t.Errorf("Content-Disposition filename missing, got %q", cd)
	}
}

func TestResume_DownloadPDF_NotFound(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{}, store.ErrNotFound
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.GET("/v1/api/opportunities/:id/resume/pdf", h.DownloadPDF)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/resume/pdf", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestResume_DownloadPDF_NoPDFPath(t *testing.T) {
	t.Parallel()

	svc := &fakeResumeService{
		getByOpportunityFn: func(ctx context.Context, opportunityID string) (models.Resume, error) {
			return models.Resume{
				ID:              "r1",
				OpportunityID:   opportunityID,
				MarkdownContent: "# Resume",
				Status:          "draft",
			}, nil
		},
	}

	router := gin.New()
	h := NewResumeHandler(svc)
	router.GET("/v1/api/opportunities/:id/resume/pdf", h.DownloadPDF)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/opportunities/o1/resume/pdf", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
