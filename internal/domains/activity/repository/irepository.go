package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/activity/entity"
)

type IRepository interface {
	// Create appends an activity event. A ULID id is generated when empty.
	Create(ctx context.Context, event entity.ActivityEvent) error
	// ListByUserID returns the most recent events for a user, newest first.
	ListByUserID(ctx context.Context, userID string, limit int) ([]entity.ActivityEvent, error)
}
