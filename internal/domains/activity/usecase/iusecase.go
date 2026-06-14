package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/activity/models"
)

type IUseCase interface {
	// RecordSignIn records a genuine human sign-in. Best-effort: a recorder
	// failure is logged and swallowed, never failing the login.
	RecordSignIn(ctx context.Context, userID, device, clientIP string)
	// RecordSignOut records a logout. Best-effort.
	RecordSignOut(ctx context.Context, userID string)
	// RecordPasswordChanged records a password change. Best-effort.
	RecordPasswordChanged(ctx context.Context, userID string)
	// RecordProfileUpdated records a self-service profile update. Best-effort.
	RecordProfileUpdated(ctx context.Context, userID string)
	// RecordInvitationSent records an invitation send. Best-effort.
	RecordInvitationSent(ctx context.Context, inviterID, email string, roleNames []string)
	// List returns the caller's most recent activity items, newest first.
	List(ctx context.Context, userID string, limit int) ([]models.ActivityItem, error)
}
