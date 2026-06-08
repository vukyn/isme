package usecase

import (
	"context"
	"slices"
	"time"

	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"
	pkgClaims "github.com/vukyn/kuery/claims"
	pkgCtx "github.com/vukyn/kuery/ctx"
	"github.com/vukyn/kuery/jwt"
)

// buildTokenScope derives the resource_access map and audience for an access token
// from the user's permissions grouped by owning app_code.
//
//   - When appCode is non-empty (SSO login), the token is aud-restricted to that
//     single app and carries only that app's perms.
//   - When appCode is empty (first-party isme login), the token spans every app the
//     user has roles in, plus the isme self-app, for the admin console's full view.
func buildTokenScope(groupedPerms map[string][]string, appCode string) (map[string][]string, []string) {
	if appCode != "" {
		return map[string][]string{
			appCode: groupedPerms[appCode],
		}, []string{appCode}
	}

	resourceAccess := make(map[string][]string, len(groupedPerms)+1)
	audience := make([]string, 0, len(groupedPerms)+1)
	for code, perms := range groupedPerms {
		resourceAccess[code] = perms
		audience = append(audience, code)
	}
	// isme is always a valid audience for a first-party token, even if the user
	// holds no isme-app perms (so the admin console can still verify the token)
	if !slices.Contains(audience, roleConstants.APP_CODE_ISME) {
		audience = append(audience, roleConstants.APP_CODE_ISME)
	}
	return resourceAccess, audience
}

func (u *usecase) generateAccessTokens(userID, email string, resourceAccess map[string][]string, audience []string) (string, pkgClaims.Claims, error) {
	authCfg := u.cfg.Auth
	claims := pkgClaims.NewClaims(userID, email, int64(authCfg.AccessTokenExpireIn)).
		WithResourceAccess(resourceAccess).
		WithAudience(audience)
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

func (u *usecase) createUserSession(ctx context.Context, userID, tokenID, email, refreshToken, appServiceID string, expiresAt time.Time) (string, error) {
	res, err := u.userSessionRepo.Create(ctx, userSessionModels.CreateRequest{
		UserID:       userID,
		TokenID:      tokenID,
		Email:        email,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ClientIP:     pkgCtx.GetClientIP(ctx),
		UserAgent:    pkgCtx.GetUserAgent(ctx),
		AppServiceID: appServiceID,
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
