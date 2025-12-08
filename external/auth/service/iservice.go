package service

import (
	"context"
	"isme/external/auth/models"
)

type IService interface {
	RequestSSOLogin(ctx context.Context, req *models.RequestSSOLoginRequest) (*models.RequestSSOLoginResponse, error)
}
