package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"

	"github.com/uptrace/bun"
	pkgErr "github.com/vukyn/kuery/http/errors"
)

type repository struct {
	db *bun.DB
}

func NewRepository(
	db *bun.DB,
) IRepository {
	return &repository{db: db}
}

func (r *repository) GetSchedule(ctx context.Context, jobKey string) (entity.ScheduleConfig, error) {
	config := entity.ScheduleConfig{}
	err := r.db.NewSelect().
		Model(&config).
		Where("job_key = ?", jobKey).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ScheduleConfig{}, nil
		}
		return entity.ScheduleConfig{}, pkgErr.DatabaseError(err.Error())
	}
	return config, nil
}

func (r *repository) UpdateSchedule(ctx context.Context, jobKey string, enabled bool, cron string, params string, updatedBy string) error {
	config := entity.ScheduleConfig{
		JobKey:    jobKey,
		Enabled:   enabled,
		Cron:      cron,
		Params:    params,
		UpdatedBy: updatedBy,
	}

	_, err := r.db.NewUpdate().
		Model(&config).
		Column("enabled", "cron", "params", "updated_at", "updated_by").
		Where("job_key = ?", jobKey).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) RecordScheduleRun(ctx context.Context, jobKey string, ranAt time.Time, result string) error {
	config := entity.ScheduleConfig{
		JobKey:     jobKey,
		LastRunAt:  &ranAt,
		LastResult: &result,
	}

	_, err := r.db.NewUpdate().
		Model(&config).
		Column("last_run_at", "last_result", "updated_at").
		Where("job_key = ?", jobKey).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}
