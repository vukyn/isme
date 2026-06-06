package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/role/entity"
	"github.com/vukyn/isme/internal/domains/role/models"
)

type IRepository interface {
	// Create role
	Create(ctx context.Context, req models.CreateRequest) (string, error)
	// Get role by ID
	GetByID(ctx context.Context, id string) (entity.Role, error)
	// Get role by code
	GetByCode(ctx context.Context, code string) (entity.Role, error)
	// List roles with member counts
	List(ctx context.Context) ([]models.RoleListItem, error)
	// Update role name and description
	Update(ctx context.Context, id string, req models.UpdateRequest) error
	// Soft delete role
	SoftDelete(ctx context.Context, id string) error
	// List the full permission catalog
	ListPermissions(ctx context.Context) ([]entity.Permission, error)
	// Get permissions assigned to a role
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]entity.Permission, error)
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
	// Get permission codes for a user; empty appServiceID resolves global assignments only
	GetPermissionCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error)
	// Get role codes for a user; empty appServiceID resolves global assignments only
	GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error)
	// Get the primary global role code per user for a set of users
	GetGlobalRoleCodesByUserIDs(ctx context.Context, userIDs []string) (map[string]string, error)
}
