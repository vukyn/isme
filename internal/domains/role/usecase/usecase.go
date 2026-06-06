package usecase

import (
	"context"

	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
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

func (u *usecase) List(ctx context.Context) ([]models.RoleListItem, error) {
	return u.roleRepo.List(ctx)
}

func (u *usecase) Create(ctx context.Context, req models.CreateRequest) (models.CreateResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.CreateResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// check code uniqueness
	existingRole, err := u.roleRepo.GetByCode(ctx, req.Code)
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
			Resource: permission.Resource,
			Action:   permission.Action,
		})
	}

	return models.RoleDetailResponse{
		ID:          role.ID,
		Code:        role.Code,
		Name:        role.Name,
		Description: role.Description,
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

func (u *usecase) ListPermissions(ctx context.Context) ([]models.PermissionItem, error) {
	permissions, err := u.roleRepo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]models.PermissionItem, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, models.PermissionItem{
			ID:       permission.ID,
			Resource: permission.Resource,
			Action:   permission.Action,
		})
	}
	return items, nil
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

	// check app service exists when scoped
	if req.AppServiceID != nil {
		appService, err := u.appServiceRepo.GetByID(ctx, *req.AppServiceID)
		if err != nil {
			return err
		}
		if appService.ID == "" {
			return pkgErr.InvalidRequest("app service not found")
		}
	}

	return u.roleRepo.AddMembers(ctx, id, req.UserIDs, req.AppServiceID)
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
