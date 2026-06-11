package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// ActivityEvent is an append-only audit record powering the Welcome "Recent
// activity" feed. Meta holds a type-specific JSON blob (e.g. {device, client_ip}
// for sign_in). The table grows unbounded — retention is deferred (see migration
// 026 for the note).
type ActivityEvent struct {
	bun.BaseModel `bun:"table:activity_events,alias:ae"`
	ID            string    `bun:"id,pk,notnull"`
	UserID        string    `bun:"user_id,notnull"`
	Type          string    `bun:"type,notnull"`
	Meta          string    `bun:"meta,notnull,default:'{}'"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp,notnull"`
}

// === Hooks ===

func (ae *ActivityEvent) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.InsertQuery); ok {
		ae.CreatedAt = time.Now().UTC()
	}
	return nil
}
