package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/marco/resume-app/internal/store"
)

// ErrProfileNotFound is the service-level sentinel returned by
// ToolsService.Profile when the user has not onboarded a profile yet.
// Handlers check errors.Is against this sentinel instead of importing
// the store package directly.
var ErrProfileNotFound = errors.New("profile not found")

// ToolsService exposes the agent tools as service methods for the
// REST contract endpoints. The handler depends on this single service
// interface instead of multiple store-like provider interfaces,
// keeping the handler layer free of business-logic dependencies.
// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
type ToolsService struct {
	shell   *ShellService
	search  *OrkaiSearchService
	profile *ProfileService
}

func NewToolsService(shell *ShellService, search *OrkaiSearchService, profile *ProfileService) *ToolsService {
	return &ToolsService{shell: shell, search: search, profile: profile}
}

// Shell executes a sandboxed shell command.
func (s *ToolsService) Shell(ctx context.Context, command, language string) (ShellResult, error) {
	return s.shell.Execute(ctx, command, language)
}

// OrkaiSearch searches the user's orkai workspace.
func (s *ToolsService) OrkaiSearch(ctx context.Context, query string) (string, error) {
	return s.search.Search(ctx, query)
}

// Profile returns the user's structured profile (read-only). Returns
// ErrProfileNotFound (a service-level sentinel) when the user has not
// onboarded a profile yet, so handlers do not need to import the store
// package.
func (s *ToolsService) Profile(ctx context.Context) (any, error) {
	profile, err := s.profile.Get(ctx)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("services.ToolsService.Profile: %w", ErrProfileNotFound)
		}
		return nil, fmt.Errorf("services.ToolsService.Profile: %w", err)
	}
	return profile, nil
}
