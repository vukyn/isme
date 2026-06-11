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

// configRowID is the fixed primary key of the single config row.
const configRowID int64 = 1

type repository struct {
	db *bun.DB
}

func NewRepository(
	db *bun.DB,
) IRepository {
	return &repository{db: db}
}

func (r *repository) Get(ctx context.Context) (entity.SessionRevokeConfig, error) {
	config := entity.SessionRevokeConfig{}
	err := r.db.NewSelect().
		Model(&config).
		Where("id = ?", configRowID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.SessionRevokeConfig{}, nil
		}
		return entity.SessionRevokeConfig{}, pkgErr.DatabaseError(err.Error())
	}
	return config, nil
}

func (r *repository) Update(ctx context.Context, enabled bool, cron string, updatedBy string) error {
	config := entity.SessionRevokeConfig{
		ID:        configRowID,
		Enabled:   enabled,
		Cron:      cron,
		UpdatedBy: updatedBy,
	}

	_, err := r.db.NewUpdate().
		Model(&config).
		Column("enabled", "cron", "updated_at", "updated_by").
		Where("id = ?", configRowID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) RecordRun(ctx context.Context, ranAt time.Time, revoked int64) error {
	config := entity.SessionRevokeConfig{
		ID:               configRowID,
		LastRunAt:        &ranAt,
		LastRevokedCount: &revoked,
	}

	_, err := r.db.NewUpdate().
		Model(&config).
		Column("last_run_at", "last_revoked_count", "updated_at").
		Where("id = ?", configRowID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) GetRotationCleanup(ctx context.Context) (entity.RotationCleanupConfig, error) {
	config := entity.RotationCleanupConfig{}
	err := r.db.NewSelect().
		Model(&config).
		Where("id = ?", configRowID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RotationCleanupConfig{}, nil
		}
		return entity.RotationCleanupConfig{}, pkgErr.DatabaseError(err.Error())
	}
	return config, nil
}

func (r *repository) UpdateRotationCleanup(ctx context.Context, enabled bool, cron string, retentionHours int64, updatedBy string) error {
	config := entity.RotationCleanupConfig{
		ID:             configRowID,
		Enabled:        enabled,
		Cron:           cron,
		RetentionHours: retentionHours,
		UpdatedBy:      updatedBy,
	}

	_, err := r.db.NewUpdate().
		Model(&config).
		Column("enabled", "cron", "retention_hours", "updated_at", "updated_by").
		Where("id = ?", configRowID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) RecordRotationCleanupRun(ctx context.Context, ranAt time.Time, cleaned int64) error {
	config := entity.RotationCleanupConfig{
		ID:               configRowID,
		LastRunAt:        &ranAt,
		LastCleanedCount: &cleaned,
	}

	_, err := r.db.NewUpdate().
		Model(&config).
		Column("last_run_at", "last_cleaned_count", "updated_at").
		Where("id = ?", configRowID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}
