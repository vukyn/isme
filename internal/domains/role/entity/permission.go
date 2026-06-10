package entity

import (
	"github.com/uptrace/bun"
)

type Permission struct {
	bun.BaseModel `bun:"table:permissions,alias:perm"`
	ID            int64  `bun:"id,pk,autoincrement"`
	AppID         string `bun:"app_id,notnull"`
	Resource      string `bun:"resource,notnull"`
	Action        string `bun:"action,notnull"`
	// Icon is a per-resource icon key (e.g. "file", "database"); empty = neutral
	// default. All rows of the same (app_id, resource) share the same icon. The
	// column is NOT NULL DEFAULT '' (no nullzero), so an empty value persists as
	// '' rather than NULL.
	Icon string `bun:"icon"`
	// Color is a per-resource color palette key (e.g. "violet"); empty = neutral
	// fallback. All rows of the same (app_id, resource) share the same color,
	// exactly like Icon. NOT NULL DEFAULT '' (no nullzero).
	Color string `bun:"color"`
}
