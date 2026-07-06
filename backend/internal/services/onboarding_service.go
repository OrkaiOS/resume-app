package services

import (
	"context"
	"fmt"

	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type OnboardingService struct {
	store store.OnboardingStore
}

func NewOnboardingService(s store.OnboardingStore) *OnboardingService {
	return &OnboardingService{store: s}
}

func (s *OnboardingService) GetStatus(ctx context.Context) (models.OnboardingState, error) {
	o, err := s.store.Get(ctx)
	if err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.GetStatus: %w", err)
	}
	return o, nil
}

func (s *OnboardingService) SaveLLMConfig(ctx context.Context, provider, model, apiKey string) (models.OnboardingState, error) {
	if provider == "" {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveLLMConfig: provider is required")
	}
	if model == "" {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveLLMConfig: model is required")
	}

	if err := s.store.UpsertLLMConfig(ctx, provider, model, apiKey); err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveLLMConfig: %w", err)
	}

	o, err := s.store.Get(ctx)
	if err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveLLMConfig: get after upsert: %w", err)
	}
	return o, nil
}

func (s *OnboardingService) SaveProfile(ctx context.Context, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) (models.OnboardingState, error) {
	categoryID := "" // T3 populates this via orkai MCP setup

	if err := s.store.UpsertOrkaiIDs(ctx, categoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID); err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveProfile: %w", err)
	}

	o, err := s.store.Get(ctx)
	if err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveProfile: get after upsert: %w", err)
	}
	return o, nil
}
