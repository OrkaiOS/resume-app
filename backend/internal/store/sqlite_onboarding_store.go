package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marco/resume-app/internal/models"
)

const onboardingRowID = "default"

type SQLiteOnboardingStore struct {
	db *sql.DB
}

func NewSQLiteOnboardingStore(db *sql.DB) *SQLiteOnboardingStore {
	return &SQLiteOnboardingStore{db: db}
}

func (s *SQLiteOnboardingStore) Get(ctx context.Context) (models.OnboardingState, error) {
	var o models.OnboardingState
	var onboardedAt, updatedAt string
	var llmProvider, llmModel, llmAPIKey sql.NullString
	var orkaiCategoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT llm_provider, llm_model, llm_api_key,
		        orkai_category_id, canonical_profile_standard_id,
		        cover_letter_principles_standard_id, pdf_pipeline_standard_id,
		        pdf_generation_skill_id, onboarded_at, updated_at
		 FROM user_settings WHERE id = ?`, onboardingRowID).
		Scan(&llmProvider, &llmModel, &llmAPIKey,
			&orkaiCategoryID, &profileStdID, &coverLetterStdID,
			&pdfPipelineStdID, &pdfSkillID,
			&onboardedAt, &updatedAt)
	if err == sql.ErrNoRows {
		return models.OnboardingState{}, fmt.Errorf("store.SQLiteOnboardingStore.Get: %w", ErrNotFound)
	}
	if err != nil {
		return models.OnboardingState{}, fmt.Errorf("store.SQLiteOnboardingStore.Get: %w", err)
	}

	o.ID = onboardingRowID
	o.LLMProvider = llmProvider.String
	o.LLMModel = llmModel.String
	o.LLMAPIKey = llmAPIKey.String
	o.OrkaiCategoryID = orkaiCategoryID.String
	o.CanonicalProfileStandardID = profileStdID.String
	o.CoverLetterPrinciplesStandardID = coverLetterStdID.String
	o.PDFPipelineStandardID = pdfPipelineStdID.String
	o.PDFGenerationSkillID = pdfSkillID.String
	o.OnboardedAt, _ = time.Parse(time.RFC3339, onboardedAt)
	o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return o, nil
}

func (s *SQLiteOnboardingStore) UpsertLLMConfig(ctx context.Context, provider, model, apiKey string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_settings (id, llm_provider, llm_model, llm_api_key, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			llm_provider=excluded.llm_provider,
			llm_model=excluded.llm_model,
			llm_api_key=excluded.llm_api_key,
			updated_at=excluded.updated_at`,
		onboardingRowID, provider, model, apiKey, now, now)
	if err != nil {
		return fmt.Errorf("store.SQLiteOnboardingStore.UpsertLLMConfig: %w", err)
	}

	return nil
}

func (s *SQLiteOnboardingStore) UpsertOrkaiIDs(ctx context.Context, categoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_settings (id, orkai_category_id, canonical_profile_standard_id,
		        cover_letter_principles_standard_id, pdf_pipeline_standard_id,
		        pdf_generation_skill_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			orkai_category_id=excluded.orkai_category_id,
			canonical_profile_standard_id=excluded.canonical_profile_standard_id,
			cover_letter_principles_standard_id=excluded.cover_letter_principles_standard_id,
			pdf_pipeline_standard_id=excluded.pdf_pipeline_standard_id,
			pdf_generation_skill_id=excluded.pdf_generation_skill_id,
			updated_at=excluded.updated_at`,
		onboardingRowID, categoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID, now, now)
	if err != nil {
		return fmt.Errorf("store.SQLiteOnboardingStore.UpsertOrkaiIDs: %w", err)
	}

	return nil
}

func (s *SQLiteOnboardingStore) MarkComplete(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_settings (id, onboarded_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			onboarded_at=excluded.onboarded_at,
			updated_at=excluded.updated_at`,
		onboardingRowID, now, now, now)
	if err != nil {
		return fmt.Errorf("store.SQLiteOnboardingStore.MarkComplete: %w", err)
	}

	return nil
}
