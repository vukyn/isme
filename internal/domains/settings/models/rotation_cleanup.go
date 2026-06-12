package models

import (
	"errors"

	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

// retentionFloorHours is the minimum retention window (24h). Pruning rotation
// events younger than a day would risk wiping the data the Welcome 24h "Token
// rotations" card depends on, so the floor is enforced in Validate.
const retentionFloorHours int64 = 24

// RotationCleanupGetResponse is the current rotation-cleanup configuration
// returned to the UI.
type RotationCleanupGetResponse struct {
	Enabled          bool   `json:"enabled"`
	Cron             string `json:"cron"`
	RetentionHours   int64  `json:"retention_hours"`
	LastRunAt        *int64 `json:"last_run_at"`
	LastCleanedCount *int64 `json:"last_cleaned_count"`
}

// RotationCleanupUpdateRequest sets the rotation-cleanup schedule and retention
// window. The cron expression is a standard 5-field cron (minute hour
// day-of-month month day-of-week).
type RotationCleanupUpdateRequest struct {
	Enabled        bool   `json:"enabled"`
	Cron           string `json:"cron"`
	RetentionHours int64  `json:"retention_hours"`
}

func (r RotationCleanupUpdateRequest) Validate() error {
	if r.RetentionHours < retentionFloorHours {
		return errors.New("retention_hours must be at least 24")
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
