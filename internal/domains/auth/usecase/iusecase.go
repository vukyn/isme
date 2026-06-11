package usecase

import (
	"context"

	activityModels "github.com/vukyn/isme/internal/domains/activity/models"
	"github.com/vukyn/isme/internal/domains/auth/models"
)

type IUseCase interface {
	GetMe(ctx context.Context) (models.GetMeResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error)
	RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (models.RefreshTokenResponse, error)
	VerifyToken(ctx context.Context, req models.VerifyTokenRequest) (models.VerifyTokenResponse, error)
	ChangePassword(ctx context.Context, req models.ChangePasswordRequest) error
	Logout(ctx context.Context) error
	RequestLogin(ctx context.Context, req models.RequestLoginRequest) (models.RequestLoginResponse, error)
	ExchangeCode(ctx context.Context, req models.ExchangeCodeRequest) (models.ExchangeCodeResponse, error)
	SSOCheck(ctx context.Context, req models.SSOCheckRequest) (models.SSOCheckResponse, error)
	SSOConsent(ctx context.Context, req models.SSOConsentRequest) (models.SSOConsentResponse, error)
	ListMySessions(ctx context.Context) ([]models.MySessionItem, error)
	CountMySessions(ctx context.Context) (models.MySessionCount, error)
	RevokeMySession(ctx context.Context, sessionID string) error
	RevokeMyOtherSessions(ctx context.Context) error
	GetMyActivity(ctx context.Context, limit int) ([]activityModels.ActivityItem, error)
}
