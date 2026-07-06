package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/marco/resume-app/internal/models"
)

type SQLiteProfileStore struct {
	db *sql.DB
}

func NewSQLiteProfileStore(db *sql.DB) *SQLiteProfileStore {
	return &SQLiteProfileStore{db: db}
}

func (s *SQLiteProfileStore) Get(ctx context.Context) (models.Profile, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, full_name, email, phone, location,
		linkedin_url, website_url, github_url, professional_summary,
		work_experience, education, skills, projects, certifications, languages,
		created_at, updated_at FROM profiles LIMIT 1`)

	var p models.Profile
	var workExp, edu, skills, projects, certs, langs string
	var createdAt, updatedAt string

	err := row.Scan(&p.ID, &p.FullName, &p.Email, &p.Phone, &p.Location,
		&p.LinkedInURL, &p.WebsiteURL, &p.GitHubURL, &p.ProfessionalSummary,
		&workExp, &edu, &skills, &projects, &certs, &langs,
		&createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: %w", ErrNotFound)
	}
	if err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: %w", err)
	}

	if err := json.Unmarshal([]byte(workExp), &p.WorkExperience); err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: unmarshal work_experience: %w", err)
	}
	if err := json.Unmarshal([]byte(edu), &p.Education); err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: unmarshal education: %w", err)
	}
	if err := json.Unmarshal([]byte(skills), &p.Skills); err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: unmarshal skills: %w", err)
	}
	if err := json.Unmarshal([]byte(projects), &p.Projects); err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: unmarshal projects: %w", err)
	}
	if err := json.Unmarshal([]byte(certs), &p.Certifications); err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: unmarshal certifications: %w", err)
	}
	if err := json.Unmarshal([]byte(langs), &p.Languages); err != nil {
		return models.Profile{}, fmt.Errorf("store.SQLiteProfileStore.Get: unmarshal languages: %w", err)
	}

	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return p, nil
}

func (s *SQLiteProfileStore) Upsert(ctx context.Context, p models.Profile) error {
	workExp, err := json.Marshal(p.WorkExperience)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: marshal work_experience: %w", err)
	}
	edu, err := json.Marshal(p.Education)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: marshal education: %w", err)
	}
	skills, err := json.Marshal(p.Skills)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: marshal skills: %w", err)
	}
	projects, err := json.Marshal(p.Projects)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: marshal projects: %w", err)
	}
	certs, err := json.Marshal(p.Certifications)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: marshal certifications: %w", err)
	}
	langs, err := json.Marshal(p.Languages)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: marshal languages: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	p.UpdatedAt = time.Now().UTC()

	_, err = s.db.ExecContext(ctx, `INSERT INTO profiles
		(id, full_name, email, phone, location, linkedin_url, website_url, github_url,
		 professional_summary, work_experience, education, skills, projects,
		 certifications, languages, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			full_name=excluded.full_name, email=excluded.email, phone=excluded.phone,
			location=excluded.location, linkedin_url=excluded.linkedin_url,
			website_url=excluded.website_url, github_url=excluded.github_url,
			professional_summary=excluded.professional_summary,
			work_experience=excluded.work_experience, education=excluded.education,
			skills=excluded.skills, projects=excluded.projects,
			certifications=excluded.certifications, languages=excluded.languages,
			updated_at=excluded.updated_at`,
		p.ID, p.FullName, p.Email, p.Phone, p.Location,
		p.LinkedInURL, p.WebsiteURL, p.GitHubURL, p.ProfessionalSummary,
		string(workExp), string(edu), string(skills), string(projects),
		string(certs), string(langs),
		p.CreatedAt.Format(time.RFC3339), now)
	if err != nil {
		return fmt.Errorf("store.SQLiteProfileStore.Upsert: %w", err)
	}

	return nil
}
