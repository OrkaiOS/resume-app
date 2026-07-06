package store

import (
	"context"

	"github.com/marco/resume-app/internal/models"
)

type CoverLetterStore interface {
	GetByOpportunity(ctx context.Context, opportunityID string) (models.CoverLetter, error)
	Upsert(ctx context.Context, cl models.CoverLetter) error
}
