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

func (r *repository) Create(ctx context.Context, invitation entity.UserInvitation, assignments []entity.UserInvitationRole) (string, error) {
	if invitation.Email == "" {
		return "", pkgErr.InvalidRequest("email is required")
	}
	if invitation.TokenHash == "" {
		return "", pkgErr.InvalidRequest("token_hash is required")
	}
	if len(assignments) == 0 {
		return "", pkgErr.InvalidRequest("at least one role assignment is required")
	}

	invitation.ID = cryp.ULID()
	invitation.Status = int32(constants.InvitationStatusPending)

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(&invitation).Exec(ctx); err != nil {
			return err
		}

		rows := make([]entity.UserInvitationRole, 0, len(assignments))
		for _, assignment := range assignments {
			rows = append(rows, entity.UserInvitationRole{
				ID:           cryp.ULID(),
				InvitationID: invitation.ID,
				RoleID:       assignment.RoleID,
				AppServiceID: assignment.AppServiceID,
				CreatedBy:    invitation.CreatedBy,
			})
		}
		_, err := tx.NewInsert().Model(&rows).Exec(ctx)
		return err
	})
	if err != nil {
		return "", pkgErr.DatabaseError(err.Error())
	}
	return invitation.ID, nil
}

func (r *repository) GetAssignmentsByInvitationID(ctx context.Context, invitationID string) ([]entity.UserInvitationRole, error) {
	if invitationID == "" {
		return nil, pkgErr.InvalidRequest("invitation_id is required")
	}

	assignments := []entity.UserInvitationRole{}
	err := r.db.NewSelect().
		Model(&assignments).
		Where("invitation_id = ?", invitationID).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return assignments, nil
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
	invitations := []entity.UserInvitation{}
	err := r.db.NewSelect().
		Model(&invitations).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	if len(invitations) == 0 {
		return []models.InvitationListItem{}, nil
	}

	// batch-load every assignment for the page joined to its role + owning app,
	// then group by invitation_id (avoids an N+1 per invitation).
	invitationIDs := make([]string, 0, len(invitations))
	for _, invitation := range invitations {
		invitationIDs = append(invitationIDs, invitation.ID)
	}

	type assignmentRow struct {
		InvitationID string `bun:"invitation_id"`
		RoleID       string `bun:"role_id"`
		AppServiceID string `bun:"app_service_id"`
		RoleName     string `bun:"role_name"`
		RoleCode     string `bun:"role_code"`
		AppCode      string `bun:"app_code"`
		AppName      string `bun:"app_name"`
	}
	rows := []assignmentRow{}
	err = r.db.NewSelect().
		TableExpr("user_invitation_roles AS uir").
		ColumnExpr("uir.invitation_id").
		ColumnExpr("uir.role_id").
		ColumnExpr("uir.app_service_id").
		ColumnExpr("rol.name AS role_name").
		ColumnExpr("rol.code AS role_code").
		ColumnExpr("app.app_code AS app_code").
		ColumnExpr("app.app_name AS app_name").
		Join("LEFT JOIN roles AS rol ON rol.id = uir.role_id").
		Join("LEFT JOIN app_services AS app ON app.id = uir.app_service_id").
		Where("uir.invitation_id IN (?)", bun.In(invitationIDs)).
		Where("uir.deleted_at IS NULL").
		Order("uir.created_at ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	assignmentsByInvitation := map[string][]models.AssignmentItem{}
	for _, row := range rows {
		assignmentsByInvitation[row.InvitationID] = append(assignmentsByInvitation[row.InvitationID], models.AssignmentItem{
			RoleID:       row.RoleID,
			RoleName:     row.RoleName,
			RoleCode:     row.RoleCode,
			AppServiceID: row.AppServiceID,
			AppCode:      row.AppCode,
			AppName:      row.AppName,
		})
	}

	items := make([]models.InvitationListItem, 0, len(invitations))
	for _, invitation := range invitations {
		acceptedAt := ""
		if !invitation.AcceptedAt.IsZero() {
			acceptedAt = invitation.AcceptedAt.Format(time.RFC3339)
		}
		assignments := assignmentsByInvitation[invitation.ID]
		if assignments == nil {
			assignments = []models.AssignmentItem{}
		}
		items = append(items, models.InvitationListItem{
			ID:          invitation.ID,
			Email:       invitation.Email,
			Status:      invitation.Status,
			Assignments: assignments,
			ExpiresAt:   invitation.ExpiresAt.Format(time.RFC3339),
			AcceptedAt:  acceptedAt,
			CreatedAt:   invitation.CreatedAt.Format(time.RFC3339),
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
