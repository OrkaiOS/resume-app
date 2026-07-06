package models

import "time"

type CoverLetter struct {
	ID              string    `json:"id"`
	OpportunityID   string    `json:"opportunityId"`
	MarkdownContent string    `json:"markdownContent"`
	PDFPath         string    `json:"pdfPath"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
