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
	// GetActivityCleanup returns the current activity-cleanup configuration.
	GetActivityCleanup(ctx context.Context) (models.ActivityCleanupGetResponse, error)
	// UpdateActivityCleanup validates and persists the cleanup schedule + retention
	// (in days), then live-reloads the scheduler.
	UpdateActivityCleanup(ctx context.Context, req models.ActivityCleanupUpdateRequest) error
	// GetDatabaseBackup returns the current database-backup configuration.
	GetDatabaseBackup(ctx context.Context) (models.DatabaseBackupGetResponse, error)
	// UpdateDatabaseBackup validates and persists the backup schedule + retain
	// count, then live-reloads the scheduler.
	UpdateDatabaseBackup(ctx context.Context, req models.DatabaseBackupUpdateRequest) error
	// GetInvitationCleanup returns the current invitation-cleanup configuration.
	GetInvitationCleanup(ctx context.Context) (models.InvitationCleanupGetResponse, error)
	// UpdateInvitationCleanup validates and persists the cleanup schedule, then
	// live-reloads the scheduler.
	UpdateInvitationCleanup(ctx context.Context, req models.InvitationCleanupUpdateRequest) error
	// GetUnverifiedUserCleanup returns the current unverified-user-cleanup configuration.
	GetUnverifiedUserCleanup(ctx context.Context) (models.UnverifiedUserCleanupGetResponse, error)
	// UpdateUnverifiedUserCleanup validates and persists the cleanup schedule +
	// minimum account age (in days), then live-reloads the scheduler.
	UpdateUnverifiedUserCleanup(ctx context.Context, req models.UnverifiedUserCleanupUpdateRequest) error
}
