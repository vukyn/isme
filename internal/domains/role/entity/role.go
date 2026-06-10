package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:roles,alias:rol"`
	ID            string `bun:"id,pk,notnull"`
	AppID         string `bun:"app_id,notnull"`
	Code          string `bun:"code,notnull"`
	Name          string `bun:"name,notnull"`
	Description   string `bun:"description"`
	// Icon is an icon key (allowlist in role/constants ICON_KEYS); empty =
	// neutral default in the UI. NOT NULL DEFAULT '' (no nullzero) so an empty
	// value persists as '' rather than NULL.
	Icon string `bun:"icon"`
	// Color is a color palette key (allowlist in role/constants COLOR_KEYS);
	// empty = neutral fallback in the UI. NOT NULL DEFAULT '' (no nullzero).
	Color     string    `bun:"color"`
	IsSystem  bool      `bun:"is_system,default:false"`
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
	CreatedBy string    `bun:"created_by,nullzero"`
	UpdatedAt time.Time `bun:"updated_at,default:current_timestamp"`
	UpdatedBy string    `bun:"updated_by,nullzero"`
	DeletedAt time.Time `bun:"deleted_at,soft_delete,nullzero"`
	DeletedBy string    `bun:"deleted_by,nullzero"`
}

// === Hooks ===

func (r *Role) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch q := query.(type) {
	case *bun.InsertQuery:
		r.CreatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		q.Column("updated_at")
		r.UpdatedAt = time.Now().UTC()
	}
	return nil
}
