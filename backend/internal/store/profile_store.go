package store

import (
	"context"

	"github.com/marco/resume-app/internal/models"
)

type ProfileStore interface {
	Get(ctx context.Context) (models.Profile, error)
	Upsert(ctx context.Context, p models.Profile) error
}
