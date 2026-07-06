package models

import "time"

type Artifact struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Description   string    `json:"description"`
	ScriptContent string    `json:"scriptContent"`
	UsageCount    int       `json:"usageCount"`
	CreatedAt     time.Time `json:"createdAt"`
	LastUsedAt    time.Time `json:"lastUsedAt"`
}
