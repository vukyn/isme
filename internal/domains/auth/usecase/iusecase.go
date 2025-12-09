package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/auth/models"
)

type IUseCase interface {
	GetMe(ctx context.Context) (models.GetMeResponse, error)
	SignUp(ctx context.Context, req models.SignUpRequest) (models.SignUpResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error)
	RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (models.RefreshTokenResponse, error)
	VerifyToken(ctx context.Context, req models.VerifyTokenRequest) (models.VerifyTokenResponse, error)
	ChangePassword(ctx context.Context, req models.ChangePasswordRequest) error
	Logout(ctx context.Context) error
	RequestLogin(ctx context.Context, req models.RequestLoginRequest) (models.RequestLoginResponse, error)
	ExchangeCode(ctx context.Context, req models.ExchangeCodeRequest) (models.ExchangeCodeResponse, error)
}
