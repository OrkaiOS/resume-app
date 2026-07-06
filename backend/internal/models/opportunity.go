package models

import "time"

type Opportunity struct {
	ID          string    `json:"id"`
	Company     string    `json:"company"`
	Role        string    `json:"role"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
