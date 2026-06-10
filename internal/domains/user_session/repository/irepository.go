package repository

import (
	"context"
	"time"

	"github.com/vukyn/isme/internal/domains/user_session/entity"
	"github.com/vukyn/isme/internal/domains/user_session/models"
)

type IRepository interface {
	// Every time user login, need to call this function to create new session data
	Create(ctx context.Context, req models.CreateRequest) (entity.UserSession, error)
	// Every time user refresh token, need to call this function to update session data
	UpdateLastLogin(ctx context.Context, req models.UpdateLastLoginRequest) error
	// Inactive all user session when revoke session
	InactiveAllUserSession(ctx context.Context, userID string) error
	// Inactive specific user session by token ID
	InactiveSessionByTokenID(ctx context.Context, tokenID string) error
	// Inactive specific user session by ID
	InactiveSessionByID(ctx context.Context, sessionID string) error
	// Inactive all active sessions whose expires_at is before the given time.
	// Returns the number of sessions revoked.
	InactiveExpiredSessions(ctx context.Context, before time.Time) (int64, error)
	// Find user session by refresh token
	FindByRefreshToken(ctx context.Context, refreshToken string) (entity.UserSession, error)
	// Find user session by token ID
	FindByTokenID(ctx context.Context, tokenID string) (entity.UserSession, error)
	// Find user session by ID
	GetByID(ctx context.Context, sessionID string) (entity.UserSession, error)
	// Get list of active user sessions by user ID
	GetListActiveByUserID(ctx context.Context, userID string) ([]entity.UserSession, error)
	// Count active sessions per user for multiple user IDs
	CountActiveByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error)
	// Count active sessions for a user created after the given time
	CountActiveByUserIDCreatedAfter(ctx context.Context, userID string, after time.Time) (int, error)
	// Count token rotation events for a user at or after the given time (sliding 24h window)
	CountRotationsByUserIDSince(ctx context.Context, userID string, since time.Time) (int, error)
	// Inactive all active sessions for a user except the one with the given token ID
	InactiveAllUserSessionExcept(ctx context.Context, userID string, exceptTokenID string) error
}
