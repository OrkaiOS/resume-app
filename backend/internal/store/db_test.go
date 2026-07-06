package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB_CreatesTables(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer db.Close()

	tables := []string{
		"profiles",
		"opportunities",
		"resumes",
		"cover_letters",
		"artifacts",
		"user_settings",
	}

	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestInitDB_Idempotent(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("first InitDB() error = %v", err)
	}
	db.Close()

	db, err = InitDB(dbPath)
	if err != nil {
		t.Fatalf("second InitDB() error = %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		t.Fatalf("count tables: %v", err)
	}
	if count != 6 {
		t.Errorf("expected 6 tables, got %d", count)
	}
}

func TestInitDB_WALMode(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer db.Close()

	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want wal", journalMode)
	}
}

func TestInitDB_CreatesParentDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nested", "subdir")
	dbPath := filepath.Join(dir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("parent directory %q was not created", dir)
	}
}

func TestInitDB_DefaultPath(t *testing.T) {
	db, err := InitDB("")
	if err != nil {
		t.Fatalf("InitDB(\"\") error = %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		t.Fatalf("count tables: %v", err)
	}
	if count != 6 {
		t.Errorf("expected 6 tables, got %d", count)
	}

	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".orkai-resume", "data.db")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected db at %q, but file does not exist", expectedPath)
	}
}

func TestInitDB_TableColumns(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer db.Close()

	tests := []struct {
		table   string
		columns []string
	}{
		{
			table: "profiles",
			columns: []string{
				"id", "full_name", "email", "phone", "location",
				"linkedin_url", "website_url", "github_url",
				"professional_summary", "work_experience", "education",
				"skills", "projects", "certifications", "languages",
				"created_at", "updated_at",
			},
		},
		{
			table: "opportunities",
			columns: []string{
				"id", "company", "role", "description", "status",
				"created_at", "updated_at",
			},
		},
		{
			table: "resumes",
			columns: []string{
				"id", "opportunity_id", "markdown_content", "pdf_path",
				"status", "created_at", "updated_at",
			},
		},
		{
			table: "cover_letters",
			columns: []string{
				"id", "opportunity_id", "markdown_content", "pdf_path",
				"status", "created_at", "updated_at",
			},
		},
		{
			table: "artifacts",
			columns: []string{
				"id", "name", "type", "description", "script_content",
				"usage_count", "created_at", "last_used_at",
			},
		},
		{
			table: "user_settings",
			columns: []string{
				"id", "onboarding_completed", "llm_provider", "llm_model",
				"llm_api_key", "orkai_profile_category_id",
				"orkai_standards_category_id", "orkai_skills_category_id",
				"created_at", "updated_at",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.table, func(t *testing.T) {
			rows, err := db.Query("PRAGMA table_info(" + tt.table + ")")
			if err != nil {
				t.Fatalf("PRAGMA table_info(%s): %v", tt.table, err)
			}
			defer rows.Close()

			got := make(map[string]bool)
			for rows.Next() {
				var cid int
				var name, colType string
				var notNull int
				var defaultVal *string
				var pk int
				if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
					t.Fatalf("scan: %v", err)
				}
				got[name] = true
			}

			for _, col := range tt.columns {
				if !got[col] {
					t.Errorf("table %q missing column %q", tt.table, col)
				}
			}
		})
	}
}

func TestInitDB_CloseDB(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if err := db.Ping(); err == nil {
		t.Error("expected error after close, got nil")
	}
}

func TestInitDB_ExecQuery(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO opportunities (id, company, role, status, created_at, updated_at) VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))",
		"test-id", "Acme Corp", "Engineer", "active")
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	var company string
	err = db.QueryRow("SELECT company FROM opportunities WHERE id = ?", "test-id").Scan(&company)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if company != "Acme Corp" {
		t.Errorf("company = %q, want Acme Corp", company)
	}
}
