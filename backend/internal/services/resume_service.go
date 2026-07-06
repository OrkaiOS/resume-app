package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type ResumeService struct {
	store store.ResumeStore
}

func NewResumeService(s store.ResumeStore) *ResumeService {
	return &ResumeService{store: s}
}

func (s *ResumeService) GetByOpportunity(ctx context.Context, opportunityID string) (models.Resume, error) {
	r, err := s.store.GetByOpportunity(ctx, opportunityID)
	if err != nil {
		return models.Resume{}, fmt.Errorf("services.ResumeService.GetByOpportunity: %w", err)
	}
	return r, nil
}

func (s *ResumeService) Upsert(ctx context.Context, r models.Resume) (models.Resume, error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = "draft"
	}

	if err := s.store.Upsert(ctx, r); err != nil {
		return models.Resume{}, fmt.Errorf("services.ResumeService.Upsert: %w", err)
	}

	return s.store.GetByOpportunity(ctx, r.OpportunityID)
}
