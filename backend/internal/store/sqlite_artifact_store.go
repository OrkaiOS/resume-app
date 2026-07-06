package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marco/resume-app/internal/models"
)

type SQLiteArtifactStore struct {
	db *sql.DB
}

func NewSQLiteArtifactStore(db *sql.DB) *SQLiteArtifactStore {
	return &SQLiteArtifactStore{db: db}
}

func (s *SQLiteArtifactStore) List(ctx context.Context) ([]models.Artifact, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, description, script_content, usage_count, created_at, last_used_at
		 FROM artifacts ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("store.SQLiteArtifactStore.List: %w", err)
	}
	defer rows.Close()

	var artifacts []models.Artifact
	for rows.Next() {
		var a models.Artifact
		var createdAt, lastUsedAt string
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Description, &a.ScriptContent,
			&a.UsageCount, &createdAt, &lastUsedAt); err != nil {
			return nil, fmt.Errorf("store.SQLiteArtifactStore.List: %w", err)
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		a.LastUsedAt, _ = time.Parse(time.RFC3339, lastUsedAt)
		artifacts = append(artifacts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.SQLiteArtifactStore.List: %w", err)
	}

	if artifacts == nil {
		artifacts = []models.Artifact{}
	}

	return artifacts, nil
}

func (s *SQLiteArtifactStore) Get(ctx context.Context, id string) (models.Artifact, error) {
	var a models.Artifact
	var createdAt, lastUsedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, description, script_content, usage_count, created_at, last_used_at
		 FROM artifacts WHERE id = ?`, id).
		Scan(&a.ID, &a.Name, &a.Type, &a.Description, &a.ScriptContent,
			&a.UsageCount, &createdAt, &lastUsedAt)
	if err == sql.ErrNoRows {
		return models.Artifact{}, fmt.Errorf("store.SQLiteArtifactStore.Get: %w", ErrNotFound)
	}
	if err != nil {
		return models.Artifact{}, fmt.Errorf("store.SQLiteArtifactStore.Get: %w", err)
	}

	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	a.LastUsedAt, _ = time.Parse(time.RFC3339, lastUsedAt)

	return a, nil
}

func (s *SQLiteArtifactStore) Create(ctx context.Context, a models.Artifact) error {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	if a.LastUsedAt.IsZero() {
		a.LastUsedAt = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO artifacts (id, name, type, description, script_content, usage_count, created_at, last_used_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Type, a.Description, a.ScriptContent, a.UsageCount,
		a.CreatedAt.Format(time.RFC3339), a.LastUsedAt.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("store.SQLiteArtifactStore.Create: %w", err)
	}

	return nil
}

func (s *SQLiteArtifactStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM artifacts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("store.SQLiteArtifactStore.Delete: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.SQLiteArtifactStore.Delete: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("store.SQLiteArtifactStore.Delete: %w", ErrNotFound)
	}

	return nil
}
