package store

import (
	"context"

	"github.com/marco/resume-app/internal/models"
)

type ArtifactStore interface {
	List(ctx context.Context) ([]models.Artifact, error)
	Get(ctx context.Context, id string) (models.Artifact, error)
	Create(ctx context.Context, a models.Artifact) error
	Delete(ctx context.Context, id string) error
}
