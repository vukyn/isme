package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/role/entity"
	"github.com/vukyn/isme/internal/domains/role/models"
)

type IRepository interface {
	// Create role (req.AppID sets the owning app)
	Create(ctx context.Context, req models.CreateRequest) (string, error)
	// Get role by ID
	GetByID(ctx context.Context, id string) (entity.Role, error)
	// Get role by app + code (codes are unique per app)
	GetByAppAndCode(ctx context.Context, appID string, code string) (entity.Role, error)
	// List roles with member counts; req filters by owning app (empty = all apps)
	List(ctx context.Context, req models.ListRequest) ([]models.RoleListItem, error)
	// Update role name and description
	Update(ctx context.Context, id string, req models.UpdateRequest) error
	// Soft delete role
	SoftDelete(ctx context.Context, id string) error
	// List the permission catalog; req filters by owning app (empty = all apps)
	ListPermissions(ctx context.Context, req models.ListPermissionsRequest) ([]entity.Permission, error)
	// Create permissions for an app (idempotent); returns the resulting permission IDs keyed by code
	CreatePermissions(ctx context.Context, appID string, perms []models.PermissionItem) (map[string]int64, error)
	// Get a single permission by ID
	GetPermissionByID(ctx context.Context, permissionID int64) (entity.Permission, error)
	// Delete a permission from the catalog; first clears any role_permissions
	// grants referencing it, then deletes the permissions row (one transaction)
	DeletePermission(ctx context.Context, permissionID int64) error
	// Get permissions assigned to a role
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]entity.Permission, error)
	// Get the resource:action permission codes granted by each role, keyed by
	// role_id (batched to avoid an N+1 over a set of roles)
	GetPermissionCodesByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error)
	// Replace all permissions of a role (delete-then-insert in a transaction)
	ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []int64) error
	// List role members with pagination and optional name/email search
	ListMembers(ctx context.Context, roleID string, req models.ListMembersRequest) ([]models.MemberItem, int, error)
	// Count members assigned to a role
	CountMembersByRoleID(ctx context.Context, roleID string) (int, error)
	// Add members to a role; nil appServiceID means a global assignment
	AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error
	// Remove a member from a role; nil appServiceID targets the global assignment
	RemoveMember(ctx context.Context, roleID string, userID string, appServiceID *string) error
	// Get permission codes for a user scoped to a concrete app_id (matched against
	// the owning role's app_id); empty appID resolves assignments across all apps
	GetPermissionCodesByUserID(ctx context.Context, userID string, appID string) ([]string, error)
	// Get permission codes for a user grouped by owning app_code (feeds resource_access)
	GetPermissionCodesGroupedByApp(ctx context.Context, userID string) (map[string][]string, error)
	// Get the app_codes a user holds any role in (feeds the token audience)
	GetAppCodesByUserID(ctx context.Context, userID string) ([]string, error)
	// Get role codes for a user; empty appServiceID resolves global assignments only
	GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error)
	// Get every user's full set of app-scoped roles, keyed by user_id (batched
	// for the user list — each entry carries app + role codes and display names)
	GetRoleCodesGroupedByAppByUserIDs(ctx context.Context, userIDs []string) (map[string][]models.UserAppRole, error)
}
