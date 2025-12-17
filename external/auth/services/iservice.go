package services

import (
	"context"

	"github.com/vukyn/isme/external/auth/models"
)

type IService interface {
	GetMe(ctx context.Context, req *models.GetMeRequest) (*models.GetMeResponse, error)
	RequestLogin(ctx context.Context, req *models.RequestLoginRequest) (*models.RequestLoginResponse, error)
	ExchangeCode(ctx context.Context, req *models.ExchangeCodeRequest) (*models.ExchangeCodeResponse, error)
}
