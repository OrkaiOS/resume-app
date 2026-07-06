package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type ArtifactService struct {
	store store.ArtifactStore
}

func NewArtifactService(s store.ArtifactStore) *ArtifactService {
	return &ArtifactService{store: s}
}

func (s *ArtifactService) List(ctx context.Context) ([]models.Artifact, error) {
	items, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("services.ArtifactService.List: %w", err)
	}
	return items, nil
}

func (s *ArtifactService) Get(ctx context.Context, id string) (models.Artifact, error) {
	a, err := s.store.Get(ctx, id)
	if err != nil {
		return models.Artifact{}, fmt.Errorf("services.ArtifactService.Get: %w", err)
	}
	return a, nil
}

func (s *ArtifactService) Create(ctx context.Context, a models.Artifact) (models.Artifact, error) {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.Name == "" {
		return models.Artifact{}, fmt.Errorf("services.ArtifactService.Create: name is required")
	}
	if a.Type == "" {
		return models.Artifact{}, fmt.Errorf("services.ArtifactService.Create: type is required")
	}

	if err := s.store.Create(ctx, a); err != nil {
		return models.Artifact{}, fmt.Errorf("services.ArtifactService.Create: %w", err)
	}

	return s.store.Get(ctx, a.ID)
}

func (s *ArtifactService) Delete(ctx context.Context, id string) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("services.ArtifactService.Delete: %w", err)
	}
	return nil
}
