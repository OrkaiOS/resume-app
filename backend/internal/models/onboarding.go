package models

import "time"

type OnboardingState struct {
	ID                              string
	LLMProvider                     string
	LLMModel                        string
	LLMAPIKey                       string
	OrkaiCategoryID                 string
	CanonicalProfileStandardID      string
	CoverLetterPrinciplesStandardID string
	PDFPipelineStandardID           string
	PDFGenerationSkillID            string
	OnboardedAt                     time.Time
	UpdatedAt                       time.Time
}
