package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type UserRole struct {
	bun.BaseModel `bun:"table:user_roles,alias:ur"`
	ID            string    `bun:"id,pk,notnull"`
	UserID        string    `bun:"user_id,notnull"`
	RoleID        string    `bun:"role_id,notnull"`
	AppServiceID  *string   `bun:"app_service_id"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
	CreatedBy     string    `bun:"created_by,nullzero"`
}

// === Hooks ===

func (ur *UserRole) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		ur.CreatedAt = time.Now().UTC()
	}
	return nil
}
