package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	userConstants "github.com/vukyn/isme/internal/domains/user/constants"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"
	pkgClaims "github.com/vukyn/kuery/claims"
	"github.com/vukyn/kuery/cryp"
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

// establishIdPSession mints a FULL-scope (isme) token pair for the user and
// always binds it to a fresh IdP user_session (AppServiceID=""). It is called on
// SSO login so the browser receives real isme cookies — enabling the silent-SSO
// consent screen on subsequent app handshakes.
//
// A fresh session is ALWAYS created (never reused/rotated): rotating an existing
// active AppServiceID="" row would log out another device, so each login gets its
// own IdP session. Sprawl is bounded by session expiry / logout.
func (u *usecase) establishIdPSession(ctx context.Context, user userEntity.User, groupedPerms map[string][]string) (accessToken, refreshToken, expiresAt string, err error) {
	// empty appCode → full/isme scope (all apps the user has roles in, plus isme)
	resourceAccess, audience := buildTokenScope(groupedPerms, "")

	accessToken, accessTokenClaims, err := u.generateAccessTokens(user.ID, user.Email, resourceAccess, audience)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, _, err = u.generateRefreshTokens(user.ID, user.Email)
	if err != nil {
		return "", "", "", err
	}

	// IdP session: empty appServiceID; always created, never rotated
	_, err = u.createUserSession(ctx, user.ID, accessTokenClaims.GetTokenID(), user.Email, refreshToken, "", accessTokenClaims.GetExpiredAt())
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, accessTokenClaims.GetExpiredAt().Format(time.RFC3339), nil
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

// mintAuthorizationCode performs the one-time-use code exchange handoff shared by
// the SSO login and consent paths: it clears the session_id from cache (so a
// double-consent can't double-mint), generates a ULID code, and stashes the
// freshly minted token triplet in cache under that code with the exchange TTL.
// The exchange endpoint later swaps the code for the tokens (also one-time use).
func (u *usecase) mintAuthorizationCode(accessToken, refreshToken, expiresAt, sessionID string) string {
	// clear session ID from cache (one-time use) — atomic guard against double-mint
	u.cache.Delete(sessionID)

	// generate authorization code
	authorizationCode := cryp.ULID()

	// set tokens to cache for exchange token
	ttl := time.Duration(u.cfg.Auth.ExternalExchangeCodeTTL) * time.Second
	u.cache.Set(keyAuthorizationCodeAccessToken(authorizationCode), accessToken, ttl)
	u.cache.Set(keyAuthorizationCodeRefreshToken(authorizationCode), refreshToken, ttl)
	u.cache.Set(keyAuthorizationCodeExpiresAt(authorizationCode), expiresAt, ttl)

	return authorizationCode
}

// validateSessionForConsent is a READ-ONLY validity probe for an existing isme
// session. It NEVER rotates the refresh token or mutates user_session — no
// rotation happens anywhere in the consent path; SSOConsent mints a fresh
// app-scoped session for the requesting app instead and leaves the caller's
// session intact. It returns the active user when the caller is still
// authenticated, or ok=false when the page must fall back to the password form.
//
// Probe order:
//   - valid (non-expired) access token whose session is active → authenticated
//   - else valid (non-expired) refresh token backed by an ACTIVE, non-expired
//     user_session and an active user → authenticated
func (u *usecase) validateSessionForConsent(ctx context.Context, accessToken, refreshToken string) (userEntity.User, bool) {
	// 1) try the access token first (cheapest path)
	if accessToken != "" {
		if user, ok := u.probeAccessToken(ctx, accessToken); ok {
			return user, true
		}
	}

	// 2) fall back to the refresh token (read-only — no rotation)
	if refreshToken != "" {
		if user, ok := u.probeRefreshToken(ctx, refreshToken); ok {
			return user, true
		}
	}

	return userEntity.User{}, false
}

// probeAccessToken validates a still-live access token and confirms its session
// is active. Read-only.
func (u *usecase) probeAccessToken(ctx context.Context, accessToken string) (userEntity.User, bool) {
	claims, err := jwt.ValidateJWTWithRSAPublicKey(accessToken, u.cfg.Auth.AccessTokenPublicKey)
	if err != nil || claims.IsExpired() {
		return userEntity.User{}, false
	}

	userSession, err := u.userSessionRepo.FindByTokenID(ctx, claims.GetTokenID())
	if err != nil || userSession.ID == "" {
		return userEntity.User{}, false
	}
	if userSession.Status != userSessionConstants.UserSessionStatusActive {
		return userEntity.User{}, false
	}
	if userSession.ExpiresAt.Before(time.Now()) {
		return userEntity.User{}, false
	}

	return u.activeUser(ctx, userSession.UserID)
}

// probeRefreshToken validates a refresh token and confirms its backing session
// is active and not expired — WITHOUT rotating it (the rotation in RefreshToken
// is deliberately not called here).
func (u *usecase) probeRefreshToken(ctx context.Context, refreshToken string) (userEntity.User, bool) {
	claims, err := jwt.ValidateJWT(refreshToken, u.cfg.Auth.RefreshTokenSecretKey)
	if err != nil || claims.IsExpired() {
		return userEntity.User{}, false
	}

	userSession, err := u.userSessionRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil || userSession.ID == "" {
		return userEntity.User{}, false
	}
	if userSession.Status != userSessionConstants.UserSessionStatusActive {
		return userEntity.User{}, false
	}
	// Do NOT gate on userSession.ExpiresAt here: that field holds the ACCESS
	// token's expiry (set from accessTokenClaims in create/updateUserSession),
	// so an expired access token makes it past — exactly the case the refresh
	// path must still accept. The refresh token's own lifetime is already
	// enforced by claims.IsExpired() above; the session only needs to be active.
	// (Mirrors RefreshToken, which likewise never checks ExpiresAt.)

	return u.activeUser(ctx, userSession.UserID)
}

// activeUser loads the user and confirms it is active.
func (u *usecase) activeUser(ctx context.Context, userID string) (userEntity.User, bool) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user.ID == "" {
		return userEntity.User{}, false
	}
	if user.Status != userConstants.UserStatusActive {
		return userEntity.User{}, false
	}
	return user, true
}

// === Consent CSRF nonce (single-use, short TTL) ===

// consentNonceTTL is the lifetime of a consent nonce — long enough for a human
// to click Allow, short enough to bound replay.
const consentNonceTTL = 5 * time.Minute

func keyConsentNonce(sessionID string) string {
	return fmt.Sprintf("auth:sso:consent:nonce:%s", sessionID)
}

// mintConsentNonce generates a fresh single-use nonce for the session and stores
// it in cache keyed by session_id.
func (u *usecase) mintConsentNonce(sessionID string) string {
	nonce := cryp.ULID()
	u.cache.Set(keyConsentNonce(sessionID), nonce, consentNonceTTL)
	return nonce
}

// consumeConsentNonce validates the provided nonce against the cache and deletes
// it (single-use). Returns false on mismatch / absence.
func (u *usecase) consumeConsentNonce(sessionID, nonce string) bool {
	stored, ok := u.cache.Get(keyConsentNonce(sessionID))
	if !ok || stored != nonce {
		return false
	}
	u.cache.Delete(keyConsentNonce(sessionID))
	return true
}
