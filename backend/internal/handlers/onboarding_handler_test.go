package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marco/resume-app/internal/models"
	"github.com/marco/resume-app/internal/store"
)

type fakeOnboardingService struct {
	getStatusFn       func(ctx context.Context) (models.OnboardingState, error)
	saveLLMConfigFn   func(ctx context.Context, provider, model, apiKey string) (models.OnboardingState, error)
	saveProfileFn     func(ctx context.Context, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) (models.OnboardingState, error)
	saveProfileDataFn func(ctx context.Context, p models.Profile) error
	hasProfileFn      func(ctx context.Context) (bool, error)
}

func (f *fakeOnboardingService) GetStatus(ctx context.Context) (models.OnboardingState, error) {
	return f.getStatusFn(ctx)
}

func (f *fakeOnboardingService) SaveLLMConfig(ctx context.Context, provider, model, apiKey string) (models.OnboardingState, error) {
	return f.saveLLMConfigFn(ctx, provider, model, apiKey)
}

func (f *fakeOnboardingService) SaveProfile(ctx context.Context, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) (models.OnboardingState, error) {
	return f.saveProfileFn(ctx, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID)
}

func (f *fakeOnboardingService) SaveProfileData(ctx context.Context, p models.Profile) error {
	if f.saveProfileDataFn == nil {
		return nil
	}
	return f.saveProfileDataFn(ctx, p)
}

func (f *fakeOnboardingService) HasProfile(ctx context.Context) (bool, error) {
	if f.hasProfileFn == nil {
		return false, nil
	}
	return f.hasProfileFn(ctx)
}

func TestOnboarding_SaveLLMConfig_ReturnsState(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		saveLLMConfigFn: func(ctx context.Context, provider, model, apiKey string) (models.OnboardingState, error) {
			return models.OnboardingState{
				LLMProvider: provider,
				LLMModel:    model,
				LLMAPIKey:   apiKey,
			}, nil
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.POST("/v1/api/onboarding/llm-config", h.SaveLLMConfig)

	body := `{"provider":"ollama","model":"llama3","apiKey":""}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/llm-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	steps, ok := data["steps"].(map[string]interface{})
	if !ok {
		t.Fatalf("steps is not an object: %T", data["steps"])
	}
	if steps["llmConfig"] != true {
		t.Errorf("steps.llmConfig = %v, want true", steps["llmConfig"])
	}
}

func TestOnboarding_SaveLLMConfig_MissingProvider(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{}
	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.POST("/v1/api/onboarding/llm-config", h.SaveLLMConfig)

	body := `{"model":"llama3"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/llm-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestOnboarding_SaveLLMConfig_MissingModel(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{}
	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.POST("/v1/api/onboarding/llm-config", h.SaveLLMConfig)

	body := `{"provider":"ollama"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/llm-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestOnboarding_SaveProfile_ReturnsState(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		saveProfileFn: func(ctx context.Context, profileStdID, coverLetterStdID, pdfPipelineStdID, pdfSkillID string) (models.OnboardingState, error) {
			return models.OnboardingState{
				CanonicalProfileStandardID:      profileStdID,
				CoverLetterPrinciplesStandardID: coverLetterStdID,
				PDFPipelineStandardID:           pdfPipelineStdID,
				PDFGenerationSkillID:            pdfSkillID,
			}, nil
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.POST("/v1/api/onboarding/profile", h.SaveProfile)

	body := `{"profileStandardId":"ps1","coverLetterStandardId":"cl1","pdfPipelineStandardId":"pp1","pdfGenerationSkillId":"pg1"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", env)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	steps, ok := data["steps"].(map[string]interface{})
	if !ok {
		t.Fatalf("steps is not an object: %T", data["steps"])
	}
	if steps["profile"] != true {
		t.Errorf("steps.profile = %v, want true", steps["profile"])
	}
}

func TestOnboarding_SaveProfile_ProfileDataOnly(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		getStatusFn: func(ctx context.Context) (models.OnboardingState, error) {
			return models.OnboardingState{
				LLMProvider: "ollama",
				LLMModel:    "llama3",
			}, nil
		},
		saveProfileDataFn: func(ctx context.Context, p models.Profile) error {
			return nil
		},
		hasProfileFn: func(ctx context.Context) (bool, error) {
			return true, nil
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.POST("/v1/api/onboarding/profile", h.SaveProfile)

	body := `{"fullName":"John Doe","email":"john@example.com"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/api/onboarding/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", env)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	steps, ok := data["steps"].(map[string]interface{})
	if !ok {
		t.Fatalf("steps is not an object: %T", data["steps"])
	}
	if steps["profile"] != true {
		t.Errorf("steps.profile = %v, want true", steps["profile"])
	}
}

func TestOnboarding_GetStatus_Onboarded(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		getStatusFn: func(ctx context.Context) (models.OnboardingState, error) {
			return models.OnboardingState{
				LLMProvider:                     "ollama",
				LLMModel:                        "llama3",
				OrkaiCategoryID:                 "cat1",
				CanonicalProfileStandardID:      "ps1",
				CoverLetterPrinciplesStandardID: "cl1",
				PDFPipelineStandardID:           "pp1",
				PDFGenerationSkillID:            "pg1",
			}, nil
		},
		hasProfileFn: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.GET("/v1/api/onboarding/status", h.GetStatus)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/status", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", env)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	steps, ok := data["steps"].(map[string]interface{})
	if !ok {
		t.Fatalf("steps is not an object: %T", data["steps"])
	}
	if steps["llmConfig"] != true {
		t.Errorf("steps.llmConfig = %v, want true", steps["llmConfig"])
	}
	if steps["profile"] != true {
		t.Errorf("steps.profile = %v, want true", steps["profile"])
	}
	if steps["orkaiSetup"] != true {
		t.Errorf("steps.orkaiSetup = %v, want true", steps["orkaiSetup"])
	}
	if data["onboarded"] != true {
		t.Errorf("onboarded = %v, want true when all three steps are done", data["onboarded"])
	}
}

func TestOnboarding_GetStatus_NotOnboarded(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		getStatusFn: func(ctx context.Context) (models.OnboardingState, error) {
			return models.OnboardingState{}, store.ErrNotFound
		},
		hasProfileFn: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.GET("/v1/api/onboarding/status", h.GetStatus)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/status", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", env)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	steps, ok := data["steps"].(map[string]interface{})
	if !ok {
		t.Fatalf("steps is not an object: %T", data["steps"])
	}
	if steps["llmConfig"] != false {
		t.Errorf("steps.llmConfig = %v, want false", steps["llmConfig"])
	}
	if steps["profile"] != false {
		t.Errorf("steps.profile = %v, want false", steps["profile"])
	}
	if steps["orkaiSetup"] != false {
		t.Errorf("steps.orkaiSetup = %v, want false", steps["orkaiSetup"])
	}
}

func TestOnboarding_GetStatus_Partial(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		getStatusFn: func(ctx context.Context) (models.OnboardingState, error) {
			return models.OnboardingState{
				LLMProvider: "openai",
				LLMModel:    "gpt-4",
			}, nil
		},
		hasProfileFn: func(ctx context.Context) (bool, error) {
			return false, nil
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.GET("/v1/api/onboarding/status", h.GetStatus)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/status", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", env)
	}
	if env.Error != nil {
		t.Fatalf("error = %+v, want nil", env.Error)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object: %T", env.Data)
	}
	steps, ok := data["steps"].(map[string]interface{})
	if !ok {
		t.Fatalf("steps is not an object: %T", data["steps"])
	}
	if steps["llmConfig"] != true {
		t.Errorf("steps.llmConfig = %v, want true", steps["llmConfig"])
	}
	if steps["profile"] != false {
		t.Errorf("steps.profile = %v, want false", steps["profile"])
	}
	if steps["orkaiSetup"] != false {
		t.Errorf("steps.orkaiSetup = %v, want false", steps["orkaiSetup"])
	}
}

func TestOnboarding_GetStatus_ServiceError(t *testing.T) {
	t.Parallel()

	svc := &fakeOnboardingService{
		getStatusFn: func(ctx context.Context) (models.OnboardingState, error) {
			return models.OnboardingState{}, errors.New("db connection failed")
		},
	}

	router := gin.New()
	h := NewOnboardingHandler(svc)
	router.GET("/v1/api/onboarding/status", h.GetStatus)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/api/onboarding/status", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	var env Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("invalid json: %v", env)
	}
	if env.Error == nil {
		t.Fatal("error is nil, want non-nil")
	}
}
