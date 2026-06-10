package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/vukyn/isme/internal/config"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	"github.com/vukyn/isme/internal/domains/auth/models"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	userConstants "github.com/vukyn/isme/internal/domains/user/constants"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"
	pkgCache "github.com/vukyn/kuery/cache"
	pkgClaims "github.com/vukyn/kuery/claims"
	"github.com/vukyn/kuery/cryp/aes"
	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
	"github.com/vukyn/kuery/jwt"

	"github.com/vukyn/kuery/cryp"
)

type usecase struct {
	cfg             *config.Config
	cache           *pkgCache.Cache[string, string]
	userRepo        userRepo.IRepository
	userSessionRepo userSessionRepo.IRepository
	appServiceRepo  appServiceRepo.IRepository
	roleRepo        roleRepo.IRepository
}

func NewUsecase(
	cfg *config.Config,
	cache *pkgCache.Cache[string, string],
	userRepo userRepo.IRepository,
	userSessionRepo userSessionRepo.IRepository,
	appServiceRepo appServiceRepo.IRepository,
	roleRepo roleRepo.IRepository,
) IUseCase {
	return &usecase{
		cfg:             cfg,
		cache:           cache,
		userRepo:        userRepo,
		userSessionRepo: userSessionRepo,
		appServiceRepo:  appServiceRepo,
		roleRepo:        roleRepo,
	}
}

func (u *usecase) GetMe(ctx context.Context) (models.GetMeResponse, error) {
	userId := pkgCtx.GetUserID(ctx)
	if userId == "" {
		return models.GetMeResponse{}, pkgErr.InvalidRequest("user not found")
	}

	user, err := u.userRepo.GetByID(ctx, userId)
	if err != nil {
		return models.GetMeResponse{}, err
	}

	return models.GetMeResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (u *usecase) VerifyToken(ctx context.Context, req models.VerifyTokenRequest) (models.VerifyTokenResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.VerifyTokenResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// validate token
	claims, err := jwt.ValidateJWTWithRSAPublicKey(req.Token, u.cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		return models.VerifyTokenResponse{}, pkgErr.InvalidRequest("invalid token")
	}

	// check if token is expired
	if claims.IsExpired() {
		return models.VerifyTokenResponse{}, pkgErr.InvalidRequest("invalid token")
	}

	// check if exist user session in database
	userSession, err := u.userSessionRepo.FindByTokenID(ctx, claims.GetTokenID())
	if err != nil {
		return models.VerifyTokenResponse{}, err
	}
	if userSession.ID == "" {
		return models.VerifyTokenResponse{}, pkgErr.InvalidRequest("invalid token")
	}
	if userSession.Status != userSessionConstants.UserSessionStatusActive {
		return models.VerifyTokenResponse{}, pkgErr.InvalidRequest("invalid token")
	}

	// check if user session is expired
	if userSession.ExpiresAt.Before(time.Now()) {
		return models.VerifyTokenResponse{
			Ok:     false,
			Claims: pkgClaims.Claims{},
		}, nil
	}

	// check if user session is active
	if userSession.Status != userSessionConstants.UserSessionStatusActive {
		return models.VerifyTokenResponse{
			Ok:     false,
			Claims: pkgClaims.Claims{},
		}, nil
	}

	return models.VerifyTokenResponse{
		Ok:     true,
		Claims: claims,
	}, nil
}

func (u *usecase) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.LoginResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// check if session ID is valid
	var redirectURL, appServiceID, appCode string
	if req.SessionID != "" {
		cacheAppServiceID, ok := u.cache.Get(req.SessionID)
		if !ok {
			return models.LoginResponse{}, pkgErr.InvalidRequest("invalid session_id")
		} else {
			appServiceID = cacheAppServiceID
		}
		appService, err := u.appServiceRepo.GetByID(ctx, appServiceID)
		if err != nil {
			return models.LoginResponse{}, err
		}
		if appService.ID == "" {
			return models.LoginResponse{}, pkgErr.InvalidRequest("invalid session_id")
		}
		redirectURL = appService.RedirectURL
		appCode = appService.AppCode
	}

	// check if user exists
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return models.LoginResponse{}, err
	}
	if user.ID == "" {
		return models.LoginResponse{}, pkgErr.InvalidRequest("invalid email or password")
	}
	if user.Status != userConstants.UserStatusActive {
		return models.LoginResponse{}, pkgErr.InvalidRequest("invalid email or password")
	}

	// check if password is correct
	ok, needsRehash := cryp.VerifyPassword(req.Password, user.Password)
	if !ok {
		return models.LoginResponse{}, pkgErr.InvalidRequest("invalid email or password")
	}

	// block unverified accounts only after the credentials checked out,
	// so a wrong password never leaks the verification state
	if !user.IsVerified {
		return models.LoginResponse{}, pkgErr.Forbidden("account pending verification")
	}

	// upgrade legacy hash to the current scheme; best-effort, must not fail the login
	if needsRehash {
		_ = u.userRepo.SetPassword(ctx, user.ID, req.Password)
	}

	// load authorization data grouped by owning app for the access token claims
	groupedPerms, err := u.roleRepo.GetPermissionCodesGroupedByApp(ctx, user.ID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// build the resource_access + audience for this token:
	//   - SSO login (appServiceID set): aud-restricted to the requesting app only
	//   - first-party isme login: full multi-app token (all apps the user has roles
	//     in, plus isme itself)
	resourceAccess, audience := buildTokenScope(groupedPerms, appCode)

	// generate access tokens
	accessToken, accessTokenClaims, err := u.generateAccessTokens(user.ID, user.Email, resourceAccess, audience)
	if err != nil {
		return models.LoginResponse{}, err
	}
	expiresAt := accessTokenClaims.GetExpiredAt().Format(time.RFC3339)

	// generate refresh tokens
	refreshToken, _, err := u.generateRefreshTokens(user.ID, user.Email)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// create user session (records the requesting app for SSO refresh scoping)
	_, err = u.createUserSession(ctx, user.ID, accessTokenClaims.GetTokenID(), user.Email, refreshToken, appServiceID, accessTokenClaims.GetExpiredAt())
	if err != nil {
		return models.LoginResponse{}, err
	}

	// update user last login
	err = u.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// if login from external app service, need exchange authorization code for tokens
	var authorizationCode string
	if appServiceID != "" {
		// SSO login produces TWO distinct token pairs — they must NEVER be crossed:
		//
		//   1. APP-SCOPED tokens (accessToken/refreshToken/expiresAt above, built
		//      with buildTokenScope(groupedPerms, appCode), aud-restricted to the
		//      requesting app) → go ONLY into mintAuthorizationCode. The app later
		//      redeems the code via ExchangeCode for these aud-restricted tokens.
		//
		//   2. IdP-SCOPED tokens (full isme scope, fresh IdP session) → go ONLY into
		//      the LoginResponse. The browser writes these as real isme cookies so the
		//      silent-SSO consent screen can trigger on later app handshakes.
		authorizationCode = u.mintAuthorizationCode(accessToken, refreshToken, expiresAt, req.SessionID)

		// establish the isme IdP browser session and return ITS (full-scope) tokens
		idpAccess, idpRefresh, idpExpires, err := u.establishIdPSession(ctx, user, groupedPerms)
		if err != nil {
			return models.LoginResponse{}, err
		}
		accessToken = idpAccess
		refreshToken = idpRefresh
		expiresAt = idpExpires
	}

	return models.LoginResponse{
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		ExpiresAt:         expiresAt,
		RedirectURL:       redirectURL,
		AuthorizationCode: authorizationCode,
	}, nil
}

func (u *usecase) RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (models.RefreshTokenResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// validate refresh token
	authCfg := u.cfg.Auth
	claims, err := jwt.ValidateJWT(req.RefreshToken, authCfg.RefreshTokenSecretKey)
	if err != nil {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest("invalid refresh token")
	}

	// check if token is expired
	if claims.IsExpired() {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest("invalid refresh token")
	}

	// check if user session exists and active
	userSession, err := u.userSessionRepo.FindByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}
	if userSession.ID == "" {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest("invalid refresh token")
	}
	if userSession.Status != userSessionConstants.UserSessionStatusActive {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest("invalid refresh token")
	}

	// check if user still exists and is active
	user, err := u.userRepo.GetByID(ctx, userSession.UserID)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}
	if user.ID == "" {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest("invalid refresh token")
	}
	if user.Status != userConstants.UserStatusActive {
		return models.RefreshTokenResponse{}, pkgErr.InvalidRequest("invalid refresh token")
	}

	// re-load authorization data from the database so revoked rights never survive a refresh
	groupedPerms, err := u.roleRepo.GetPermissionCodesGroupedByApp(ctx, user.ID)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// preserve the original token scope: SSO sessions (with a recorded app) refresh
	// aud-restricted to that app; first-party sessions refresh as a full token
	appCode := ""
	if userSession.AppServiceID != "" {
		appService, err := u.appServiceRepo.GetByID(ctx, userSession.AppServiceID)
		if err != nil {
			return models.RefreshTokenResponse{}, err
		}
		appCode = appService.AppCode
	}
	resourceAccess, audience := buildTokenScope(groupedPerms, appCode)

	// generate new access tokens
	newAccessToken, accessTokenClaims, err := u.generateAccessTokens(user.ID, user.Email, resourceAccess, audience)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// generate new refresh tokens
	newRefreshToken, _, err := u.generateRefreshTokens(user.ID, user.Email)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// update user session
	err = u.updateUserSession(ctx, userSession.ID, accessTokenClaims.GetTokenID(), newRefreshToken, accessTokenClaims.GetExpiredAt())
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	return models.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    accessTokenClaims.GetExpiredAt().Format(time.RFC3339),
	}, nil
}

func (u *usecase) ChangePassword(ctx context.Context, req models.ChangePasswordRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// get user ID from context
	userID := pkgCtx.GetUserID(ctx)
	if userID == "" {
		return pkgErr.InvalidRequest("user not found")
	}

	// get user from database
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.ID == "" {
		return pkgErr.InvalidRequest("user not found")
	}

	// check if user is active
	if user.Status != userConstants.UserStatusActive {
		return pkgErr.InvalidRequest("user account is inactive")
	}

	// verify old password
	ok, needsRehash := cryp.VerifyPassword(req.OldPassword, user.Password)
	if !ok {
		return pkgErr.InvalidRequest("old password is incorrect")
	}

	// update password
	if req.NewPassword != req.OldPassword {
		err = u.userRepo.SetPassword(ctx, userID, req.NewPassword)
		if err != nil {
			return err
		}
	} else if needsRehash {
		// upgrade legacy hash to the current scheme; best-effort, must not fail the request
		_ = u.userRepo.SetPassword(ctx, userID, req.OldPassword)
	}

	// revoke all user sessions
	err = u.userSessionRepo.InactiveAllUserSession(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (u *usecase) Logout(ctx context.Context) error {
	// get user ID and token ID from context
	userID := pkgCtx.GetUserID(ctx)
	tokenID := pkgCtx.GetTokenID(ctx)

	if userID == "" {
		return pkgErr.InvalidRequest("user not found")
	}
	if tokenID == "" {
		return pkgErr.InvalidRequest("token not found")
	}

	// invalidate the current session
	err := u.userSessionRepo.InactiveSessionByTokenID(ctx, tokenID)
	if err != nil {
		return err
	}

	return nil
}

func (u *usecase) RequestLogin(ctx context.Context, req models.RequestLoginRequest) (models.RequestLoginResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.RequestLoginResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// get app service by code
	appService, err := u.appServiceRepo.GetByCode(ctx, req.AppCode)
	if err != nil {
		return models.RequestLoginResponse{}, err
	}

	if appService.ID == "" {
		return models.RequestLoginResponse{}, pkgErr.InvalidRequest("app service not found")
	}

	// verify ctx_info matches
	if req.CtxInfo != appService.CtxInfo {
		return models.RequestLoginResponse{}, pkgErr.InvalidRequest("invalid ctx_info")
	}

	// decrypt app_secret from request and database
	decryptedAppSecret, err := aes.Decrypt(appService.AppSecret, u.cfg.AES.Secret, appService.CtxInfo)
	if err != nil {
		return models.RequestLoginResponse{}, err
	}
	if decryptedAppSecret != req.AppSecret {
		return models.RequestLoginResponse{}, pkgErr.InvalidRequest("invalid app_secret")
	}

	// generate session ID and set to cache
	sessionID := cryp.ULID()
	u.cache.Set(sessionID, appService.ID, time.Duration(u.cfg.Auth.ExternalLoginSessionTTL)*time.Second)

	// return response
	return models.RequestLoginResponse{
		RedirectURL: fmt.Sprintf("%s?session_id=%s", u.cfg.Auth.EndpointWebSSOLogin, sessionID),
	}, nil
}

func (u *usecase) ExchangeCode(ctx context.Context, req models.ExchangeCodeRequest) (models.ExchangeCodeResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.ExchangeCodeResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// get access token from cache
	accessTokenKey := keyAuthorizationCodeAccessToken(req.AuthorizationCode)
	accessToken, ok := u.cache.Get(accessTokenKey)
	if !ok {
		return models.ExchangeCodeResponse{}, pkgErr.InvalidRequest("invalid authorization code")
	}

	// get refresh token from cache
	refreshTokenKey := keyAuthorizationCodeRefreshToken(req.AuthorizationCode)
	refreshToken, ok := u.cache.Get(refreshTokenKey)
	if !ok {
		return models.ExchangeCodeResponse{}, pkgErr.InvalidRequest("invalid authorization code")
	}

	// get expires at from cache
	expiresAtKey := keyAuthorizationCodeExpiresAt(req.AuthorizationCode)
	expiresAt, ok := u.cache.Get(expiresAtKey)
	if !ok {
		return models.ExchangeCodeResponse{}, pkgErr.InvalidRequest("invalid authorization code")
	}

	// delete cache entries after successful retrieval (one-time use)
	u.cache.Delete(accessTokenKey)
	u.cache.Delete(refreshTokenKey)
	u.cache.Delete(expiresAtKey)

	// return response
	return models.ExchangeCodeResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// ssoConsentScopes are the static consent line items shown on the SSO consent
// screen. Rendered from data on the frontend, not hardcoded in TSX.
var ssoConsentScopes = []models.SSOScope{
	{Title: "View your profile & email", Description: "Your name and email address."},
	{Title: "Access your roles & permissions", Description: "Your assigned access for this application."},
	{Title: "Act on your behalf", Description: "Perform actions in the application as you."},
}

// SSOCheck is the READ-ONLY silent-authorize probe. It resolves the SSO session
// to its requesting app and validates the caller's existing isme tokens WITHOUT
// rotating them. On a valid session it returns the user identity, the requesting
// app, the static consent scopes, and a fresh single-use CSRF nonce. On an
// invalid/expired session it returns {valid:false} WITHOUT erroring, so the
// consent page can fall back to the password form.
func (u *usecase) SSOCheck(ctx context.Context, req models.SSOCheckRequest) (models.SSOCheckResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.SSOCheckResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// resolve session_id → requesting app service (same as Login)
	appServiceID, ok := u.cache.Get(req.SessionID)
	if !ok {
		return models.SSOCheckResponse{}, pkgErr.InvalidRequest("invalid session_id")
	}
	appService, err := u.appServiceRepo.GetByID(ctx, appServiceID)
	if err != nil {
		return models.SSOCheckResponse{}, err
	}
	if appService.ID == "" {
		return models.SSOCheckResponse{}, pkgErr.InvalidRequest("invalid session_id")
	}

	// app identity is always returned (even when invalid) so the password form
	// can render "continue to <app>" instead of a generic placeholder
	app := models.SSOCheckApp{
		Name:        appService.AppName,
		RedirectURL: appService.RedirectURL,
		AppCode:     appService.AppCode,
		Icon:        appService.Icon,
		Color:       appService.Color,
	}

	// read-only validity probe (NO rotation, NO session mutation)
	user, valid := u.validateSessionForConsent(ctx, req.AccessToken, req.RefreshToken)
	if !valid {
		// not an error: the page renders the password form instead, still
		// carrying the app name so the header reads "continue to <app>"
		return models.SSOCheckResponse{Valid: false, App: app}, nil
	}

	// mint a fresh single-use nonce that /sso/consent will require
	nonce := u.mintConsentNonce(req.SessionID)

	return models.SSOCheckResponse{
		Valid: true,
		User: models.SSOCheckUser{
			Name:  user.Name,
			Email: user.Email,
		},
		App:    app,
		Scopes: ssoConsentScopes,
		Nonce:  nonce,
	}, nil
}

// SSOConsent is the authorize step of silent SSO. It re-validates everything
// server-side independently (never trusting a prior /sso/check), consumes the
// single-use nonce, re-resolves the requesting app from the session_id, and
// mints a FRESH app-scoped token pair bound to a NEW user_session for the
// requesting app — exactly like an SSO login would, minus password verification.
// It never reads or mutates the caller's own (isme browser) session, so the
// browser stays logged in, and the issued token is aud-restricted to the
// requesting app via buildTokenScope(appService.AppCode).
func (u *usecase) SSOConsent(ctx context.Context, req models.SSOConsentRequest) (models.SSOConsentResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.SSOConsentResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// resolve session_id → requesting app service (independent re-resolution)
	appServiceID, ok := u.cache.Get(req.SessionID)
	if !ok {
		return models.SSOConsentResponse{}, pkgErr.InvalidRequest("invalid session_id")
	}
	appService, err := u.appServiceRepo.GetByID(ctx, appServiceID)
	if err != nil {
		return models.SSOConsentResponse{}, err
	}
	if appService.ID == "" {
		return models.SSOConsentResponse{}, pkgErr.InvalidRequest("invalid session_id")
	}

	// re-run the read-only validity probe server-side — never trust prior /check.
	// capture the user so we can mint a fresh app-scoped session for them below.
	user, valid := u.validateSessionForConsent(ctx, req.AccessToken, req.RefreshToken)
	if !valid {
		return models.SSOConsentResponse{}, pkgErr.InvalidRequest("session is no longer valid")
	}

	// validate + consume the single-use CSRF nonce (replay guard)
	if !u.consumeConsentNonce(req.SessionID, req.Nonce) {
		return models.SSOConsentResponse{}, pkgErr.InvalidRequest("invalid or expired nonce")
	}

	// mint a fresh app-scoped token pair bound to a NEW session for the
	// requesting app — mirrors the Login SSO branch, minus password checks.
	// The caller's own browser session is never read or mutated here.

	// re-load authorization data from the DB so revoked rights never leak
	groupedPerms, err := u.roleRepo.GetPermissionCodesGroupedByApp(ctx, user.ID)
	if err != nil {
		return models.SSOConsentResponse{}, err
	}

	// aud-restrict the token to the requesting app (medioa2), not a full token
	resourceAccess, audience := buildTokenScope(groupedPerms, appService.AppCode)

	// generate access tokens
	accessToken, accessTokenClaims, err := u.generateAccessTokens(user.ID, user.Email, resourceAccess, audience)
	if err != nil {
		return models.SSOConsentResponse{}, err
	}
	expiresAt := accessTokenClaims.GetExpiredAt().Format(time.RFC3339)

	// generate refresh tokens
	refreshToken, _, err := u.generateRefreshTokens(user.ID, user.Email)
	if err != nil {
		return models.SSOConsentResponse{}, err
	}

	// create a NEW user session bound to the requesting app (the medioa2
	// session); the browser's isme session row is untouched
	_, err = u.createUserSession(ctx, user.ID, accessTokenClaims.GetTokenID(), user.Email, refreshToken, appServiceID, accessTokenClaims.GetExpiredAt())
	if err != nil {
		return models.SSOConsentResponse{}, err
	}

	// mint a one-time authorization code (deletes session_id atomically)
	authorizationCode := u.mintAuthorizationCode(accessToken, refreshToken, expiresAt, req.SessionID)

	return models.SSOConsentResponse{
		RedirectURL:       appService.RedirectURL,
		AuthorizationCode: authorizationCode,
	}, nil
}

func keyAuthorizationCodeAccessToken(authorizationCode string) string {
	return fmt.Sprintf("auth:external:code:%s:access_token", authorizationCode)
}

func keyAuthorizationCodeRefreshToken(authorizationCode string) string {
	return fmt.Sprintf("auth:external:code:%s:refresh_token", authorizationCode)
}

func keyAuthorizationCodeExpiresAt(authorizationCode string) string {
	return fmt.Sprintf("auth:external:code:%s:expires_at", authorizationCode)
}
