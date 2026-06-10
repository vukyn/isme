package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vukyn/isme/internal/domains/user_session/constants"
	"github.com/vukyn/isme/internal/domains/user_session/entity"
	"github.com/vukyn/isme/internal/domains/user_session/models"

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

func (r *repository) Create(ctx context.Context, req models.CreateRequest) (entity.UserSession, error) {
	if err := req.Validate(); err != nil {
		return entity.UserSession{}, pkgErr.InvalidRequest(err.Error())
	}

	userSession := entity.UserSession{
		ID:           cryp.ULID(),
		Status:       constants.UserSessionStatusActive,
		UserID:       req.UserID,
		Email:        req.Email,
		RefreshToken: cryp.HashSHA256(req.RefreshToken),
		ExpiresAt:    req.ExpiresAt,
		LastLoginAt:  time.Now(),
		ClientIP:     req.ClientIP,
		UserAgent:    req.UserAgent,
		TokenID:      req.TokenID,
		AppServiceID: req.AppServiceID,
	}

	_, err := r.db.NewInsert().
		Model(&userSession).
		Exec(ctx)
	if err != nil {
		return entity.UserSession{}, pkgErr.DatabaseError(err.Error())
	}
	return userSession, nil
}

func (r *repository) UpdateLastLogin(ctx context.Context, req models.UpdateLastLoginRequest) error {
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	userSession := entity.UserSession{
		ID:           req.ID,
		LastLoginAt:  time.Now(),
		ClientIP:     req.ClientIP,
		UserAgent:    req.UserAgent,
		TokenID:      req.TokenID,
		RefreshToken: cryp.HashSHA256(req.RefreshToken),
		ExpiresAt:    req.ExpiresAt,
	}

	_, err := r.db.NewUpdate().
		Model(&userSession).
		Column("last_login_at", "refresh_token", "client_ip", "user_agent", "expires_at", "token_id").
		Where("id = ?", req.ID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) InactiveAllUserSession(ctx context.Context, userID string) error {
	if userID == "" {
		return pkgErr.InvalidRequest("user_id is required")
	}

	userSession := entity.UserSession{
		Status: constants.UserSessionStatusInactive,
	}

	_, err := r.db.NewUpdate().
		Model(&userSession).
		Column("status").
		Where("user_id = ?", userID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) InactiveSessionByTokenID(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return pkgErr.InvalidRequest("token_id is required")
	}

	userSession := entity.UserSession{
		Status: constants.UserSessionStatusInactive,
	}

	_, err := r.db.NewUpdate().
		Model(&userSession).
		Column("status").
		Where("token_id = ?", tokenID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) FindByRefreshToken(ctx context.Context, refreshToken string) (entity.UserSession, error) {
	if refreshToken == "" {
		return entity.UserSession{}, pkgErr.InvalidRequest("refresh_token is required")
	}

	userSession := entity.UserSession{}
	err := r.db.NewSelect().
		Model(&userSession).
		Where("refresh_token = ?", cryp.HashSHA256(refreshToken)).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserSession{}, nil
		}
		return entity.UserSession{}, pkgErr.DatabaseError(err.Error())
	}
	return userSession, nil
}

func (r *repository) FindByTokenID(ctx context.Context, tokenID string) (entity.UserSession, error) {
	if tokenID == "" {
		return entity.UserSession{}, pkgErr.InvalidRequest("token_id is required")
	}

	userSession := entity.UserSession{}
	err := r.db.NewSelect().
		Model(&userSession).
		Where("token_id = ?", tokenID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserSession{}, nil
		}
		return entity.UserSession{}, pkgErr.DatabaseError(err.Error())
	}
	return userSession, nil
}

func (r *repository) GetListActiveByUserID(ctx context.Context, userID string) ([]entity.UserSession, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	var userSessions []entity.UserSession
	err := r.db.NewSelect().
		Model(&userSessions).
		Where("user_id = ? AND status = ?", userID, constants.UserSessionStatusActive).
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return userSessions, nil
}

func (r *repository) CountActiveByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error) {
	if len(userIDs) == 0 {
		return make(map[string]int), nil
	}

	var results []struct {
		UserID string `bun:"user_id"`
		Count  int    `bun:"count"`
	}

	err := r.db.NewSelect().
		Model((*entity.UserSession)(nil)).
		Column("user_id").
		ColumnExpr("COUNT(*) as count").
		Where("user_id IN (?)", bun.In(userIDs)).
		Where("status = ?", constants.UserSessionStatusActive).
		Group("user_id").
		Scan(ctx, &results)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	counts := make(map[string]int)
	for _, userID := range userIDs {
		counts[userID] = 0
	}
	for _, result := range results {
		counts[result.UserID] = result.Count
	}
	return counts, nil
}

func (r *repository) GetByID(ctx context.Context, sessionID string) (entity.UserSession, error) {
	if sessionID == "" {
		return entity.UserSession{}, pkgErr.InvalidRequest("session_id is required")
	}

	userSession := entity.UserSession{}
	err := r.db.NewSelect().
		Model(&userSession).
		Where("id = ?", sessionID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserSession{}, nil
		}
		return entity.UserSession{}, pkgErr.DatabaseError(err.Error())
	}
	return userSession, nil
}

func (r *repository) InactiveExpiredSessions(ctx context.Context, before time.Time) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*entity.UserSession)(nil)).
		Set("status = ?", constants.UserSessionStatusInactive).
		Where("status = ?", constants.UserSessionStatusActive).
		Where("expires_at < ?", before).
		Exec(ctx)
	if err != nil {
		return 0, pkgErr.DatabaseError(err.Error())
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, pkgErr.DatabaseError(err.Error())
	}
	return count, nil
}

func (r *repository) InactiveSessionByID(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return pkgErr.InvalidRequest("session_id is required")
	}

	userSession := entity.UserSession{
		Status: constants.UserSessionStatusInactive,
	}

	_, err := r.db.NewUpdate().
		Model(&userSession).
		Column("status").
		Where("id = ?", sessionID).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}
