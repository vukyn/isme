package repository

import (
	"context"
	"database/sql"
	"errors"
	"isme/internal/domains/app_service/entity"

	pkgCtx "isme/pkg/ctx"
	pkgErr "isme/pkg/http/errors"

	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/cryp"
)

type repository struct {
	db *bun.DB
}

func NewRepository(
	db *bun.DB,
) IRepository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, req entity.CreateRequest) (string, error) {
	userID := pkgCtx.GetUserId(ctx)
	appService := &entity.AppService{
		ID:          cryp.ULID(),
		AppCode:     req.AppCode,
		AppName:     req.AppName,
		AppSecret:   req.AppSecret,
		RedirectURL: req.RedirectURL,
		CtxInfo:     req.CtxInfo,
		Status:      req.Status,
		CreatedBy:   userID,
	}
	_, err := r.db.NewInsert().
		Model(appService).
		Exec(ctx)
	if err != nil {
		return "", pkgErr.DatabaseError(err.Error())
	}
	return appService.ID, nil
}

func (r *repository) GetByCode(ctx context.Context, code string) (entity.AppService, error) {
	if code == "" {
		return entity.AppService{}, pkgErr.InvalidRequest("code is required")
	}

	appService := entity.AppService{}
	err := r.db.NewSelect().
		Model(&appService).
		Where("app_code = ?", code).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AppService{}, nil
		}
		return entity.AppService{}, pkgErr.DatabaseError(err.Error())
	}
	return appService, nil
}
