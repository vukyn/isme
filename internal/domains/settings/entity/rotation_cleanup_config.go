package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// RotationCleanupConfig is the single-row (id=1) typed config that drives the
// token-rotation-events cleanup scheduler. The retention window (RetentionHours)
// is read fresh on every run, so a retention-only change takes effect on the
// next run without a scheduler reload.
type RotationCleanupConfig struct {
	bun.BaseModel `bun:"table:rotation_cleanup_config,alias:rcc"`

	ID               int64      `bun:"id,pk"`
	Enabled          bool       `bun:"enabled,notnull"`
	Cron             string     `bun:"cron,notnull"`
	RetentionHours   int64      `bun:"retention_hours,notnull"`
	LastRunAt        *time.Time `bun:"last_run_at"`
	LastCleanedCount *int64     `bun:"last_cleaned_count"`
	UpdatedAt        *time.Time `bun:"updated_at"`
	UpdatedBy        string     `bun:"updated_by"`
}

// === Hooks ===

func (rcc *RotationCleanupConfig) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery, *bun.UpdateQuery:
		now := time.Now().UTC()
		rcc.UpdatedAt = &now
	}
	return nil
}
