package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marco/resume-app/internal/orkai"
	"github.com/marco/resume-app/internal/store"
)

const userInsightsStandardName = "User Insights — Resume & Cover Letter"

// InterruptedAt records the state of a conversation when the user clicks Stop.
// It is serialized into the session summary so the agent can resume context
// in a future session via overview (FR-030, FR-034).
type InterruptedAt struct {
	LastUserMessage   string `json:"lastUserMessage"`
	LastAssistantText string `json:"lastAssistantText"`
	Iteration         int    `json:"iteration"`
	Phase             string `json:"phase"` // "llm-call" or "tool-exec"
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision SessionService wraps the orkai client for session and user-insight persistence. It is the single place where session names, metadata shapes, and the User Insights standard name are defined. The agent tools (save_session, update_session, save_user_insight) delegate to this service.
type SessionService struct {
	client          *orkai.OrkaiClient
	onboardingStore store.OnboardingStore
}

// NewSessionService creates a SessionService.
func NewSessionService(client *orkai.OrkaiClient, onboardingStore store.OnboardingStore) *SessionService {
	return &SessionService{client: client, onboardingStore: onboardingStore}
}

// Save creates a new orkai session entity in the personal category with
// opportunity metadata. Returns the new session ID.
func (s *SessionService) Save(ctx context.Context, opportunityID, company, role, summary string) (string, error) {
	state, err := s.onboardingStore.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("services.SessionService.Save: getting category ID: %w", err)
	}
	if state.OrkaiCategoryID == "" {
		return "", fmt.Errorf("services.SessionService.Save: orkai category not configured")
	}

	name := fmt.Sprintf("Session: %s %s — %s", company, role, time.Now().Format("2006-01-02"))
	metadata := map[string]string{
		"opportunityId": opportunityID,
		"company":       company,
		"role":          role,
		"date":          time.Now().Format("2006-01-02"),
	}

	return s.client.CreateSession(ctx, name, summary, state.OrkaiCategoryID, metadata)
}

// Update updates an existing orkai session by ID with new summary text.
func (s *SessionService) Update(ctx context.Context, sessionID, summary string) error {
	return s.client.UpdateSession(ctx, sessionID, summary)
}

// SaveUserInsight creates or updates the User Insights standard in the
// personal orkai category. If the standard already exists, it appends the
// new insight; otherwise it creates a new standard.
func (s *SessionService) SaveUserInsight(ctx context.Context, insight string) error {
	state, err := s.onboardingStore.Get(ctx)
	if err != nil {
		return fmt.Errorf("services.SessionService.SaveUserInsight: getting category ID: %w", err)
	}
	if state.OrkaiCategoryID == "" {
		return fmt.Errorf("services.SessionService.SaveUserInsight: orkai category not configured")
	}

	// Search for existing User Insights standard.
	items, err := s.client.SearchStandards(ctx, state.OrkaiCategoryID, "User Insights")
	if err != nil {
		return fmt.Errorf("services.SessionService.SaveUserInsight: search: %w", err)
	}

	for _, item := range items {
		if strings.Contains(item.Name, "User Insights") {
			// Update existing standard — append the new insight.
			updatedText := item.Text
			if updatedText != "" {
				updatedText += "\n\n"
			}
			updatedText += fmt.Sprintf("- %s (added %s)", insight, time.Now().Format("2006-01-02"))
			return s.client.UpdateStandard(ctx, item.ID, updatedText)
		}
	}

	// No existing standard found — create one.
	text := fmt.Sprintf("# User Insights — Resume & Cover Letter\n\nDurable user-specific guidance captured by the agent during chat sessions. These insights are loaded into the system prompt for every future session.\n\n- %s (added %s)", insight, time.Now().Format("2006-01-02"))
	_, err = s.client.CreateStandard(ctx, userInsightsStandardName, text, state.OrkaiCategoryID)
	if err != nil {
		return fmt.Errorf("services.SessionService.SaveUserInsight: create standard: %w", err)
	}
	return nil
}

// SaveInterrupted saves or updates a session with an interrupted_at marker.
// If sessionID is empty, a new session is created. If set, the existing
// session is updated. The interrupted_at marker records the last user
// message, last assistant text, iteration count, and whether the stop
// happened mid-LLM-call or mid-tool-execution.
func (s *SessionService) SaveInterrupted(ctx context.Context, opportunityID, sessionID, company, role, summary string, interrupted InterruptedAt) error {
	marker := fmt.Sprintf("\n\n---\n## Interrupted at\n\n- **Last user message**: %s\n- **Last assistant text**: %s\n- **Iteration**: %d\n- **Phase**: %s\n",
		interrupted.LastUserMessage, interrupted.LastAssistantText, interrupted.Iteration, interrupted.Phase)

	if sessionID != "" {
		// Update existing session — append the interrupted marker.
		// We need the current text to append to it. Since we don't have
		// a GetSession method, we update with just the marker appended
		// to the summary. The orkai session tool's update replaces text,
		// so we include the summary + marker.
		text := summary + marker
		return s.client.UpdateSession(ctx, sessionID, text)
	}

	// No existing session — create a new one with the interrupted marker.
	text := summary + marker
	_, err := s.Save(ctx, opportunityID, company, role, text)
	return err
}

// GetUserInsightsText searches for the User Insights standard and returns
// its text content. Returns empty string if not found.
func (s *SessionService) GetUserInsightsText(ctx context.Context) (string, error) {
	state, err := s.onboardingStore.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("services.SessionService.GetUserInsightsText: getting category ID: %w", err)
	}
	if state.OrkaiCategoryID == "" {
		return "", nil
	}

	items, err := s.client.SearchStandards(ctx, state.OrkaiCategoryID, "User Insights")
	if err != nil {
		return "", fmt.Errorf("services.SessionService.GetUserInsightsText: search: %w", err)
	}

	for _, item := range items {
		if strings.Contains(item.Name, "User Insights") {
			return item.Text, nil
		}
	}
	return "", nil
}
