package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type AppService struct {
	bun.BaseModel `bun:"table:app_services,alias:app"`
	ID            string    `bun:"id,pk,notnull"`
	AppCode       string    `bun:"app_code,unique,notnull"`
	AppName       string    `bun:"app_name,notnull"`
	AppSecret     string    `bun:"app_secret,notnull"`
	RedirectURL   string    `bun:"redirect_url,notnull"`
	CtxInfo       string    `bun:"ctx_info,notnull"`
	Status        int32     `bun:"status,notnull"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
	CreatedBy     string    `bun:"created_by,nullzero"`
	UpdatedAt     time.Time `bun:"updated_at,default:current_timestamp"`
	UpdatedBy     string    `bun:"updated_by,nullzero"`
	DeletedAt     time.Time `bun:"deleted_at,soft_delete,nullzero"`
	DeletedBy     string    `bun:"deleted_by,nullzero"`
}

type CreateRequest struct {
	AppCode     string
	AppName     string
	AppSecret   string
	RedirectURL string
	CtxInfo     string
	Status      int32
}

// === Hooks ===

func (a *AppService) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch q := query.(type) {
	case *bun.InsertQuery:
		a.CreatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		q.Column("updated_at")
		a.UpdatedAt = time.Now().UTC()
	}
	return nil
}
