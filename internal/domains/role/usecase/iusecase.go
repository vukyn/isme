package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/role/models"
)

type IUseCase interface {
	// List roles with member counts, filtered by owning app
	List(ctx context.Context, req models.ListRequest) ([]models.RoleListItem, error)
	// Create a role, optionally cloning permissions from another role
	Create(ctx context.Context, req models.CreateRequest) (models.CreateResponse, error)
	// Get role detail including its permissions
	GetDetail(ctx context.Context, id string) (models.RoleDetailResponse, error)
	// Update role name and description
	Update(ctx context.Context, id string, req models.UpdateRequest) error
	// Delete a role (rejected for system roles and roles with members)
	Delete(ctx context.Context, id string) error
	// Replace the permissions of a role
	SetPermissions(ctx context.Context, id string, req models.SetPermissionsRequest) error
	// List the permission catalog, filtered by owning app
	ListPermissions(ctx context.Context, req models.ListPermissionsRequest) ([]models.PermissionItem, error)
	// Create resource:action permissions for an app (rejected for the isme system app)
	CreatePermissions(ctx context.Context, req models.CreatePermissionsRequest) ([]models.PermissionItem, error)
	// Delete a catalog permission and clear its grants (rejected for the isme system app)
	DeletePermission(ctx context.Context, permissionID int64) error
	// ProvisionDefaultRoles seeds the default per-app role set (an admin role
	// holding the app's full CRUD permission catalog) for a newly created app
	ProvisionDefaultRoles(ctx context.Context, appID string) error
	// List role members with pagination
	ListMembers(ctx context.Context, id string, req models.ListMembersRequest) (models.ListMembersResponse, error)
	// Add members to a role
	AddMembers(ctx context.Context, id string, req models.AddMembersRequest) error
	// Remove a member from a role
	RemoveMember(ctx context.Context, id string, userID string, appServiceID *string) error
}
