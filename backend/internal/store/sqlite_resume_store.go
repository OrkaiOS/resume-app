package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marco/resume-app/internal/models"
)

type SQLiteResumeStore struct {
	db *sql.DB
}

func NewSQLiteResumeStore(db *sql.DB) *SQLiteResumeStore {
	return &SQLiteResumeStore{db: db}
}

func (s *SQLiteResumeStore) GetByOpportunity(ctx context.Context, opportunityID string) (models.Resume, error) {
	var r models.Resume
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, opportunity_id, markdown_content, pdf_path, status, created_at, updated_at
		 FROM resumes WHERE opportunity_id = ?`, opportunityID).
		Scan(&r.ID, &r.OpportunityID, &r.MarkdownContent, &r.PDFPath, &r.Status,
			&createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return models.Resume{}, fmt.Errorf("store.SQLiteResumeStore.GetByOpportunity: %w", ErrNotFound)
	}
	if err != nil {
		return models.Resume{}, fmt.Errorf("store.SQLiteResumeStore.GetByOpportunity: %w", err)
	}

	r.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	r.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return r, nil
}

func (s *SQLiteResumeStore) Upsert(ctx context.Context, r models.Resume) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	r.UpdatedAt = time.Now().UTC()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO resumes (id, opportunity_id, markdown_content, pdf_path, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			markdown_content=excluded.markdown_content,
			pdf_path=excluded.pdf_path,
			status=excluded.status,
			updated_at=excluded.updated_at`,
		r.ID, r.OpportunityID, r.MarkdownContent, r.PDFPath, r.Status,
		r.CreatedAt.Format(time.RFC3339), now)
	if err != nil {
		return fmt.Errorf("store.SQLiteResumeStore.Upsert: %w", err)
	}

	return nil
}
