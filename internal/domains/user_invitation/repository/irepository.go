package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/user_invitation/entity"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"
)

type IRepository interface {
	// Create invitation (token hash, expiry and audit fields set by caller)
	Create(ctx context.Context, invitation entity.UserInvitation) (string, error)
	// Get invitation by ID
	GetByID(ctx context.Context, id string) (entity.UserInvitation, error)
	// Get invitation by token hash
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.UserInvitation, error)
	// Get pending invitation for an email
	GetPendingByEmail(ctx context.Context, email string) (entity.UserInvitation, error)
	// List invitations with role names
	List(ctx context.Context) ([]models.InvitationListItem, error)
	// Atomically claim a pending invitation as accepted; false when it was not pending
	MarkAccepted(ctx context.Context, id string) (bool, error)
	// Atomically flip a pending invitation to revoked; false when it was not pending
	MarkRevoked(ctx context.Context, id string) (bool, error)
	// Roll an accepted claim back to pending (compensation when user creation fails)
	RevertToPending(ctx context.Context, id string) error
}
