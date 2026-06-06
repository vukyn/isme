package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/role/models"
)

type IUseCase interface {
	// List roles with member counts
	List(ctx context.Context) ([]models.RoleListItem, error)
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
	// List the full permission catalog
	ListPermissions(ctx context.Context) ([]models.PermissionItem, error)
	// List role members with pagination
	ListMembers(ctx context.Context, id string, req models.ListMembersRequest) (models.ListMembersResponse, error)
	// Add members to a role
	AddMembers(ctx context.Context, id string, req models.AddMembersRequest) error
	// Remove a member from a role
	RemoveMember(ctx context.Context, id string, userID string, appServiceID *string) error
}
