package models

import (
	"errors"

	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

// InvitationCleanupGetResponse is the current invitation-cleanup configuration
// returned to the UI.
type InvitationCleanupGetResponse struct {
	Enabled          bool   `json:"enabled"`
	Cron             string `json:"cron"`
	LastRunAt        *int64 `json:"last_run_at"`
	LastExpiredCount *int64 `json:"last_expired_count"`
}

// InvitationCleanupUpdateRequest sets the invitation-cleanup schedule. The cron
// expression is a standard 5-field cron (minute hour day-of-month month day-of-week).
type InvitationCleanupUpdateRequest struct {
	Enabled bool   `json:"enabled"`
	Cron    string `json:"cron"`
}

func (r InvitationCleanupUpdateRequest) Validate() error {
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
