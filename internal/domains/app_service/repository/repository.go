package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/app_service/models"

	pkgCtx "github.com/vukyn/kuery/ctx"
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
	userID := pkgCtx.GetUserID(ctx)
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

func (r *repository) GetByIDs(ctx context.Context, ids []string) (map[string]entity.AppService, error) {
	appsByID := map[string]entity.AppService{}
	if len(ids) == 0 {
		return appsByID, nil
	}

	appServices := []entity.AppService{}
	err := r.db.NewSelect().
		Model(&appServices).
		Where("id IN (?)", bun.In(ids)).
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	for _, appService := range appServices {
		appsByID[appService.ID] = appService
	}
	return appsByID, nil
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
		userID := pkgCtx.GetUserID(ctx)
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

func (r *repository) List(ctx context.Context, req models.ListRequest) ([]entity.AppService, int64, error) {
	query := r.db.NewSelect().
		Model((*entity.AppService)(nil))

	// Apply search filter (LOWER + LIKE for SQLite compatibility)
	if req.Search != "" {
		pattern := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("(LOWER(app_name) LIKE ? OR LOWER(app_code) LIKE ?)", pattern, pattern)
	}

	// Apply status filter (only if status is set and non-zero)
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}

	// Apply ctx_info filter
	if req.CtxInfo != "" {
		query = query.Where("ctx_info = ?", req.CtxInfo)
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, pkgErr.DatabaseError(err.Error())
	}

	// Apply pagination
	appServices := make([]entity.AppService, 0)
	offset := (req.Page - 1) * req.PageSize
	err = query.
		Offset(offset).
		Limit(req.PageSize).
		Order("created_at DESC").
		Scan(ctx, &appServices)
	if err != nil {
		return nil, 0, pkgErr.DatabaseError(err.Error())
	}

	return appServices, int64(total), nil
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status int32) error {
	// validation
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	userID := pkgCtx.GetUserID(ctx)
	appService := &entity.AppService{
		ID:        id,
		Status:    status,
		UpdatedBy: userID,
	}
	_, err := r.db.NewUpdate().
		Model(appService).
		Column("status", "updated_by").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}
