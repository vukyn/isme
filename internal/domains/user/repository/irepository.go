package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/user/entity"
	"github.com/vukyn/isme/internal/domains/user/models"
)

type IRepository interface {
	// Create user
	Create(ctx context.Context, req models.CreateRequest) (string, error)
	// Get user by ID
	GetByID(ctx context.Context, id string) (entity.User, error)
	// Get user by email
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	// Set password for user
	SetPassword(ctx context.Context, id string, password string) error
	// Update last login to current time for user (only for successful login)
	UpdateLastLogin(ctx context.Context, id string) error
	// Promote admin: update isAdmin to 1
	PromoteAdmin(ctx context.Context, id string) error
	// IsAdmin: check if isAdmin equals 1
	IsAdmin(ctx context.Context, id string) (bool, error)
}
