package repository

import (
	"context"
	"time"

	"github.com/vukyn/isme/internal/domains/activity/entity"
)

type IRepository interface {
	// Create appends an activity event. A ULID id is generated when empty.
	Create(ctx context.Context, event entity.ActivityEvent) error
	// ListByUserID returns the most recent events for a user, newest first.
	ListByUserID(ctx context.Context, userID string, limit int) ([]entity.ActivityEvent, error)
	// PruneBefore deletes activity events created before the given time and
	// returns the number of rows removed. Driven by the activity-cleanup
	// scheduler to keep activity_events bounded.
	PruneBefore(ctx context.Context, before time.Time) (int64, error)
}
