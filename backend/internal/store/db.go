package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("store.InitDB: cannot determine home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".orkai-resume", "data.db")
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("store.InitDB: cannot create directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store.InitDB: open: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.InitDB: WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.InitDB: busy_timeout: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.InitDB: ping: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("store.InitDB: migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id TEXT PRIMARY KEY,
			full_name TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			linkedin_url TEXT NOT NULL DEFAULT '',
			website_url TEXT NOT NULL DEFAULT '',
			github_url TEXT NOT NULL DEFAULT '',
			professional_summary TEXT NOT NULL DEFAULT '',
			work_experience TEXT NOT NULL DEFAULT '[]',
			education TEXT NOT NULL DEFAULT '[]',
			skills TEXT NOT NULL DEFAULT '[]',
			projects TEXT NOT NULL DEFAULT '[]',
			certifications TEXT NOT NULL DEFAULT '[]',
			languages TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS opportunities (
			id TEXT PRIMARY KEY,
			company TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'active',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS resumes (
			id TEXT PRIMARY KEY,
			opportunity_id TEXT NOT NULL DEFAULT '',
			markdown_content TEXT NOT NULL DEFAULT '',
			pdf_path TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS cover_letters (
			id TEXT PRIMARY KEY,
			opportunity_id TEXT NOT NULL DEFAULT '',
			markdown_content TEXT NOT NULL DEFAULT '',
			pdf_path TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS artifacts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			script_content TEXT NOT NULL DEFAULT '',
			usage_count INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			last_used_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS user_settings (
			id TEXT PRIMARY KEY,
			onboarding_completed INTEGER NOT NULL DEFAULT 0,
			llm_provider TEXT NOT NULL DEFAULT '',
			llm_model TEXT NOT NULL DEFAULT '',
			llm_api_key TEXT NOT NULL DEFAULT '',
			orkai_profile_category_id TEXT NOT NULL DEFAULT '',
			orkai_standards_category_id TEXT NOT NULL DEFAULT '',
			orkai_skills_category_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("store.migrate: %w", err)
		}
	}

	return nil
}
