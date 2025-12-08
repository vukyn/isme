package entity

import (
	"github.com/uptrace/bun"
)

type AppService struct {
	bun.BaseModel `bun:"table:app_services,alias:app"`
	ID            string `bun:"id,pk,notnull"`
	AppCode       string `bun:"app_code,unique,notnull"`
	AppName       string `bun:"app_name,notnull"`
	AppSecret     string `bun:"app_secret,notnull"`
	RedirectURL   string `bun:"redirect_url,notnull"`
	CtxInfo       string `bun:"ctx_info,notnull"`
	Status        int32  `bun:"status,notnull"`
}

type CreateRequest struct {
	AppCode     string
	AppName     string
	AppSecret   string
	RedirectURL string
	CtxInfo     string
	Status      int32
}
