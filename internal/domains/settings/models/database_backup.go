package models

import (
	"errors"

	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

// retainFloorCount / retainCeilingCount bound the number of backup files kept.
// At least one backup must be retained; the ceiling caps unbounded disk growth.
const (
	retainFloorCount   int64 = 1
	retainCeilingCount int64 = 90
)

// DatabaseBackupGetResponse is the current database-backup configuration
// returned to the UI.
type DatabaseBackupGetResponse struct {
	Enabled        bool    `json:"enabled"`
	Cron           string  `json:"cron"`
	RetainCount    int64   `json:"retain_count"`
	LastRunAt      *int64  `json:"last_run_at"`
	LastBackupPath *string `json:"last_backup_path"`
	LastKeptCount  *int64  `json:"last_kept_count"`
}

// DatabaseBackupUpdateRequest sets the database-backup schedule and the number
// of backup files to retain. The cron expression is a standard 5-field cron
// (minute hour day-of-month month day-of-week).
type DatabaseBackupUpdateRequest struct {
	Enabled     bool   `json:"enabled"`
	Cron        string `json:"cron"`
	RetainCount int64  `json:"retain_count"`
}

func (r DatabaseBackupUpdateRequest) Validate() error {
	if r.RetainCount < retainFloorCount || r.RetainCount > retainCeilingCount {
		return errors.New("retain_count must be between 1 and 90")
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
