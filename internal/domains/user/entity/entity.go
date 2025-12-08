package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:usr"`
	ID            string    `bun:"id,pk,notnull"`
	Name          string    `bun:"name,notnull"`
	Email         string    `bun:"email,unique"`
	Password      string    `bun:"password,notnull"`
	Status        int32     `bun:"status,notnull"`
	IsAdmin       bool      `bun:"is_admin,default:false"`
	LastLoginAt   time.Time `bun:"last_login_at,nullzero"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
	CreatedBy     int64     `bun:"created_by,nullzero"`
	UpdatedAt     time.Time `bun:"updated_at,default:current_timestamp"`
	UpdatedBy     int64     `bun:"updated_by,nullzero"`
	DeletedAt     time.Time `bun:"deleted_at,soft_delete,nullzero"`
	DeletedBy     int64     `bun:"deleted_by,nullzero"`
}

// === Hooks ===

func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch q := query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		q.Column("updated_at")
		u.UpdatedAt = time.Now().UTC()
	}
	return nil
}
