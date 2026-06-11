package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/activity/entity"

	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/cryp"
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

func (r *repository) Create(ctx context.Context, event entity.ActivityEvent) error {
	if event.UserID == "" {
		return pkgErr.InvalidRequest("user_id is required")
	}
	if event.Type == "" {
		return pkgErr.InvalidRequest("type is required")
	}
	if event.ID == "" {
		event.ID = cryp.ULID()
	}
	if event.Meta == "" {
		event.Meta = "{}"
	}

	_, err := r.db.NewInsert().
		Model(&event).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) ListByUserID(ctx context.Context, userID string, limit int) ([]entity.ActivityEvent, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	events := make([]entity.ActivityEvent, 0, limit)
	err := r.db.NewSelect().
		Model(&events).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return events, nil
}
