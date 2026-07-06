package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marco/resume-app/internal/models"
)

type SQLiteCoverLetterStore struct {
	db *sql.DB
}

func NewSQLiteCoverLetterStore(db *sql.DB) *SQLiteCoverLetterStore {
	return &SQLiteCoverLetterStore{db: db}
}

func (s *SQLiteCoverLetterStore) GetByOpportunity(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
	var cl models.CoverLetter
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, opportunity_id, markdown_content, pdf_path, status, created_at, updated_at
		 FROM cover_letters WHERE opportunity_id = ?`, opportunityID).
		Scan(&cl.ID, &cl.OpportunityID, &cl.MarkdownContent, &cl.PDFPath, &cl.Status,
			&createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return models.CoverLetter{}, fmt.Errorf("store.SQLiteCoverLetterStore.GetByOpportunity: %w", ErrNotFound)
	}
	if err != nil {
		return models.CoverLetter{}, fmt.Errorf("store.SQLiteCoverLetterStore.GetByOpportunity: %w", err)
	}

	cl.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	cl.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return cl, nil
}

func (s *SQLiteCoverLetterStore) Upsert(ctx context.Context, cl models.CoverLetter) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if cl.CreatedAt.IsZero() {
		cl.CreatedAt = time.Now().UTC()
	}
	cl.UpdatedAt = time.Now().UTC()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO cover_letters (id, opportunity_id, markdown_content, pdf_path, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			markdown_content=excluded.markdown_content,
			pdf_path=excluded.pdf_path,
			status=excluded.status,
			updated_at=excluded.updated_at`,
		cl.ID, cl.OpportunityID, cl.MarkdownContent, cl.PDFPath, cl.Status,
		cl.CreatedAt.Format(time.RFC3339), now)
	if err != nil {
		return fmt.Errorf("store.SQLiteCoverLetterStore.Upsert: %w", err)
	}

	return nil
}
