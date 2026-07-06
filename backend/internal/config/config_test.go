package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_RequiresBackendPort(t *testing.T) {

	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "missing BACKEND_PORT",
			env:  map[string]string{},
			want: "BACKEND_PORT is required",
		},
		{
			name: "empty BACKEND_PORT",
			env:  map[string]string{"BACKEND_PORT": ""},
			want: "BACKEND_PORT is required",
		},
		{
			name: "non-numeric BACKEND_PORT",
			env:  map[string]string{"BACKEND_PORT": "abc"},
			want: "BACKEND_PORT must be a positive integer",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("BACKEND_PORT", tc.env["BACKEND_PORT"])
			for k, v := range tc.env {
				if k == "BACKEND_PORT" {
					continue
				}
				t.Setenv(k, v)
			}
			// Ensure no leftover value from a previous subtest.
			if _, set := tc.env["BACKEND_PORT"]; !set {
				t.Setenv("BACKEND_PORT", "")
			}

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() expected error containing %q, got nil", tc.want)
			}
			if !contains(err.Error(), tc.want) {
				t.Fatalf("Load() error = %q, want substring %q", err.Error(), tc.want)
			}
		})
	}
}

func TestLoad_SuccessReadsAllVars(t *testing.T) {

	t.Setenv("BACKEND_PORT", "8080")
	t.Setenv("DB_PATH", "/tmp/resume.db")
	t.Setenv("LLM_PROVIDER", "ollama")
	t.Setenv("LLM_MODEL", "llama3")
	t.Setenv("LLM_API_KEY", "secret")
	t.Setenv("OUTPUT_DIR", "/tmp/out")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	t.Setenv("ORKAI_HEALTH_URL", "http://localhost:18787/health")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.DBPath != "/tmp/resume.db" {
		t.Errorf("DBPath = %q, want /tmp/resume.db", cfg.DBPath)
	}
	if cfg.LLMProvider != "ollama" {
		t.Errorf("LLMProvider = %q, want ollama", cfg.LLMProvider)
	}
	if cfg.LLMModel != "llama3" {
		t.Errorf("LLMModel = %q, want llama3", cfg.LLMModel)
	}
	if cfg.LLMAPIKey != "secret" {
		t.Errorf("LLMAPIKey = %q, want secret", cfg.LLMAPIKey)
	}
	if cfg.OutputDir != "/tmp/out" {
		t.Errorf("OutputDir = %q, want /tmp/out", cfg.OutputDir)
	}
	if cfg.CORSAllowedOrigins != "http://localhost:5173" {
		t.Errorf("CORSAllowedOrigins = %q, want http://localhost:5173", cfg.CORSAllowedOrigins)
	}
	if cfg.OrkaiHealthURL != "http://localhost:18787/health" {
		t.Errorf("OrkaiHealthURL = %q, want http://localhost:18787/health", cfg.OrkaiHealthURL)
	}
}

func TestLoad_DBPathDefault(t *testing.T) {
	t.Setenv("BACKEND_PORT", "8080")
	t.Setenv("ORKAI_HEALTH_URL", "http://localhost:18787/health")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	expected := filepath.Join(home, ".orkai-resume", "data.db")
	if cfg.DBPath != expected {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, expected)
	}
}

func TestLoad_RequiresOrkaiHealthURL(t *testing.T) {
	t.Setenv("BACKEND_PORT", "8080")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() expected error for missing ORKAI_HEALTH_URL, got nil")
	}
	if !contains(err.Error(), "ORKAI_HEALTH_URL is required") {
		t.Fatalf("Load() error = %q, want substring %q", err.Error(), "ORKAI_HEALTH_URL is required")
	}

	t.Setenv("ORKAI_HEALTH_URL", "")
	_, err = Load()
	if err == nil {
		t.Fatal("Load() expected error for empty ORKAI_HEALTH_URL, got nil")
	}
	if !contains(err.Error(), "ORKAI_HEALTH_URL is required") {
		t.Fatalf("Load() error = %q, want substring %q", err.Error(), "ORKAI_HEALTH_URL is required")
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && stringsContains(haystack, needle))
}

func stringsContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
