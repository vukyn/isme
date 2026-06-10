package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/settings/models"
)

type IUseCase interface {
	// Get returns the current session auto-revoke configuration.
	Get(ctx context.Context) (models.GetResponse, error)
	// Update validates and persists the schedule, then live-reloads the scheduler.
	Update(ctx context.Context, req models.UpdateRequest) error
}
