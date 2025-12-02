package repository

import (
	"context"
		"isme/internal/domains/user_session/entity"
	"isme/internal/domains/user_session/models"
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
	// Find user session by refresh token
	FindByRefreshToken(ctx context.Context, refreshToken string) (entity.UserSession, error)
	// Find user session by token ID
	FindByTokenID(ctx context.Context, tokenID string) (entity.UserSession, error)
	// Get list of active user sessions by user ID
	GetListActiveByUserID(ctx context.Context, userID string) ([]entity.UserSession, error)
}
