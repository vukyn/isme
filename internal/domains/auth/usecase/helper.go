package usecase

import (
	"context"
	"time"

	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"
	pkgClaims "github.com/vukyn/kuery/claims"
	pkgCtx "github.com/vukyn/kuery/ctx"
	"github.com/vukyn/kuery/jwt"
)

func (u *usecase) generateAccessTokens(userID, email string, isAdmin bool, permissionCodes []string) (string, pkgClaims.Claims, error) {
	authCfg := u.cfg.Auth
	claims := pkgClaims.NewClaims(userID, email, int64(authCfg.AccessTokenExpireIn)).
		WithIsAdmin(isAdmin).
		WithPerms(permissionCodes)
	accessToken, err := jwt.GenerateJWTWithRSAPrivateKeyFromClaims(authCfg.AccessTokenPrivateKey, claims)
	if err != nil {
		return "", pkgClaims.Claims{}, err
	}
	return accessToken, claims, nil
}

func (u *usecase) generateRefreshTokens(userID, email string) (string, pkgClaims.Claims, error) {
	authCfg := u.cfg.Auth
	refreshToken, claims, err := jwt.GenerateJWT(authCfg.RefreshTokenSecretKey, authCfg.RefreshTokenExpireIn, userID, email)
	if err != nil {
		return "", pkgClaims.Claims{}, err
	}
	return refreshToken, claims, nil
}

func (u *usecase) createUserSession(ctx context.Context, userID, tokenID, email, refreshToken string, expiresAt time.Time) (string, error) {
	res, err := u.userSessionRepo.Create(ctx, userSessionModels.CreateRequest{
		UserID:       userID,
		TokenID:      tokenID,
		Email:        email,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ClientIP:     pkgCtx.GetClientIP(ctx),
		UserAgent:    pkgCtx.GetUserAgent(ctx),
	})
	if err != nil {
		return "", err
	}
	return res.ID, nil
}

func (u *usecase) updateUserSession(ctx context.Context, sessionID, tokenID, refreshToken string, expiresAt time.Time) error {
	err := u.userSessionRepo.UpdateLastLogin(ctx, userSessionModels.UpdateLastLoginRequest{
		ID:           sessionID,
		TokenID:      tokenID,
		RefreshToken: refreshToken,
		ClientIP:     pkgCtx.GetClientIP(ctx),
		UserAgent:    pkgCtx.GetUserAgent(ctx),
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return err
	}
	return nil
}
