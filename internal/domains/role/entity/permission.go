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
	// default. All rows of the same (app_id, resource) share the same icon.
	Icon string `bun:"icon,nullzero"`
}
