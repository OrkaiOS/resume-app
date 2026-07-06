package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type CoverLetterService struct {
	store store.CoverLetterStore
}

func NewCoverLetterService(s store.CoverLetterStore) *CoverLetterService {
	return &CoverLetterService{store: s}
}

func (s *CoverLetterService) GetByOpportunity(ctx context.Context, opportunityID string) (models.CoverLetter, error) {
	cl, err := s.store.GetByOpportunity(ctx, opportunityID)
	if err != nil {
		return models.CoverLetter{}, fmt.Errorf("services.CoverLetterService.GetByOpportunity: %w", err)
	}
	return cl, nil
}

func (s *CoverLetterService) Upsert(ctx context.Context, cl models.CoverLetter) (models.CoverLetter, error) {
	if cl.ID == "" {
		cl.ID = uuid.New().String()
	}
	if cl.Status == "" {
		cl.Status = "draft"
	}

	if err := s.store.Upsert(ctx, cl); err != nil {
		return models.CoverLetter{}, fmt.Errorf("services.CoverLetterService.Upsert: %w", err)
	}

	return s.store.GetByOpportunity(ctx, cl.OpportunityID)
}
