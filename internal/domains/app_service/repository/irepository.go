package repository

import (
	"context"
	"isme/internal/domains/app_service/entity"
)

type IRepository interface {
	// Create app service
	Create(ctx context.Context, req entity.CreateRequest) (string, error)
	// Get app service by code
	GetByCode(ctx context.Context, code string) (entity.AppService, error)
}
