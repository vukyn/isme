package usecase

import (
	"context"

	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	"github.com/vukyn/isme/internal/domains/role/entity"
	"github.com/vukyn/isme/internal/domains/role/models"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"

	pkgErr "github.com/vukyn/kuery/http/errors"
)

type usecase struct {
	roleRepo       roleRepo.IRepository
	userRepo       userRepo.IRepository
	appServiceRepo appServiceRepo.IRepository
}

func NewUsecase(
	roleRepo roleRepo.IRepository,
	userRepo userRepo.IRepository,
	appServiceRepo appServiceRepo.IRepository,
) IUseCase {
	return &usecase{
		roleRepo:       roleRepo,
		userRepo:       userRepo,
		appServiceRepo: appServiceRepo,
	}
}

func (u *usecase) List(ctx context.Context, req models.ListRequest) ([]models.RoleListItem, error) {
	return u.roleRepo.List(ctx, req)
}

func (u *usecase) Create(ctx context.Context, req models.CreateRequest) (models.CreateResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.CreateResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// the owning app must exist
	app, err := u.appServiceRepo.GetByID(ctx, req.AppID)
	if err != nil {
		return models.CreateResponse{}, err
	}
	if app.ID == "" {
		return models.CreateResponse{}, pkgErr.InvalidRequest("app service not found")
	}

	// check code uniqueness within the app
	existingRole, err := u.roleRepo.GetByAppAndCode(ctx, req.AppID, req.Code)
	if err != nil {
		return models.CreateResponse{}, err
	}
	if existingRole.ID != "" {
		return models.CreateResponse{}, pkgErr.InvalidRequest("role code already exists")
	}

	// resolve clone source before creating
	var clonedPermissions []entity.Permission
	if req.CloneFromRoleID != "" {
		sourceRole, err := u.roleRepo.GetByID(ctx, req.CloneFromRoleID)
		if err != nil {
			return models.CreateResponse{}, err
		}
		if sourceRole.ID == "" {
			return models.CreateResponse{}, pkgErr.InvalidRequest("clone source role not found")
		}
		// cross-app clone is rejected — a role can only inherit permissions
		// from another role owned by the same app (decision 3)
		if sourceRole.AppID != req.AppID {
			return models.CreateResponse{}, pkgErr.InvalidRequest("clone source role must belong to the same app")
		}
		clonedPermissions, err = u.roleRepo.GetPermissionsByRoleID(ctx, sourceRole.ID)
		if err != nil {
			return models.CreateResponse{}, err
		}
	}

	// create role
	roleID, err := u.roleRepo.Create(ctx, req)
	if err != nil {
		return models.CreateResponse{}, err
	}

	// copy permissions from the clone source
	if len(clonedPermissions) > 0 {
		permissionIDs := make([]int64, 0, len(clonedPermissions))
		for _, permission := range clonedPermissions {
			permissionIDs = append(permissionIDs, permission.ID)
		}
		if err := u.roleRepo.ReplaceRolePermissions(ctx, roleID, permissionIDs); err != nil {
			return models.CreateResponse{}, err
		}
	}

	return models.CreateResponse{
		ID: roleID,
	}, nil
}

func (u *usecase) GetDetail(ctx context.Context, id string) (models.RoleDetailResponse, error) {
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return models.RoleDetailResponse{}, err
	}
	if role.ID == "" {
		return models.RoleDetailResponse{}, pkgErr.NotFound("role not found")
	}

	permissions, err := u.roleRepo.GetPermissionsByRoleID(ctx, role.ID)
	if err != nil {
		return models.RoleDetailResponse{}, err
	}

	permissionItems := make([]models.PermissionItem, 0, len(permissions))
	for _, permission := range permissions {
		permissionItems = append(permissionItems, models.PermissionItem{
			ID:       permission.ID,
			AppID:    permission.AppID,
			Resource: permission.Resource,
			Action:   permission.Action,
			Icon:     permission.Icon,
			Color:    permission.Color,
		})
	}

	// resolve the owning app_code for the response
	appCode := ""
	if role.AppID != "" {
		app, err := u.appServiceRepo.GetByID(ctx, role.AppID)
		if err != nil {
			return models.RoleDetailResponse{}, err
		}
		appCode = app.AppCode
	}

	return models.RoleDetailResponse{
		ID:          role.ID,
		AppID:       role.AppID,
		AppCode:     appCode,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
		Icon:        role.Icon,
		Color:       role.Color,
		IsSystem:    role.IsSystem,
		Permissions: permissionItems,
	}, nil
}

func (u *usecase) Update(ctx context.Context, id string, req models.UpdateRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// check role exists and is editable
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.ID == "" {
		return pkgErr.NotFound("role not found")
	}
	if role.IsSystem {
		return pkgErr.Forbidden("system role cannot be modified")
	}

	return u.roleRepo.Update(ctx, id, req)
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	// check role exists and is deletable
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.ID == "" {
		return pkgErr.NotFound("role not found")
	}
	if role.IsSystem {
		return pkgErr.Forbidden("system role cannot be deleted")
	}

	// reject when members still hold the role
	membersCount, err := u.roleRepo.CountMembersByRoleID(ctx, id)
	if err != nil {
		return err
	}
	if membersCount > 0 {
		return pkgErr.InvalidRequest("role has members; reassign first")
	}

	return u.roleRepo.SoftDelete(ctx, id)
}

func (u *usecase) SetPermissions(ctx context.Context, id string, req models.SetPermissionsRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// check role exists and is editable
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.ID == "" {
		return pkgErr.NotFound("role not found")
	}
	if role.IsSystem {
		return pkgErr.Forbidden("system role cannot be modified")
	}

	return u.roleRepo.ReplaceRolePermissions(ctx, id, req.PermissionIDs)
}

func (u *usecase) ListPermissions(ctx context.Context, req models.ListPermissionsRequest) ([]models.PermissionItem, error) {
	permissions, err := u.roleRepo.ListPermissions(ctx, req)
	if err != nil {
		return nil, err
	}

	// per-resource consistency: every row of a resource reports the same icon and
	// color — the ones on the resource's lowest-id (first) row. Rows are returned
	// ordered by id ASC, so the first row seen per (app_id, resource) is
	// authoritative.
	iconByResource := map[string]string{}
	colorByResource := map[string]string{}
	for _, permission := range permissions {
		key := permission.AppID + "\x00" + permission.Resource
		if _, seen := iconByResource[key]; !seen {
			iconByResource[key] = permission.Icon
			colorByResource[key] = permission.Color
		}
	}

	items := make([]models.PermissionItem, 0, len(permissions))
	for _, permission := range permissions {
		key := permission.AppID + "\x00" + permission.Resource
		items = append(items, models.PermissionItem{
			ID:       permission.ID,
			AppID:    permission.AppID,
			Resource: permission.Resource,
			Action:   permission.Action,
			Icon:     iconByResource[key],
			Color:    colorByResource[key],
		})
	}
	return items, nil
}

// CreatePermissions adds resource:action permissions to an app's catalog and
// returns the resulting permission items (with ids) so callers can refresh.
// Creating permissions for the isme system app is rejected — its catalog is
// seeded and read-only.
func (u *usecase) CreatePermissions(ctx context.Context, req models.CreatePermissionsRequest) ([]models.PermissionItem, error) {
	// validation
	if err := req.Validate(); err != nil {
		return nil, pkgErr.InvalidRequest(err.Error())
	}

	// the isme system app owns its permission catalog and is read-only
	if req.AppID == roleConstants.APP_ID_ISME {
		return nil, pkgErr.Forbidden("isme system app permissions are read-only")
	}

	// the owning app must exist
	app, err := u.appServiceRepo.GetByID(ctx, req.AppID)
	if err != nil {
		return nil, err
	}
	if app.ID == "" {
		return nil, pkgErr.InvalidRequest("app service not found")
	}
	// guard against the system app by code too (defense in depth)
	if app.AppCode == roleConstants.APP_CODE_ISME {
		return nil, pkgErr.Forbidden("isme system app permissions are read-only")
	}

	perms := make([]models.PermissionItem, 0, len(req.Permissions))
	for _, permission := range req.Permissions {
		perms = append(perms, models.PermissionItem{Resource: permission.Resource, Action: permission.Action, Icon: permission.Icon, Color: permission.Color})
	}

	permissionIDsByCode, err := u.roleRepo.CreatePermissions(ctx, req.AppID, perms)
	if err != nil {
		return nil, err
	}

	// resolve each resource's authoritative icon and color (the repo may have
	// reused an existing resource's values instead of the requested ones) by
	// re-reading the app catalog once and indexing the lowest-id row per resource.
	catalog, err := u.ListPermissions(ctx, models.ListPermissionsRequest{AppID: req.AppID})
	if err != nil {
		return nil, err
	}
	iconByResource := map[string]string{}
	colorByResource := map[string]string{}
	for _, item := range catalog {
		if _, seen := iconByResource[item.Resource]; !seen {
			iconByResource[item.Resource] = item.Icon
			colorByResource[item.Resource] = item.Color
		}
	}

	items := make([]models.PermissionItem, 0, len(perms))
	for _, permission := range perms {
		items = append(items, models.PermissionItem{
			ID:       permissionIDsByCode[permission.Resource+":"+permission.Action],
			AppID:    req.AppID,
			Resource: permission.Resource,
			Action:   permission.Action,
			Icon:     iconByResource[permission.Resource],
			Color:    colorByResource[permission.Resource],
		})
	}
	return items, nil
}

// DeletePermission removes a resource:action permission from an app's catalog
// and clears any role grants referencing it. Deleting a permission owned by the
// isme system app is rejected — its catalog is seeded and read-only.
func (u *usecase) DeletePermission(ctx context.Context, permissionID int64) error {
	if permissionID == 0 {
		return pkgErr.InvalidRequest("permission_id is required")
	}

	// the permission must exist
	permission, err := u.roleRepo.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return err
	}
	if permission.ID == 0 {
		return pkgErr.NotFound("permission not found")
	}

	// the isme system app owns its permission catalog and is read-only
	if permission.AppID == roleConstants.APP_ID_ISME {
		return pkgErr.Forbidden("isme system app permissions are read-only")
	}

	// resolve the owning app and guard by code too (defense in depth)
	app, err := u.appServiceRepo.GetByID(ctx, permission.AppID)
	if err != nil {
		return err
	}
	if app.AppCode == roleConstants.APP_CODE_ISME {
		return pkgErr.Forbidden("isme system app permissions are read-only")
	}

	return u.roleRepo.DeletePermission(ctx, permissionID)
}

// UpdatePermissionAppearance changes a resource's per-resource icon and color
// across every resource:action row of an (app_id, resource). Editing the isme
// system app is rejected — its catalog is seeded and read-only. The resource
// must already exist in the app's catalog.
func (u *usecase) UpdatePermissionAppearance(ctx context.Context, req models.UpdatePermissionAppearanceRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// the isme system app owns its permission catalog and is read-only
	if req.AppID == roleConstants.APP_ID_ISME {
		return pkgErr.Forbidden("isme system app permissions are read-only")
	}

	// the owning app must exist
	app, err := u.appServiceRepo.GetByID(ctx, req.AppID)
	if err != nil {
		return err
	}
	if app.ID == "" {
		return pkgErr.InvalidRequest("app service not found")
	}
	// guard against the system app by code too (defense in depth)
	if app.AppCode == roleConstants.APP_CODE_ISME {
		return pkgErr.Forbidden("isme system app permissions are read-only")
	}

	// the resource must exist in the app's catalog
	items, err := u.ListPermissions(ctx, models.ListPermissionsRequest{AppID: req.AppID})
	if err != nil {
		return err
	}
	found := false
	for _, item := range items {
		if item.Resource == req.Resource {
			found = true
			break
		}
	}
	if !found {
		return pkgErr.NotFound("resource not found in app catalog")
	}

	return u.roleRepo.UpdatePermissionAppearance(ctx, req.AppID, req.Resource, req.Icon, req.Color)
}

// ProvisionDefaultRoles seeds an empty "admin" role for a newly created app.
// The role starts with zero permissions — resource:action permissions are
// created and assigned manually afterward (see CreatePermissions). Idempotent:
// re-running for an app that already has the admin role is a no-op.
func (u *usecase) ProvisionDefaultRoles(ctx context.Context, appID string) error {
	if appID == "" {
		return pkgErr.InvalidRequest("app_id is required")
	}

	// idempotency: skip when the admin role already exists for this app
	existing, err := u.roleRepo.GetByAppAndCode(ctx, appID, roleConstants.ROLE_CODE_ADMIN)
	if err != nil {
		return err
	}
	if existing.ID != "" {
		return nil
	}

	// create the empty admin role; no permissions are seeded
	_, err = u.roleRepo.Create(ctx, models.CreateRequest{
		AppID:       appID,
		Code:        roleConstants.ROLE_CODE_ADMIN,
		Name:        "Admin",
		Description: "Full access to every resource",
	})
	return err
}

func (u *usecase) ListMembers(ctx context.Context, id string, req models.ListMembersRequest) (models.ListMembersResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.ListMembersResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// check role exists
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return models.ListMembersResponse{}, err
	}
	if role.ID == "" {
		return models.ListMembersResponse{}, pkgErr.NotFound("role not found")
	}

	items, total, err := u.roleRepo.ListMembers(ctx, id, req)
	if err != nil {
		return models.ListMembersResponse{}, err
	}

	return models.ListMembersResponse{
		Items: items,
		Total: total,
		Page:  req.Page,
	}, nil
}

func (u *usecase) AddMembers(ctx context.Context, id string, req models.AddMembersRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// check role exists
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.ID == "" {
		return pkgErr.NotFound("role not found")
	}

	// check users exist
	for _, userID := range req.UserIDs {
		user, err := u.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if user.ID == "" {
			return pkgErr.InvalidRequest("user not found: " + userID)
		}
	}

	// app_service_id MUST equal the role's owning app_id — the perm query enforces
	// ur.app_service_id = rol.app_id, so a missing or mismatched value silently
	// grants no perms in the issued token. Derive it from the role rather than
	// trusting the client (the UI add-to-role flow omits it).
	appServiceID := role.AppID
	return u.roleRepo.AddMembers(ctx, id, req.UserIDs, &appServiceID)
}

func (u *usecase) RemoveMember(ctx context.Context, id string, userID string, appServiceID *string) error {
	// check role exists
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role.ID == "" {
		return pkgErr.NotFound("role not found")
	}

	return u.roleRepo.RemoveMember(ctx, id, userID, appServiceID)
}
