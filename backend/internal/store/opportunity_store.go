package store

import (
	"context"

	"github.com/marco/resume-app/internal/models"
)

type OpportunityStore interface {
	List(ctx context.Context, cursor string, limit int) ([]models.Opportunity, string, error)
	Get(ctx context.Context, id string) (models.Opportunity, error)
	Create(ctx context.Context, o models.Opportunity) error
	Update(ctx context.Context, o models.Opportunity) error
	Delete(ctx context.Context, id string) error
	SetArchived(ctx context.Context, id string, archived bool) error
}
