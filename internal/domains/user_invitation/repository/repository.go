package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vukyn/isme/internal/domains/user_invitation/constants"
	"github.com/vukyn/isme/internal/domains/user_invitation/entity"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"

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

func (r *repository) Create(ctx context.Context, invitation entity.UserInvitation) (string, error) {
	if invitation.Email == "" {
		return "", pkgErr.InvalidRequest("email is required")
	}
	if invitation.RoleID == "" {
		return "", pkgErr.InvalidRequest("role_id is required")
	}
	if invitation.TokenHash == "" {
		return "", pkgErr.InvalidRequest("token_hash is required")
	}

	invitation.ID = cryp.ULID()
	invitation.Status = int32(constants.InvitationStatusPending)
	_, err := r.db.NewInsert().
		Model(&invitation).
		Exec(ctx)
	if err != nil {
		return "", pkgErr.DatabaseError(err.Error())
	}
	return invitation.ID, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (entity.UserInvitation, error) {
	if id == "" {
		return entity.UserInvitation{}, pkgErr.InvalidRequest("id is required")
	}

	invitation := entity.UserInvitation{}
	err := r.db.NewSelect().
		Model(&invitation).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserInvitation{}, nil
		}
		return entity.UserInvitation{}, pkgErr.DatabaseError(err.Error())
	}
	return invitation, nil
}

func (r *repository) GetByTokenHash(ctx context.Context, tokenHash string) (entity.UserInvitation, error) {
	if tokenHash == "" {
		return entity.UserInvitation{}, pkgErr.InvalidRequest("token_hash is required")
	}

	invitation := entity.UserInvitation{}
	err := r.db.NewSelect().
		Model(&invitation).
		Where("token_hash = ?", tokenHash).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserInvitation{}, nil
		}
		return entity.UserInvitation{}, pkgErr.DatabaseError(err.Error())
	}
	return invitation, nil
}

func (r *repository) GetPendingByEmail(ctx context.Context, email string) (entity.UserInvitation, error) {
	if email == "" {
		return entity.UserInvitation{}, pkgErr.InvalidRequest("email is required")
	}

	invitation := entity.UserInvitation{}
	err := r.db.NewSelect().
		Model(&invitation).
		Where("email = ?", email).
		Where("status = ?", constants.InvitationStatusPending).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserInvitation{}, nil
		}
		return entity.UserInvitation{}, pkgErr.DatabaseError(err.Error())
	}
	return invitation, nil
}

func (r *repository) List(ctx context.Context) ([]models.InvitationListItem, error) {
	type invitationListRow struct {
		entity.UserInvitation `bun:",extend"`
		RoleName              string `bun:"role_name,scanonly"`
	}

	rows := []invitationListRow{}
	err := r.db.NewSelect().
		Model(&rows).
		ColumnExpr("uin.*").
		ColumnExpr("rol.name AS role_name").
		Join("LEFT JOIN roles AS rol ON rol.id = uin.role_id").
		Order("uin.created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	items := make([]models.InvitationListItem, 0, len(rows))
	for _, row := range rows {
		acceptedAt := ""
		if !row.AcceptedAt.IsZero() {
			acceptedAt = row.AcceptedAt.Format(time.RFC3339)
		}
		items = append(items, models.InvitationListItem{
			ID:         row.ID,
			Email:      row.Email,
			RoleID:     row.RoleID,
			RoleName:   row.RoleName,
			Status:     row.Status,
			ExpiresAt:  row.ExpiresAt.Format(time.RFC3339),
			AcceptedAt: acceptedAt,
			CreatedAt:  row.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, nil
}

func (r *repository) MarkAccepted(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, pkgErr.InvalidRequest("id is required")
	}

	now := time.Now().UTC()
	result, err := r.db.NewUpdate().
		Model((*entity.UserInvitation)(nil)).
		Set("status = ?", constants.InvitationStatusAccepted).
		Set("accepted_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("status = ?", constants.InvitationStatusPending).
		Exec(ctx)
	if err != nil {
		return false, pkgErr.DatabaseError(err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, pkgErr.DatabaseError(err.Error())
	}
	return rowsAffected > 0, nil
}

func (r *repository) MarkRevoked(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, pkgErr.InvalidRequest("id is required")
	}

	now := time.Now().UTC()
	result, err := r.db.NewUpdate().
		Model((*entity.UserInvitation)(nil)).
		Set("status = ?", constants.InvitationStatusRevoked).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("status = ?", constants.InvitationStatusPending).
		Exec(ctx)
	if err != nil {
		return false, pkgErr.DatabaseError(err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, pkgErr.DatabaseError(err.Error())
	}
	return rowsAffected > 0, nil
}

func (r *repository) RevertToPending(ctx context.Context, id string) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}

	now := time.Now().UTC()
	_, err := r.db.NewUpdate().
		Model((*entity.UserInvitation)(nil)).
		Set("status = ?", constants.InvitationStatusPending).
		Set("accepted_at = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}
