package usecase

import (
	"context"
	userSessionModels "isme/internal/domains/user_session/models"
	pkgClaims "isme/pkg/claims"
	pkgCtx "isme/pkg/ctx"
	"isme/pkg/jwt"
	"time"
)

func (u *usecase) generateAccessTokens(userID, email string) (string, pkgClaims.Claims, error) {
	authCfg := u.cfg.Auth
	accessToken, claims, err := jwt.GenerateJWTWithRSAPrivateKey(authCfg.AccessTokenPrivateKey, authCfg.AccessTokenExpireIn, userID, email)
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
