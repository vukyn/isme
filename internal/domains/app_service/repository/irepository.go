package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/app_service/entity"
)

type IRepository interface {
	// Create app service
	Create(ctx context.Context, req entity.CreateRequest) (string, error)
	GetByID(ctx context.Context, id string) (entity.AppService, error)
	// Get app service by code
	GetByCode(ctx context.Context, code string) (entity.AppService, error)
	// Update app service
	Update(ctx context.Context, req entity.UpdateRequest) error
}
