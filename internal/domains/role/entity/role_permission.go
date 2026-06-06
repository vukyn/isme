package entity

import (
	"github.com/uptrace/bun"
)

type RolePermission struct {
	bun.BaseModel `bun:"table:role_permissions,alias:rp"`
	RoleID        string `bun:"role_id,pk,notnull"`
	PermissionID  int64  `bun:"permission_id,pk,notnull"`
}
