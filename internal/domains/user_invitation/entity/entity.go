package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type UserInvitation struct {
	bun.BaseModel `bun:"table:user_invitations,alias:uin"`
	ID            string    `bun:"id,pk,notnull"`
	Email         string    `bun:"email,notnull"`
	RoleID        string    `bun:"role_id,notnull"`
	TokenHash     string    `bun:"token_hash,notnull"`
	Status        int32     `bun:"status,notnull,default:1"`
	ExpiresAt     time.Time `bun:"expires_at,notnull"`
	AcceptedAt    time.Time `bun:"accepted_at,nullzero"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
	CreatedBy     string    `bun:"created_by,nullzero"`
	UpdatedAt     time.Time `bun:"updated_at,default:current_timestamp"`
	UpdatedBy     string    `bun:"updated_by,nullzero"`
	DeletedAt     time.Time `bun:"deleted_at,soft_delete,nullzero"`
	DeletedBy     string    `bun:"deleted_by,nullzero"`
}

// === Hooks ===

func (u *UserInvitation) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch q := query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		q.Column("updated_at")
		u.UpdatedAt = time.Now().UTC()
	}
	return nil
}
