package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/app_service/models"
)

type IUseCase interface {
	RegisterApp(ctx context.Context, req models.RegisterRequest) (models.RegisterResponse, error)
	VerifyApp(ctx context.Context, req models.VerifyRequest) (models.VerifyResponse, error)
	RefreshApp(ctx context.Context, req models.RefreshRequest) (models.RefreshResponse, error)
}
