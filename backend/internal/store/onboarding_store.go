package store

import (
	"context"

	"github.com/marco/resume-app/internal/models"
)

type OnboardingStore interface {
	Get(ctx context.Context) (models.OnboardingState, error)
	UpsertLLMConfig(ctx context.Context, provider, model, apiKey string) error
	UpsertOrkaiIDs(ctx context.Context, categoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) error
	MarkComplete(ctx context.Context) error
}
