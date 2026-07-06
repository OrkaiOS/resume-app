package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marco/resume-app/internal/models"
)

const defaultLimit = 12

type SQLiteOpportunityStore struct {
	db *sql.DB
}

func NewSQLiteOpportunityStore(db *sql.DB) *SQLiteOpportunityStore {
	return &SQLiteOpportunityStore{db: db}
}

func (s *SQLiteOpportunityStore) List(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error) {
	if limit <= 0 || limit > 100 {
		limit = defaultLimit
	}

	var rows *sql.Rows
	var err error

	if cursor == "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, company, role, description, status, created_at, updated_at
			 FROM opportunities ORDER BY id LIMIT ?`, limit+1)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, company, role, description, status, created_at, updated_at
			 FROM opportunities WHERE id > ? ORDER BY id LIMIT ?`, cursor, limit+1)
	}
	if err != nil {
		return nil, "", fmt.Errorf("store.SQLiteOpportunityStore.List: %w", err)
	}
	defer rows.Close()

	var results []models.Opportunity
	for rows.Next() {
		var o models.Opportunity
		var createdAt, updatedAt string
		if err := rows.Scan(&o.ID, &o.Company, &o.Role, &o.Description, &o.Status,
			&createdAt, &updatedAt); err != nil {
			return nil, "", fmt.Errorf("store.SQLiteOpportunityStore.List: scan: %w", err)
		}
		o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		results = append(results, o)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("store.SQLiteOpportunityStore.List: rows: %w", err)
	}

	var nextCursor string
	if len(results) > limit {
		nextCursor = results[limit-1].ID
		results = results[:limit]
	}

	return results, nextCursor, nil
}

func (s *SQLiteOpportunityStore) Get(ctx context.Context, id string) (models.Opportunity, error) {
	var o models.Opportunity
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, company, role, description, status, created_at, updated_at
		 FROM opportunities WHERE id = ?`, id).
		Scan(&o.ID, &o.Company, &o.Role, &o.Description, &o.Status,
			&createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return models.Opportunity{}, fmt.Errorf("store.SQLiteOpportunityStore.Get: %w", ErrNotFound)
	}
	if err != nil {
		return models.Opportunity{}, fmt.Errorf("store.SQLiteOpportunityStore.Get: %w", err)
	}

	o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return o, nil
}

func (s *SQLiteOpportunityStore) Create(ctx context.Context, o models.Opportunity) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	o.UpdatedAt = time.Now().UTC()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO opportunities (id, company, role, description, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.Company, o.Role, o.Description, o.Status,
		o.CreatedAt.Format(time.RFC3339), now)
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.Create: %w", err)
	}
	return nil
}

func (s *SQLiteOpportunityStore) Update(ctx context.Context, o models.Opportunity) error {
	now := time.Now().UTC().Format(time.RFC3339)
	o.UpdatedAt = time.Now().UTC()

	result, err := s.db.ExecContext(ctx,
		`UPDATE opportunities SET company=?, role=?, description=?, status=?, updated_at=?
		 WHERE id=?`,
		o.Company, o.Role, o.Description, o.Status, now, o.ID)
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.Update: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.Update: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("store.SQLiteOpportunityStore.Update: %w", ErrNotFound)
	}

	return nil
}

func (s *SQLiteOpportunityStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM opportunities WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.Delete: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.Delete: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("store.SQLiteOpportunityStore.Delete: %w", ErrNotFound)
	}

	return nil
}

func (s *SQLiteOpportunityStore) SetArchived(ctx context.Context, id string, archived bool) error {
	status := "active"
	if archived {
		status = "archived"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx,
		`UPDATE opportunities SET status=?, updated_at=? WHERE id=?`,
		status, now, id)
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.SetArchived: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.SQLiteOpportunityStore.SetArchived: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("store.SQLiteOpportunityStore.SetArchived: %w", ErrNotFound)
	}

	return nil
}
