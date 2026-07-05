// Package config loads resume-app configuration from environment variables.
//
// Required variables produce an actionable error when missing (no silent
// fallback, per user preference P7). Optional variables are read if present
// and consumed by the packages that need them.
package config

import (
	"fmt"
	"os"
	"strconv"
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
}

// Load reads environment variables and returns a Config.
//
// BACKEND_PORT is required and must be a positive integer; any other
// variable is optional at this stage and consumed by later packages.
// A missing required variable returns an actionable error rather than a
// silent default.
func Load() (Config, error) {
	port, ok := os.LookupEnv("BACKEND_PORT")
	if !ok || port == "" {
		return Config{}, fmt.Errorf("config.Load: BACKEND_PORT is required (set it to the port the server should listen on, e.g. 8080)")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return Config{}, fmt.Errorf("config.Load: BACKEND_PORT must be a positive integer, got %q", port)
	}

	return Config{
		Port:               port,
		DBPath:             os.Getenv("DB_PATH"),
		LLMProvider:        os.Getenv("LLM_PROVIDER"),
		LLMModel:           os.Getenv("LLM_MODEL"),
		LLMAPIKey:          os.Getenv("LLM_API_KEY"),
		OutputDir:          os.Getenv("OUTPUT_DIR"),
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
	}, nil
}
