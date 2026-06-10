package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// SessionRevokeConfig is the single-row (id=1) typed config that drives the
// session auto-revoke scheduler.
type SessionRevokeConfig struct {
	bun.BaseModel `bun:"table:session_revoke_config,alias:src"`

	ID               int64      `bun:"id,pk"`
	Enabled          bool       `bun:"enabled,notnull"`
	Cron             string     `bun:"cron,notnull"`
	LastRunAt        *time.Time `bun:"last_run_at"`
	LastRevokedCount *int64     `bun:"last_revoked_count"`
	UpdatedAt        *time.Time `bun:"updated_at"`
	UpdatedBy        string     `bun:"updated_by"`
}

// === Hooks ===

func (src *SessionRevokeConfig) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery, *bun.UpdateQuery:
		now := time.Now().UTC()
		src.UpdatedAt = &now
	}
	return nil
}
