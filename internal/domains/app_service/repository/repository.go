package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/vukyn/isme/internal/domains/app_service/entity"

	pkgCtx "github.com/vukyn/isme/pkg/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"

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

func (r *repository) GetByID(ctx context.Context, id string) (entity.AppService, error) {
	if id == "" {
		return entity.AppService{}, pkgErr.InvalidRequest("id is required")
	}

	appService := entity.AppService{}
	err := r.db.NewSelect().
		Model(&appService).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AppService{}, nil
		}
		return entity.AppService{}, pkgErr.DatabaseError(err.Error())
	}
	return appService, nil
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

func (r *repository) Update(ctx context.Context, req entity.UpdateRequest) error {
	// validation
	if req.ID == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	appService := &entity.AppService{
		ID: req.ID,
	}
	fields := []string{}

	if req.AppSecret != nil {
		appService.AppSecret = *req.AppSecret
		fields = append(fields, "app_secret")
	}

	if len(fields) > 0 {
		userID := pkgCtx.GetUserId(ctx)
		appService.UpdatedBy = userID
		fields = append(fields, "updated_by")

		_, err := r.db.NewUpdate().
			Model(appService).
			Column(fields...).
			Where("id = ?", req.ID).
			Exec(ctx)
		if err != nil {
			return pkgErr.DatabaseError(err.Error())
		}
	}

	return nil
}
