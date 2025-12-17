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
		Status: constants.UserStatusActive,
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
		Password: cryp.HashBcrypt(password, 10),
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

func (r *repository) PromoteAdmin(ctx context.Context, id string) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	user := &entity.User{
		ID:      id,
		IsAdmin: true,
	}
	_, err := r.db.NewUpdate().
		Model(user).
		Column("is_admin").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) IsAdmin(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, pkgErr.InvalidRequest("id is required")
	}

	user := entity.User{}
	err := r.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, pkgErr.InvalidRequest("user not found")
		}
		return false, pkgErr.DatabaseError(err.Error())
	}

	return user.IsAdmin, nil
}
