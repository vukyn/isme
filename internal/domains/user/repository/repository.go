package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vukyn/isme/internal/domains/user/constants"
	"github.com/vukyn/isme/internal/domains/user/entity"
	"github.com/vukyn/isme/internal/domains/user/models"

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

func (r *repository) Create(ctx context.Context, req models.CreateRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", pkgErr.InvalidRequest(err.Error())
	}

	user := &entity.User{
		ID:     cryp.ULID(),
		Name:   req.Name,
		Email:  req.Email,
		Status: int32(constants.UserStatusActive),
	}
	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)
	if err != nil {
		return "", pkgErr.DatabaseError(err.Error())
	}
	return user.ID, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (entity.User, error) {
	if id == "" {
		return entity.User{}, pkgErr.InvalidRequest("id is required")
	}

	user := entity.User{}
	err := r.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return entity.User{}, pkgErr.DatabaseError(err.Error())
	}
	return user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	if email == "" {
		return entity.User{}, pkgErr.InvalidRequest("email is required")
	}

	user := entity.User{}
	err := r.db.NewSelect().
		Model(&user).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, nil
		}
		return entity.User{}, pkgErr.DatabaseError(err.Error())
	}
	return user, nil
}

func (r *repository) SetPassword(ctx context.Context, id string, password string) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}
	if password == "" {
		return pkgErr.InvalidRequest("password is required")
	}

	user := &entity.User{
		ID:       id,
		Password: cryp.HashArgon2id(password),
	}
	columns := []string{"password"}
	_, err := r.db.NewUpdate().
		Model(user).
		Column(columns...).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) UpdateLastLogin(ctx context.Context, id string) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	user := &entity.User{
		ID:          id,
		LastLoginAt: time.Now().UTC(),
	}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("last_login_at").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) List(ctx context.Context, req models.ListRequest) ([]entity.User, int64, error) {
	// validation
	if err := req.Validate(); err != nil {
		return nil, 0, pkgErr.InvalidRequest(err.Error())
	}

	query := r.db.NewSelect().
		Model((*entity.User)(nil)).
		Where("deleted_at IS NULL")

	// Apply search filter
	if req.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// Apply status filter (only if status is set and non-zero)
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}

	// Apply verified filter
	if req.Verified != nil {
		query = query.Where("is_verified = ?", *req.Verified)
	}

	// Apply app + role filter (app-owned RBAC).
	// Role codes are not unique across apps, so role filtering is only meaningful
	// when scoped to an app — mirror the UI contract: ignore the role predicate
	// unless an app is also chosen. Restrict to users holding a matching role via
	// the user_roles → roles → app_services chain, keeping soft-delete semantics
	// on roles and matching the assignment scope to the owning role's app.
	if req.AppCode != "" {
		subQuery := r.db.NewSelect().
			TableExpr("user_roles AS ur").
			ColumnExpr("ur.user_id").
			Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
			Join("JOIN app_services AS app ON app.id = rol.app_id").
			Where("ur.app_service_id = rol.app_id").
			Where("app.app_code = ?", req.AppCode)
		if req.RoleID != "" {
			subQuery = subQuery.Where("rol.code = ?", req.RoleID)
		}
		query = query.Where("id IN (?)", subQuery)
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, pkgErr.DatabaseError(err.Error())
	}

	// Apply pagination
	users := make([]entity.User, 0)
	offset := (req.Page - 1) * req.PageSize
	err = query.
		Offset(offset).
		Limit(req.PageSize).
		Order("created_at DESC").
		Scan(ctx, &users)
	if err != nil {
		return nil, 0, pkgErr.DatabaseError(err.Error())
	}

	return users, int64(total), nil
}

func (r *repository) Verify(ctx context.Context, id string) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	user := &entity.User{
		ID:         id,
		IsVerified: true,
	}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("is_verified").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status int32) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}
	if status != 1 && status != 2 {
		return pkgErr.InvalidRequest("invalid status, must be 1 (active) or 2 (inactive)")
	}

	user := &entity.User{
		ID:     id,
		Status: status,
	}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("status").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) SoftDelete(ctx context.Context, id string) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	user := &entity.User{
		ID:        id,
		DeletedAt: time.Now().UTC(),
	}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("deleted_at").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}
