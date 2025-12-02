package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type UserSession struct {
	bun.BaseModel `bun:"table:user_sessions,alias:us"`
	ID            string    `bun:"id,pk,notnull"`
	UserID        string    `bun:"user_id,notnull"`
	Email         string    `bun:"email,notnull"`
	RefreshToken  string    `bun:"refresh_token"`
	ExpiresAt     time.Time `bun:"expires_at"`
	LastLoginAt   time.Time `bun:"last_login_at"`
	Status        int32     `bun:"status,default:1"`
	ClientIP      string    `bun:"client_ip,notnull"`
	UserAgent     string    `bun:"user_agent"`
	TokenID       string    `bun:"token_id,notnull"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp,notnull"`
}

// === Hooks ===

func (us *UserSession) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		us.CreatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		// No updated_at field in user_sessions table
	}
	return nil
}
