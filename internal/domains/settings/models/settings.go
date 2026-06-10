package models

import (
	"errors"

	"github.com/robfig/cron/v3"
)

// GetResponse is the current session auto-revoke configuration returned to the UI.
type GetResponse struct {
	Enabled          bool   `json:"enabled"`
	Cron             string `json:"cron"`
	LastRunAt        *int64 `json:"last_run_at"`
	LastRevokedCount *int64 `json:"last_revoked_count"`
}

// UpdateRequest sets the session auto-revoke schedule. The cron expression is a
// standard 5-field cron (minute hour day-of-month month day-of-week).
type UpdateRequest struct {
	Enabled bool   `json:"enabled"`
	Cron    string `json:"cron"`
}

func (r UpdateRequest) Validate() error {
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
