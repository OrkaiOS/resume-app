package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type ProfileService struct {
	store store.ProfileStore
}

func NewProfileService(s store.ProfileStore) *ProfileService {
	return &ProfileService{store: s}
}

func (s *ProfileService) Get(ctx context.Context) (models.Profile, error) {
	p, err := s.store.Get(ctx)
	if err != nil {
		return models.Profile{}, fmt.Errorf("services.ProfileService.Get: %w", err)
	}
	return p, nil
}

func (s *ProfileService) Upsert(ctx context.Context, p models.Profile) (models.Profile, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	if err := s.store.Upsert(ctx, p); err != nil {
		return models.Profile{}, fmt.Errorf("services.ProfileService.Upsert: %w", err)
	}

	return s.store.Get(ctx)
}
