package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/user/models"
)

type IUseCase interface {
	// List users with pagination, search, status and role filters
	List(ctx context.Context, req models.ListRequest) (models.ListResponse, error)
	// Update user status (active/inactive)
	UpdateStatus(ctx context.Context, id string, req models.UpdateStatusRequest) error
	// Verify a user account (one-way — unblocks login; there is no unverify)
	VerifyUser(ctx context.Context, id string) error
	// Soft delete a user and revoke all their sessions
	SoftDelete(ctx context.Context, id string) error
	// List active sessions of a user
	ListSessions(ctx context.Context, userID string) ([]models.SessionItem, error)
	// Revoke a specific session of a user
	RevokeSession(ctx context.Context, userID string, sessionID string) error
}
