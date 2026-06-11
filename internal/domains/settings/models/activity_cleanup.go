package models

import (
	"errors"

	"github.com/robfig/cron/v3"
)

// retentionFloorDays is the minimum retention window (7 days). Activity history
// is kept on a day scale; pruning younger than a week would risk wiping the data
// the Welcome "Recent activity" feed depends on, so the floor is enforced in
// Validate.
const retentionFloorDays int64 = 7

// ActivityCleanupGetResponse is the current activity-cleanup configuration
// returned to the UI.
type ActivityCleanupGetResponse struct {
	Enabled         bool   `json:"enabled"`
	Cron            string `json:"cron"`
	RetentionDays   int64  `json:"retention_days"`
	LastRunAt       *int64 `json:"last_run_at"`
	LastPrunedCount *int64 `json:"last_pruned_count"`
}

// ActivityCleanupUpdateRequest sets the activity-cleanup schedule and retention
// window (in DAYS). The cron expression is a standard 5-field cron (minute hour
// day-of-month month day-of-week).
type ActivityCleanupUpdateRequest struct {
	Enabled       bool   `json:"enabled"`
	Cron          string `json:"cron"`
	RetentionDays int64  `json:"retention_days"`
}

func (r ActivityCleanupUpdateRequest) Validate() error {
	if r.RetentionDays < retentionFloorDays {
		return errors.New("retention_days must be at least 7")
	}
	if r.Enabled {
		if r.Cron == "" {
			return errors.New("cron is required when the scheduler is enabled")
		}
		if _, err := cron.ParseStandard(r.Cron); err != nil {
			return errors.New("cron must be a valid 5-field cron expression")
		}
	}
	return nil
}
