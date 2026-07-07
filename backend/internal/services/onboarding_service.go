package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type OnboardingService struct {
	store    store.OnboardingStore
	profiles store.ProfileStore
}

func NewOnboardingService(s store.OnboardingStore) *OnboardingService {
	return &OnboardingService{store: s}
}

func NewOnboardingServiceWithProfiles(s store.OnboardingStore, p store.ProfileStore) *OnboardingService {
	return &OnboardingService{store: s, profiles: p}
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
	categoryID := ""

	if err := s.store.UpsertOrkaiIDs(ctx, categoryID, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID); err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveProfile: %w", err)
	}

	o, err := s.store.Get(ctx)
	if err != nil {
		return models.OnboardingState{}, fmt.Errorf("services.OnboardingService.SaveProfile: get after upsert: %w", err)
	}
	return o, nil
}

func (s *OnboardingService) SaveProfileData(ctx context.Context, p models.Profile) error {
	if s.profiles == nil {
		return fmt.Errorf("services.OnboardingService.SaveProfileData: profile store not configured")
	}
	if err := s.profiles.Upsert(ctx, p); err != nil {
		return fmt.Errorf("services.OnboardingService.SaveProfileData: %w", err)
	}
	return nil
}

func (s *OnboardingService) HasProfile(ctx context.Context) (bool, error) {
	if s.profiles == nil {
		return false, nil
	}
	_, err := s.profiles.Get(ctx)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("services.OnboardingService.HasProfile: %w", err)
	}
	return true, nil
}
