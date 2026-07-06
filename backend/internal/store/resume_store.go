package store

import (
	"context"

	"github.com/marco/resume-app/internal/models"
)

type ResumeStore interface {
	GetByOpportunity(ctx context.Context, opportunityID string) (models.Resume, error)
	Upsert(ctx context.Context, r models.Resume) error
}
