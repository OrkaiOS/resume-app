package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
)

type mockParser struct {
	pdfResult *models.Profile
	pdfErr    error
	mdResult  *models.Profile
	mdErr     error
}

func (m *mockParser) ParsePDF(r io.Reader) (*models.Profile, error) {
	if _, err := io.ReadAll(r); err != nil {
		return nil, err
	}
	return m.pdfResult, m.pdfErr
}

func (m *mockParser) ParseMarkdown(r io.Reader) (*models.Profile, error) {
	if _, err := io.ReadAll(r); err != nil {
		return nil, err
	}
	return m.mdResult, m.mdErr
}

func newMultipartBody(t *testing.T, fieldName, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	w.Close()
	return &buf, w.FormDataContentType()
}

func TestProfileUpload_PDF(t *testing.T) {
	t.Parallel()

	parser := &mockParser{
		pdfResult: &models.Profile{
			FullName:            "Jane Doe",
			ProfessionalSummary: "Experienced engineer",
		},
	}
	handler := NewProfileUploadHandler(parser)

	body, ct := newMultipartBody(t, "file", "resume.pdf", "fake pdf content")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/profile/upload", body)
	req.Header.Set("Content-Type", ct)

	router := gin.New()
	router.POST("/v1/api/profile/upload", handler.Upload)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
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
	if data["professionalSummary"] != "Experienced engineer" {
		t.Errorf("professionalSummary = %v, want Experienced engineer", data["professionalSummary"])
	}
}

func TestProfileUpload_Markdown(t *testing.T) {
	t.Parallel()

	parser := &mockParser{
		mdResult: &models.Profile{
			FullName: "John Smith",
			Email:    "john@example.com",
		},
	}
	handler := NewProfileUploadHandler(parser)

	body, ct := newMultipartBody(t, "file", "resume.md", "# John Smith\n\n## Contact\n- Email: john@example.com")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/profile/upload", body)
	req.Header.Set("Content-Type", ct)

	router := gin.New()
	router.POST("/v1/api/profile/upload", handler.Upload)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
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
	if data["email"] != "john@example.com" {
		t.Errorf("email = %v, want john@example.com", data["email"])
	}
}

func TestProfileUpload_MissingFile(t *testing.T) {
	t.Parallel()

	parser := &mockParser{}
	handler := NewProfileUploadHandler(parser)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	w.WriteField("other", "value")
	w.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/profile/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	router := gin.New()
	router.POST("/v1/api/profile/upload", handler.Upload)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if env.Error.Code != ErrCodeValidation {
		t.Errorf("error code = %s, want %s", env.Error.Code, ErrCodeValidation)
	}
}

func TestProfileUpload_UnsupportedFileType(t *testing.T) {
	t.Parallel()

	parser := &mockParser{}
	handler := NewProfileUploadHandler(parser)

	body, ct := newMultipartBody(t, "file", "resume.docx", "fake docx content")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/profile/upload", body)
	req.Header.Set("Content-Type", ct)

	router := gin.New()
	router.POST("/v1/api/profile/upload", handler.Upload)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if env.Error.Code != ErrCodeValidation {
		t.Errorf("error code = %s, want %s", env.Error.Code, ErrCodeValidation)
	}
}

func TestProfileUpload_ParserError(t *testing.T) {
	t.Parallel()

	parser := &mockParser{
		pdfErr: io.ErrUnexpectedEOF,
	}
	handler := NewProfileUploadHandler(parser)

	body, ct := newMultipartBody(t, "file", "resume.pdf", "fake pdf content")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/profile/upload", body)
	req.Header.Set("Content-Type", ct)

	router := gin.New()
	router.POST("/v1/api/profile/upload", handler.Upload)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if env.Error.Code != ErrCodeInternal {
		t.Errorf("error code = %s, want %s", env.Error.Code, ErrCodeInternal)
	}
}

func TestProfileUpload_FileTooLarge(t *testing.T) {
	t.Parallel()

	parser := &mockParser{}
	handler := NewProfileUploadHandler(parser)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "large.pdf")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	chunk := bytes.Repeat([]byte("x"), 1024*1024)
	for i := 0; i < 11; i++ {
		if _, err := part.Write(chunk); err != nil {
			t.Fatalf("write chunk: %v", err)
		}
	}
	w.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/profile/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	router := gin.New()
	router.POST("/v1/api/profile/upload", handler.Upload)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d for too-large file, want 400: %s", rec.Code, rec.Body.String())
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if env.Error.Code != ErrCodeValidation {
		t.Errorf("error code = %s, want %s", env.Error.Code, ErrCodeValidation)
	}
}
