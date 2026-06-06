package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vukyn/isme/internal/domains/role/entity"
	"github.com/vukyn/isme/internal/domains/role/models"

	pkgBunQuery "github.com/vukyn/kuery/bun/query"
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

func (r *repository) Create(ctx context.Context, req models.CreateRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", pkgErr.InvalidRequest(err.Error())
	}

	role := &entity.Role{
		ID:          cryp.ULID(),
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   pkgCtx.GetUserID(ctx),
	}
	_, err := r.db.NewInsert().
		Model(role).
		Exec(ctx)
	if err != nil {
		return "", pkgErr.DatabaseError(err.Error())
	}
	return role.ID, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (entity.Role, error) {
	if id == "" {
		return entity.Role{}, pkgErr.InvalidRequest("id is required")
	}

	role := entity.Role{}
	err := r.db.NewSelect().
		Model(&role).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Role{}, nil
		}
		return entity.Role{}, pkgErr.DatabaseError(err.Error())
	}
	return role, nil
}

func (r *repository) GetByCode(ctx context.Context, code string) (entity.Role, error) {
	if code == "" {
		return entity.Role{}, pkgErr.InvalidRequest("code is required")
	}

	role := entity.Role{}
	err := r.db.NewSelect().
		Model(&role).
		Where("code = ?", code).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Role{}, nil
		}
		return entity.Role{}, pkgErr.DatabaseError(err.Error())
	}
	return role, nil
}

func (r *repository) List(ctx context.Context) ([]models.RoleListItem, error) {
	type roleListRow struct {
		entity.Role  `bun:",extend"`
		MembersCount int `bun:"members_count,scanonly"`
	}

	rows := []roleListRow{}
	err := r.db.NewSelect().
		Model(&rows).
		ColumnExpr("rol.*").
		ColumnExpr("(SELECT COUNT(*) FROM user_roles ur WHERE ur.role_id = rol.id) AS members_count").
		Order("rol.created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	items := make([]models.RoleListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.RoleListItem{
			ID:           row.ID,
			Code:         row.Code,
			Name:         row.Name,
			Description:  row.Description,
			IsSystem:     row.IsSystem,
			MembersCount: row.MembersCount,
		})
	}
	return items, nil
}

func (r *repository) Update(ctx context.Context, id string, req models.UpdateRequest) error {
	if id == "" {
		return pkgErr.InvalidRequest("id is required")
	}
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	role := &entity.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   pkgCtx.GetUserID(ctx),
	}
	_, err := r.db.NewUpdate().
		Model(role).
		Column("name", "description", "updated_by").
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

	role := &entity.Role{
		ID:        id,
		DeletedAt: time.Now().UTC(),
		DeletedBy: pkgCtx.GetUserID(ctx),
	}
	_, err := r.db.NewUpdate().
		Model(role).
		Column("deleted_at", "deleted_by").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	permissions := []entity.Permission{}
	err := r.db.NewSelect().
		Model(&permissions).
		Order("id ASC").
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return permissions, nil
}

func (r *repository) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]entity.Permission, error) {
	if roleID == "" {
		return nil, pkgErr.InvalidRequest("role_id is required")
	}

	permissions := []entity.Permission{}
	err := r.db.NewSelect().
		Model(&permissions).
		Join("JOIN role_permissions AS rp ON rp.permission_id = perm.id").
		Where("rp.role_id = ?", roleID).
		Order("perm.id ASC").
		Scan(ctx)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return permissions, nil
}

func (r *repository) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []int64) error {
	if roleID == "" {
		return pkgErr.InvalidRequest("role_id is required")
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*entity.RolePermission)(nil)).
			Where("role_id = ?", roleID).
			Exec(ctx)
		if err != nil {
			return err
		}

		if len(permissionIDs) == 0 {
			return nil
		}

		rolePermissions := make([]entity.RolePermission, 0, len(permissionIDs))
		for _, permissionID := range permissionIDs {
			rolePermissions = append(rolePermissions, entity.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			})
		}
		_, err = tx.NewInsert().
			Model(&rolePermissions).
			Exec(ctx)
		return err
	})
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) ListMembers(ctx context.Context, roleID string, req models.ListMembersRequest) ([]models.MemberItem, int, error) {
	if roleID == "" {
		return nil, 0, pkgErr.InvalidRequest("role_id is required")
	}

	type memberRow struct {
		UserID       string    `bun:"user_id"`
		Name         string    `bun:"name"`
		Email        string    `bun:"email"`
		AppServiceID *string   `bun:"app_service_id"`
		CreatedAt    time.Time `bun:"created_at"`
	}

	buildQuery := func() *bun.SelectQuery {
		query := r.db.NewSelect().
			TableExpr("user_roles AS ur").
			Join("JOIN users AS usr ON usr.id = ur.user_id AND usr.deleted_at IS NULL").
			Where("ur.role_id = ?", roleID)
		if req.Query != "" {
			search := "%" + req.Query + "%"
			query = query.Where("(usr.name LIKE ? OR usr.email LIKE ?)", search, search)
		}
		return query
	}

	total, err := buildQuery().Count(ctx)
	if err != nil {
		return nil, 0, pkgErr.DatabaseError(err.Error())
	}

	rows := []memberRow{}
	query := buildQuery().
		ColumnExpr("ur.user_id").
		ColumnExpr("usr.name").
		ColumnExpr("usr.email").
		ColumnExpr("ur.app_service_id").
		ColumnExpr("ur.created_at")
	query = pkgBunQuery.SelectWithPagination(query, req.Pagination, "ur.created_at DESC")
	if err := query.Scan(ctx, &rows); err != nil {
		return nil, 0, pkgErr.DatabaseError(err.Error())
	}

	items := make([]models.MemberItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.MemberItem{
			UserID:       row.UserID,
			Name:         row.Name,
			Email:        row.Email,
			AppServiceID: row.AppServiceID,
			CreatedAt:    row.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, total, nil
}

func (r *repository) CountMembersByRoleID(ctx context.Context, roleID string) (int, error) {
	if roleID == "" {
		return 0, pkgErr.InvalidRequest("role_id is required")
	}

	count, err := r.db.NewSelect().
		Model((*entity.UserRole)(nil)).
		Where("role_id = ?", roleID).
		Count(ctx)
	if err != nil {
		return 0, pkgErr.DatabaseError(err.Error())
	}
	return count, nil
}

func (r *repository) AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error {
	if roleID == "" {
		return pkgErr.InvalidRequest("role_id is required")
	}
	if len(userIDs) == 0 {
		return pkgErr.InvalidRequest("user_ids is required")
	}

	createdBy := pkgCtx.GetUserID(ctx)
	userRoles := make([]entity.UserRole, 0, len(userIDs))
	for _, userID := range userIDs {
		userRoles = append(userRoles, entity.UserRole{
			ID:           cryp.ULID(),
			UserID:       userID,
			RoleID:       roleID,
			AppServiceID: appServiceID,
			CreatedBy:    createdBy,
		})
	}

	_, err := r.db.NewInsert().
		Model(&userRoles).
		Ignore().
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) RemoveMember(ctx context.Context, roleID string, userID string, appServiceID *string) error {
	if roleID == "" {
		return pkgErr.InvalidRequest("role_id is required")
	}
	if userID == "" {
		return pkgErr.InvalidRequest("user_id is required")
	}

	query := r.db.NewDelete().
		Model((*entity.UserRole)(nil)).
		Where("role_id = ?", roleID).
		Where("user_id = ?", userID)
	if appServiceID == nil {
		query = query.Where("app_service_id IS NULL")
	} else {
		query = query.Where("app_service_id = ?", *appServiceID)
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

func (r *repository) GetPermissionCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	query := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("DISTINCT perm.resource || ':' || perm.action").
		Join("JOIN role_permissions AS rp ON rp.role_id = ur.role_id").
		Join("JOIN permissions AS perm ON perm.id = rp.permission_id").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Where("ur.user_id = ?", userID)
	if appServiceID == "" {
		query = query.Where("ur.app_service_id IS NULL")
	} else {
		query = query.Where("(ur.app_service_id IS NULL OR ur.app_service_id = ?)", appServiceID)
	}

	codes := []string{}
	if err := query.Scan(ctx, &codes); err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return codes, nil
}

func (r *repository) GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	query := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("DISTINCT rol.code").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Where("ur.user_id = ?", userID)
	if appServiceID == "" {
		query = query.Where("ur.app_service_id IS NULL")
	} else {
		query = query.Where("(ur.app_service_id IS NULL OR ur.app_service_id = ?)", appServiceID)
	}

	codes := []string{}
	if err := query.Scan(ctx, &codes); err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return codes, nil
}

func (r *repository) GetGlobalRoleCodesByUserIDs(ctx context.Context, userIDs []string) (map[string]string, error) {
	codes := map[string]string{}
	if len(userIDs) == 0 {
		return codes, nil
	}

	type userRoleRow struct {
		UserID string `bun:"user_id"`
		Code   string `bun:"code"`
	}

	rows := []userRoleRow{}
	err := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("ur.user_id").
		ColumnExpr("rol.code").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Where("ur.user_id IN (?)", bun.In(userIDs)).
		Where("ur.app_service_id IS NULL").
		Order("ur.created_at ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	for _, row := range rows {
		if _, ok := codes[row.UserID]; !ok {
			codes[row.UserID] = row.Code
		}
	}
	return codes, nil
}
