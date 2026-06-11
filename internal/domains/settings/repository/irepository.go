package repository

import (
	"context"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"
)

type IRepository interface {
	// GetSchedule returns the config row for the given job key. A missing row
	// yields an empty struct (not an error).
	GetSchedule(ctx context.Context, jobKey string) (entity.ScheduleConfig, error)
	// UpdateSchedule persists the enabled flag, cron expression and job-specific
	// params JSON for the given job key.
	UpdateSchedule(ctx context.Context, jobKey string, enabled bool, cron string, params string, updatedBy string) error
	// RecordScheduleRun stores the outcome of a scheduler run (timestamp +
	// job-specific result JSON) for the given job key.
	RecordScheduleRun(ctx context.Context, jobKey string, ranAt time.Time, result string) error
}
