// Package config loads resume-app configuration from environment variables.
//
// Required variables produce an actionable error when missing (no silent
// fallback, per user preference P7). Optional variables are read if present
// and consumed by the packages that need them.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all resume-app configuration loaded from the environment.
type Config struct {
	// Port is the TCP port the Gin server listens on (required).
	Port string
	// DBPath is the filesystem path to the SQLite database file.
	DBPath string
	// LLMProvider selects the LLM backend (e.g. "ollama", "openai").
	LLMProvider string
	// LLMModel names the model to call on the selected provider.
	LLMModel string
	// LLMAPIKey is the API key for hosted LLM providers (empty for local).
	LLMAPIKey string
	// OutputDir is the directory where generated PDFs are written.
	OutputDir string
	// CORSAllowedOrigins is a comma-separated list of allowed CORS origins.
	CORSAllowedOrigins string
	// OrkaiHealthURL is the URL of the orkai daemon health endpoint (required).
	OrkaiHealthURL string
	// OrkaiMCPURL is the SSE endpoint of the orkai MCP server (default http://127.0.0.1:8787/v2/sse).
	OrkaiMCPURL string
	// OrkaiMCPToken is the auth token for the orkai MCP API (required for onboarding).
	OrkaiMCPToken string
}

// Load reads environment variables and returns a Config.
//
// BACKEND_PORT is required and must be a positive integer; any other
// variable is optional at this stage and consumed by later packages.
// A missing required variable returns an actionable error rather than a
// silent default.
func Load() (Config, error) {
	_ = godotenv.Load()

	port, ok := os.LookupEnv("BACKEND_PORT")
	if !ok || port == "" {
		return Config{}, fmt.Errorf("config.Load: BACKEND_PORT is required (set it to the port the server should listen on, e.g. 8080)")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("config.Load: BACKEND_PORT must be a positive integer, got %q", port)
	}

	orkaiHealthURL, ok := os.LookupEnv("ORKAI_HEALTH_URL")
	if !ok || orkaiHealthURL == "" {
		return Config{}, fmt.Errorf("config.Load: ORKAI_HEALTH_URL is required (set it to the orkai daemon health URL, e.g. http://localhost:18787/health)")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("config.Load: cannot determine home directory for DB_PATH default: %w", err)
		}
		dbPath = filepath.Join(home, ".orkai-resume", "data.db")
	}

	orkaiMCPURL := os.Getenv("ORKAI_MCP_URL")
	if orkaiMCPURL == "" {
		orkaiMCPURL = "http://localhost:18787/mcp"
	}

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("config.Load: cannot determine home directory for OUTPUT_DIR default: %w", err)
		}
		outputDir = filepath.Join(home, ".orkai-resume", "pdfs")
	}

	return Config{
		Port:               port,
		DBPath:             dbPath,
		LLMProvider:        os.Getenv("LLM_PROVIDER"),
		LLMModel:           os.Getenv("LLM_MODEL"),
		LLMAPIKey:          os.Getenv("LLM_API_KEY"),
		OutputDir:          outputDir,
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
		OrkaiHealthURL:     orkaiHealthURL,
		OrkaiMCPURL:        orkaiMCPURL,
		OrkaiMCPToken:      os.Getenv("ORKAI_MCP_TOKEN"),
	}, nil
}
