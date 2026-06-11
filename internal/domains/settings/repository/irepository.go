package repository

import (
	"context"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"
)

type IRepository interface {
	// Get returns the single (id=1) session auto-revoke config row.
	Get(ctx context.Context) (entity.SessionRevokeConfig, error)
	// Update persists the enabled flag and cron expression.
	Update(ctx context.Context, enabled bool, cron string, updatedBy string) error
	// RecordRun stores the outcome of a scheduler run (timestamp + revoked count).
	RecordRun(ctx context.Context, ranAt time.Time, revoked int64) error

	// GetRotationCleanup returns the single (id=1) rotation-cleanup config row.
	GetRotationCleanup(ctx context.Context) (entity.RotationCleanupConfig, error)
	// UpdateRotationCleanup persists the enabled flag, cron expression and retention window.
	UpdateRotationCleanup(ctx context.Context, enabled bool, cron string, retentionHours int64, updatedBy string) error
	// RecordRotationCleanupRun stores the outcome of a cleanup run (timestamp + cleaned count).
	RecordRotationCleanupRun(ctx context.Context, ranAt time.Time, cleaned int64) error
}
