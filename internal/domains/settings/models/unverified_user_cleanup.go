package models

import (
	"errors"

	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

// UnverifiedUserMinAgeFloorDays is the youngest an unfinished account may be
// before the cleanup job deletes it. The job applies the same floor to the
// stored params, so a missing or corrupted value never widens the deletion.
const UnverifiedUserMinAgeFloorDays int64 = 1

// UnverifiedUserCleanupGetResponse is the current unverified-user-cleanup
// configuration returned to the UI.
type UnverifiedUserCleanupGetResponse struct {
	Enabled          bool   `json:"enabled"`
	Cron             string `json:"cron"`
	MinAgeDays       int64  `json:"min_age_days"`
	LastRunAt        *int64 `json:"last_run_at"`
	LastDeletedCount *int64 `json:"last_deleted_count"`
}

// UnverifiedUserCleanupUpdateRequest sets the unverified-user-cleanup schedule
// and the minimum account age (in DAYS). The cron expression is a standard
// 5-field cron (minute hour day-of-month month day-of-week).
type UnverifiedUserCleanupUpdateRequest struct {
	Enabled    bool   `json:"enabled"`
	Cron       string `json:"cron"`
	MinAgeDays int64  `json:"min_age_days"`
}

func (r UnverifiedUserCleanupUpdateRequest) Validate() error {
	if r.MinAgeDays < UnverifiedUserMinAgeFloorDays {
		return errors.New("min_age_days must be at least 1")
	}
	if r.Enabled {
		if r.Cron == "" {
			return errors.New("cron is required when the scheduler is enabled")
		}
		if err := pkgScheduler.ValidateCron(r.Cron); err != nil {
			return errors.New("cron must be a valid 5-field cron expression")
		}
	}
	return nil
}
