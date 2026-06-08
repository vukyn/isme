package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/user_invitation/entity"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"
)

type IRepository interface {
	// Create invitation with its app-scoped role assignments in one transaction
	// (token hash, expiry and audit fields set by caller). Returns the new id.
	Create(ctx context.Context, invitation entity.UserInvitation, assignments []entity.UserInvitationRole) (string, error)
	// Get invitation by ID
	GetByID(ctx context.Context, id string) (entity.UserInvitation, error)
	// Get invitation by token hash
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.UserInvitation, error)
	// Get pending invitation for an email
	GetPendingByEmail(ctx context.Context, email string) (entity.UserInvitation, error)
	// Get the role assignments carried by an invitation
	GetAssignmentsByInvitationID(ctx context.Context, invitationID string) ([]entity.UserInvitationRole, error)
	// List invitations with their role assignments
	List(ctx context.Context) ([]models.InvitationListItem, error)
	// Atomically claim a pending invitation as accepted; false when it was not pending
	MarkAccepted(ctx context.Context, id string) (bool, error)
	// Atomically flip a pending invitation to revoked; false when it was not pending
	MarkRevoked(ctx context.Context, id string) (bool, error)
	// Roll an accepted claim back to pending (compensation when user creation fails)
	RevertToPending(ctx context.Context, id string) error
}
