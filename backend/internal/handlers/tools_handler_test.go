package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/services"
)

type fakeToolsService struct {
	shellResult  services.ShellResult
	shellErr     error
	searchResult string
	searchErr    error
	profile      any
	profileErr   error
	shellCmd     string
	shellLang    string
	searchQuery  string
}

func (f *fakeToolsService) Shell(ctx context.Context, command, language string) (services.ShellResult, error) {
	f.shellCmd = command
	f.shellLang = language
	return f.shellResult, f.shellErr
}

func (f *fakeToolsService) OrkaiSearch(ctx context.Context, query string) (string, error) {
	f.searchQuery = query
	return f.searchResult, f.searchErr
}

func (f *fakeToolsService) Profile(ctx context.Context) (any, error) {
	return f.profile, f.profileErr
}

func TestToolsHandlerShell(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	t.Run("executes shell command and returns envelope", func(t *testing.T) {
		t.Parallel()
		svc := &fakeToolsService{shellResult: services.ShellResult{Stdout: "hello\n", Stderr: "", ExitCode: 0}}
		handler := NewToolsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/tools/shell", strings.NewReader(`{"command":"echo hello","language":"bash"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Shell(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"stdout":"hello\n"`) {
			t.Errorf("expected stdout in response, got %q", w.Body.String())
		}
		if svc.shellCmd != "echo hello" {
			t.Errorf("expected command passed through, got %q", svc.shellCmd)
		}
		if svc.shellLang != "bash" {
			t.Errorf("expected language passed through, got %q", svc.shellLang)
		}
	})

	t.Run("returns 400 on missing command", func(t *testing.T) {
		t.Parallel()
		handler := NewToolsHandler(&fakeToolsService{})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/tools/shell", strings.NewReader(`{"language":"bash"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Shell(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("returns 500 on executor error", func(t *testing.T) {
		t.Parallel()
		handler := NewToolsHandler(&fakeToolsService{shellErr: errors.New("boom")})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/tools/shell", strings.NewReader(`{"command":"bad"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Shell(c)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestToolsHandlerOrkaiSearch(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	t.Run("searches and returns results", func(t *testing.T) {
		t.Parallel()
		svc := &fakeToolsService{searchResult: "cover letter principles..."}
		handler := NewToolsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/tools/orkai-search", strings.NewReader(`{"query":"cover letter"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.OrkaiSearch(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"results":"cover letter principles..."`) {
			t.Errorf("expected results in response, got %q", w.Body.String())
		}
		if svc.searchQuery != "cover letter" {
			t.Errorf("expected query passed through, got %q", svc.searchQuery)
		}
	})

	t.Run("returns 400 on missing query", func(t *testing.T) {
		t.Parallel()
		handler := NewToolsHandler(&fakeToolsService{})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/api/tools/orkai-search", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.OrkaiSearch(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})
}

func TestToolsHandlerProfile(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	t.Run("returns profile envelope", func(t *testing.T) {
		t.Parallel()
		svc := &fakeToolsService{profile: map[string]string{"fullName": "Marco"}}
		handler := NewToolsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/api/tools/profile", nil)

		handler.Profile(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"fullName":"Marco"`) {
			t.Errorf("expected fullName in response, got %q", w.Body.String())
		}
	})

	t.Run("returns null data when profile not found", func(t *testing.T) {
		t.Parallel()
		svc := &fakeToolsService{profileErr: services.ErrProfileNotFound}
		handler := NewToolsHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/v1/api/tools/profile", nil)

		handler.Profile(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 (null data for not-found profile), got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), `"data":null`) {
			t.Errorf("expected null data in response, got %q", w.Body.String())
		}
	})
}
