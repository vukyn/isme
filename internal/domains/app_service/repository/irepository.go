package repository

import (
	"context"

	"github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/app_service/models"
)

type IRepository interface {
	// Create app service
	Create(ctx context.Context, req entity.CreateRequest) (string, error)
	GetByID(ctx context.Context, id string) (entity.AppService, error)
	// Get app services for a set of ids, keyed by id (batched lookup)
	GetByIDs(ctx context.Context, ids []string) (map[string]entity.AppService, error)
	// Get app service by code
	GetByCode(ctx context.Context, code string) (entity.AppService, error)
	// Update app service
	Update(ctx context.Context, req entity.UpdateRequest) error
	// List app services with pagination and filters
	List(ctx context.Context, req models.ListRequest) ([]entity.AppService, int64, error)
	// Update app service status
	UpdateStatus(ctx context.Context, id string, status int32) error
}
