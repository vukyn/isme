package usecase

import (
	"context"
	"isme/internal/domains/app_service/models"
)

type IUseCase interface {
	RegisterApp(ctx context.Context, req models.RegisterRequest) (models.RegisterResponse, error)
	VerifyApp(ctx context.Context, req models.VerifyRequest) (models.VerifyResponse, error)
}
