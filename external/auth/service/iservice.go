package service

import (
	"context"
	"isme/external/auth/models"
)

type IService interface {
	RequestLogin(ctx context.Context, req *models.RequestLoginRequest) (*models.RequestLoginResponse, error)
	ExchangeCode(ctx context.Context, req *models.ExchangeCodeRequest) (*models.ExchangeCodeResponse, error)
}
