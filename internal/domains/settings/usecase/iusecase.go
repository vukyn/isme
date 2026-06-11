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
	// GetRotationCleanup returns the current rotation-cleanup configuration.
	GetRotationCleanup(ctx context.Context) (models.RotationCleanupGetResponse, error)
	// UpdateRotationCleanup validates and persists the cleanup schedule + retention,
	// then live-reloads the scheduler.
	UpdateRotationCleanup(ctx context.Context, req models.RotationCleanupUpdateRequest) error
}
