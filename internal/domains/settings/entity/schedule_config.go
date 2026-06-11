package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Job-key constants are the single source of truth for the scheduled-job
// identifiers. They are referenced by the migration (via a cross-package test),
// the settings repository/usecase, and the scheduler's JobKey consts, so the
// "session_revoke"/"rotation_cleanup" strings are never scattered as literals.
const (
	JobKeySessionRevoke   = "session_revoke"
	JobKeyRotationCleanup = "rotation_cleanup"
	JobKeyActivityCleanup = "activity_cleanup"
)

// ScheduleConfig is the generic, job-keyed config that drives every scheduled
// job. Shared columns (enabled/cron/run history) are typed; job-specific knobs
// live in the Params JSON blob and per-run results in the LastResult JSON blob.
// One row per job, keyed by JobKey (e.g. "session_revoke", "rotation_cleanup").
type ScheduleConfig struct {
	bun.BaseModel `bun:"table:schedule_config,alias:sc"`

	JobKey     string     `bun:"job_key,pk"`
	Enabled    bool       `bun:"enabled,notnull"`
	Cron       string     `bun:"cron,notnull"`
	Params     string     `bun:"params,notnull"`
	LastRunAt  *time.Time `bun:"last_run_at"`
	LastResult *string    `bun:"last_result"`
	UpdatedAt  *time.Time `bun:"updated_at"`
	UpdatedBy  string     `bun:"updated_by"`
}

// === Hooks ===

func (sc *ScheduleConfig) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery, *bun.UpdateQuery:
		now := time.Now().UTC()
		sc.UpdatedAt = &now
	}
	return nil
}
