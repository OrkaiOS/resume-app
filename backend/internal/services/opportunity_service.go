package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type OpportunityService struct {
	store store.OpportunityStore
}

func NewOpportunityService(s store.OpportunityStore) *OpportunityService {
	return &OpportunityService{store: s}
}

func (s *OpportunityService) List(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error) {
	items, next, err := s.store.List(ctx, cursor, limit)
	if err != nil {
		return nil, "", fmt.Errorf("services.OpportunityService.List: %w", err)
	}
	return items, next, nil
}

func (s *OpportunityService) Get(ctx context.Context, id string) (models.Opportunity, error) {
	o, err := s.store.Get(ctx, id)
	if err != nil {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Get: %w", err)
	}
	return o, nil
}

func (s *OpportunityService) Create(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	if o.Company == "" {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Create: company is required")
	}
	if o.Role == "" {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Create: role is required")
	}
	if o.Status == "" {
		o.Status = "active"
	}

	if err := s.store.Create(ctx, o); err != nil {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Create: %w", err)
	}

	return s.store.Get(ctx, o.ID)
}

func (s *OpportunityService) Update(ctx context.Context, o models.Opportunity) (models.Opportunity, error) {
	if o.Company == "" {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Update: company is required")
	}
	if o.Role == "" {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Update: role is required")
	}

	if err := s.store.Update(ctx, o); err != nil {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.Update: %w", err)
	}

	return s.store.Get(ctx, o.ID)
}

func (s *OpportunityService) Delete(ctx context.Context, id string) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("services.OpportunityService.Delete: %w", err)
	}
	return nil
}

func (s *OpportunityService) SetArchived(ctx context.Context, id string, archived bool) (models.Opportunity, error) {
	if err := s.store.SetArchived(ctx, id, archived); err != nil {
		return models.Opportunity{}, fmt.Errorf("services.OpportunityService.SetArchived: %w", err)
	}
	return s.store.Get(ctx, id)
}
