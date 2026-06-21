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
		AppID:       req.AppID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
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

func (r *repository) GetByAppAndCode(ctx context.Context, appID string, code string) (entity.Role, error) {
	if appID == "" {
		return entity.Role{}, pkgErr.InvalidRequest("app_id is required")
	}
	if code == "" {
		return entity.Role{}, pkgErr.InvalidRequest("code is required")
	}

	role := entity.Role{}
	err := r.db.NewSelect().
		Model(&role).
		Where("app_id = ?", appID).
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

func (r *repository) List(ctx context.Context, req models.ListRequest) ([]models.RoleListItem, error) {
	type roleListRow struct {
		entity.Role  `bun:",extend"`
		AppCode      string `bun:"app_code,scanonly"`
		MembersCount int    `bun:"members_count,scanonly"`
	}

	rows := []roleListRow{}
	query := r.db.NewSelect().
		Model(&rows).
		ColumnExpr("rol.*").
		ColumnExpr("app.app_code AS app_code").
		ColumnExpr("(SELECT COUNT(*) FROM user_roles ur WHERE ur.role_id = rol.id) AS members_count").
		Join("LEFT JOIN app_services AS app ON app.id = rol.app_id").
		Order("rol.created_at ASC")
	if req.AppID != "" {
		query = query.Where("rol.app_id = ?", req.AppID)
	}
	if req.AppCode != "" {
		query = query.Where("app.app_code = ?", req.AppCode)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	items := make([]models.RoleListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.RoleListItem{
			ID:           row.ID,
			AppID:        row.AppID,
			AppCode:      row.AppCode,
			Code:         row.Code,
			Name:         row.Name,
			Description:  row.Description,
			Icon:         row.Icon,
			Color:        row.Color,
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
		Icon:        req.Icon,
		Color:       req.Color,
		UpdatedBy:   pkgCtx.GetUserID(ctx),
	}
	_, err := r.db.NewUpdate().
		Model(role).
		Column("name", "description", "icon", "color", "updated_by").
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

func (r *repository) ListPermissions(ctx context.Context, req models.ListPermissionsRequest) ([]entity.Permission, error) {
	permissions := []entity.Permission{}
	query := r.db.NewSelect().
		Model(&permissions).
		Order("perm.id ASC")
	if req.AppID != "" {
		query = query.Where("perm.app_id = ?", req.AppID)
	}
	if req.AppCode != "" {
		query = query.
			Join("JOIN app_services AS app ON app.id = perm.app_id").
			Where("app.app_code = ?", req.AppCode)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return permissions, nil
}

// CreatePermissions idempotently inserts the given resource:action permissions for
// an app and returns their permission IDs keyed by "resource:action". The icon and
// color are stored per resource: if the (app_id, resource) already has rows, those
// rows' existing icon/color are reused for new rows of that resource (never
// overwritten) so a resource keeps one consistent icon and color; only a
// brand-new resource takes the requested icon/color.
func (r *repository) CreatePermissions(ctx context.Context, appID string, perms []models.PermissionItem) (map[string]int64, error) {
	if appID == "" {
		return nil, pkgErr.InvalidRequest("app_id is required")
	}

	// resolve each resource's effective icon/color once: an existing resource keeps
	// its stored icon/color; a new resource takes the values supplied on its first
	// pair.
	iconByResource := map[string]string{}
	colorByResource := map[string]string{}
	for _, perm := range perms {
		if _, resolved := iconByResource[perm.Resource]; resolved {
			continue
		}
		existingIcon, err := r.getResourceIcon(ctx, appID, perm.Resource)
		if err != nil {
			return nil, err
		}
		if existingIcon != "" {
			iconByResource[perm.Resource] = existingIcon
		} else {
			iconByResource[perm.Resource] = perm.Icon
		}
		existingColor, err := r.getResourceColor(ctx, appID, perm.Resource)
		if err != nil {
			return nil, err
		}
		if existingColor != "" {
			colorByResource[perm.Resource] = existingColor
		} else {
			colorByResource[perm.Resource] = perm.Color
		}
	}

	permissionIDs := map[string]int64{}
	for _, perm := range perms {
		if _, err := r.db.NewInsert().
			Model(&entity.Permission{
				AppID:    appID,
				Resource: perm.Resource,
				Action:   perm.Action,
				Icon:     iconByResource[perm.Resource],
				Color:    colorByResource[perm.Resource],
			}).
			Ignore().
			Exec(ctx); err != nil {
			return nil, pkgErr.DatabaseError(err.Error())
		}

		var permissionID int64
		err := r.db.NewSelect().
			Model((*entity.Permission)(nil)).
			Column("id").
			Where("app_id = ?", appID).
			Where("resource = ?", perm.Resource).
			Where("action = ?", perm.Action).
			Scan(ctx, &permissionID)
		if err != nil {
			return nil, pkgErr.DatabaseError(err.Error())
		}
		permissionIDs[perm.Resource+":"+perm.Action] = permissionID
	}
	return permissionIDs, nil
}

// getResourceIcon returns the icon stored on the lowest-id (first) row of the
// given (app_id, resource), or "" when the resource has no rows yet.
func (r *repository) getResourceIcon(ctx context.Context, appID string, resource string) (string, error) {
	var icon string
	err := r.db.NewSelect().
		Model((*entity.Permission)(nil)).
		Column("icon").
		Where("app_id = ?", appID).
		Where("resource = ?", resource).
		Order("id ASC").
		Limit(1).
		Scan(ctx, &icon)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", pkgErr.DatabaseError(err.Error())
	}
	return icon, nil
}

// getResourceColor returns the color stored on the lowest-id (first) row of the
// given (app_id, resource), or "" when the resource has no rows yet. Mirrors
// getResourceIcon — the per-resource color is read from the resource's first row.
func (r *repository) getResourceColor(ctx context.Context, appID string, resource string) (string, error) {
	var color string
	err := r.db.NewSelect().
		Model((*entity.Permission)(nil)).
		Column("color").
		Where("app_id = ?", appID).
		Where("resource = ?", resource).
		Order("id ASC").
		Limit(1).
		Scan(ctx, &color)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", pkgErr.DatabaseError(err.Error())
	}
	return color, nil
}

func (r *repository) GetPermissionByID(ctx context.Context, permissionID int64) (entity.Permission, error) {
	if permissionID == 0 {
		return entity.Permission{}, pkgErr.InvalidRequest("permission_id is required")
	}

	permission := entity.Permission{}
	err := r.db.NewSelect().
		Model(&permission).
		Where("id = ?", permissionID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Permission{}, nil
		}
		return entity.Permission{}, pkgErr.DatabaseError(err.Error())
	}
	return permission, nil
}

// DeletePermission removes a permission from the catalog. Any role_permissions
// rows granting it are cleared first so no role keeps a dangling grant, then the
// permissions row is deleted — both in one transaction.
func (r *repository) DeletePermission(ctx context.Context, permissionID int64) error {
	if permissionID == 0 {
		return pkgErr.InvalidRequest("permission_id is required")
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// clear grants referencing this permission
		_, err := tx.NewDelete().
			Model((*entity.RolePermission)(nil)).
			Where("permission_id = ?", permissionID).
			Exec(ctx)
		if err != nil {
			return err
		}

		// delete the catalog entry
		_, err = tx.NewDelete().
			Model((*entity.Permission)(nil)).
			Where("id = ?", permissionID).
			Exec(ctx)
		return err
	})
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
}

// UpdatePermissionAppearance sets the per-resource icon and color on every row
// of the given (app_id, resource) in one statement. Appearance is shared by all
// resource:action rows of a resource, so a single NewUpdate touches them all.
// The permissions table has no audit columns, so none are set (mirrors
// CreatePermissions).
func (r *repository) UpdatePermissionAppearance(ctx context.Context, appID string, resource string, icon string, color string) error {
	if appID == "" {
		return pkgErr.InvalidRequest("app_id is required")
	}
	if resource == "" {
		return pkgErr.InvalidRequest("resource is required")
	}

	_, err := r.db.NewUpdate().
		Model((*entity.Permission)(nil)).
		Set("icon = ?", icon).
		Set("color = ?", color).
		Where("app_id = ?", appID).
		Where("resource = ?", resource).
		Exec(ctx)
	if err != nil {
		return pkgErr.DatabaseError(err.Error())
	}
	return nil
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

// GetPermissionCodesByRoleIDs returns the resource:action permission codes
// granted by each role, keyed by role_id. Used pre-auth to preview what an
// invited role grants — scoped strictly to the given role ids.
func (r *repository) GetPermissionCodesByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	codesByRole := map[string][]string{}
	if len(roleIDs) == 0 {
		return codesByRole, nil
	}

	type roleCodeRow struct {
		RoleID string `bun:"role_id"`
		Code   string `bun:"code"`
	}

	rows := []roleCodeRow{}
	err := r.db.NewSelect().
		TableExpr("role_permissions AS rp").
		ColumnExpr("rp.role_id").
		ColumnExpr("perm.resource || ':' || perm.action AS code").
		Join("JOIN permissions AS perm ON perm.id = rp.permission_id").
		Where("rp.role_id IN (?)", bun.In(roleIDs)).
		Order("perm.id ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	for _, row := range rows {
		codesByRole[row.RoleID] = append(codesByRole[row.RoleID], row.Code)
	}
	return codesByRole, nil
}

// GetPermissionCodesGroupedByApp returns the user's permission codes grouped by the
// owning app's app_code. The owning role's app_id is authoritative and the
// assignment scope (user_roles.app_service_id) must match it.
func (r *repository) GetPermissionCodesGroupedByApp(ctx context.Context, userID string) (map[string][]string, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	type groupedRow struct {
		AppCode string `bun:"app_code"`
		Code    string `bun:"code"`
	}

	rows := []groupedRow{}
	err := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("DISTINCT app.app_code AS app_code").
		ColumnExpr("perm.resource || ':' || perm.action AS code").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Join("JOIN app_services AS app ON app.id = rol.app_id").
		Join("JOIN role_permissions AS rp ON rp.role_id = ur.role_id").
		Join("JOIN permissions AS perm ON perm.id = rp.permission_id").
		Where("ur.user_id = ?", userID).
		Where("ur.app_service_id = rol.app_id").
		Scan(ctx, &rows)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	grouped := map[string][]string{}
	for _, row := range rows {
		grouped[row.AppCode] = append(grouped[row.AppCode], row.Code)
	}
	return grouped, nil
}

// GetAppCodesByUserID returns the distinct app_codes the user holds any role in.
func (r *repository) GetAppCodesByUserID(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	appCodes := []string{}
	err := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("DISTINCT app.app_code").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Join("JOIN app_services AS app ON app.id = rol.app_id").
		Where("ur.user_id = ?", userID).
		Where("ur.app_service_id = rol.app_id").
		Scan(ctx, &appCodes)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return appCodes, nil
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
			// dialect-aware case-insensitive match (ILIKE on Postgres, LIKE on SQLite)
			query = query.Where(
				"("+pkgBunQuery.ILike(r.db, "usr.name")+" OR "+pkgBunQuery.ILike(r.db, "usr.email")+")",
				search, search,
			)
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

func (r *repository) GetPermissionCodesByUserID(ctx context.Context, userID string, appID string) ([]string, error) {
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user_id is required")
	}

	query := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("DISTINCT perm.resource || ':' || perm.action").
		Join("JOIN role_permissions AS rp ON rp.role_id = ur.role_id").
		Join("JOIN permissions AS perm ON perm.id = rp.permission_id").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Where("ur.user_id = ?", userID).
		// the assignment scope must match the owning role's app
		Where("ur.app_service_id = rol.app_id")
	if appID != "" {
		query = query.Where("rol.app_id = ?", appID)
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
		Where("ur.user_id = ?", userID).
		Where("ur.app_service_id = rol.app_id")
	if appServiceID != "" {
		query = query.Where("rol.app_id = ?", appServiceID)
	}

	codes := []string{}
	if err := query.Scan(ctx, &codes); err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}
	return codes, nil
}

// GetRoleCodesGroupedByAppByUserIDs returns every app-scoped role each user
// holds, keyed by user_id. The owning role's app_id is authoritative and the
// assignment scope (user_roles.app_service_id) must match it. Batched over the
// whole page to avoid an N+1.
func (r *repository) GetRoleCodesGroupedByAppByUserIDs(ctx context.Context, userIDs []string) (map[string][]models.UserAppRole, error) {
	rolesByUser := map[string][]models.UserAppRole{}
	if len(userIDs) == 0 {
		return rolesByUser, nil
	}

	type rolesRow struct {
		UserID   string `bun:"user_id"`
		AppCode  string `bun:"app_code"`
		AppName  string `bun:"app_name"`
		RoleCode string `bun:"role_code"`
		RoleName string `bun:"role_name"`
	}

	rows := []rolesRow{}
	err := r.db.NewSelect().
		TableExpr("user_roles AS ur").
		ColumnExpr("ur.user_id").
		ColumnExpr("app.app_code AS app_code").
		ColumnExpr("app.app_name AS app_name").
		ColumnExpr("rol.code AS role_code").
		ColumnExpr("rol.name AS role_name").
		Join("JOIN roles AS rol ON rol.id = ur.role_id AND rol.deleted_at IS NULL").
		Join("JOIN app_services AS app ON app.id = rol.app_id").
		Where("ur.user_id IN (?)", bun.In(userIDs)).
		Where("ur.app_service_id = rol.app_id").
		Order("app.app_code ASC").
		Order("rol.code ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, pkgErr.DatabaseError(err.Error())
	}

	for _, row := range rows {
		rolesByUser[row.UserID] = append(rolesByUser[row.UserID], models.UserAppRole{
			AppCode:  row.AppCode,
			AppName:  row.AppName,
			RoleCode: row.RoleCode,
			RoleName: row.RoleName,
		})
	}
	return rolesByUser, nil
}
