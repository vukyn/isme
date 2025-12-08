package usecase

import (
	"context"
	"fmt"
	"isme/cache"
	"isme/internal/config"
	"isme/internal/constants"
	appServiceRepo "isme/internal/domains/app_service/repository"
	"isme/internal/domains/auth/models"
	userConstants "isme/internal/domains/user/constants"
	userModels "isme/internal/domains/user/models"
	userRepo "isme/internal/domains/user/repository"
	userSessionConstants "isme/internal/domains/user_session/constants"
	userSessionRepo "isme/internal/domains/user_session/repository"
	pkgClaims "isme/pkg/claims"
	"isme/pkg/cryp/aes"
	pkgCtx "isme/pkg/ctx"
	pkgErr "isme/pkg/http/errors"
	"isme/pkg/jwt"
	"time"

	"github.com/vukyn/kuery/cryp"
)

type usecase struct {
	cfg             *config.Config
	cache           *cache.Cache
	userRepo        userRepo.IRepository
	userSessionRepo userSessionRepo.IRepository
	appServiceRepo  appServiceRepo.IRepository
}

func NewUsecase(
	cfg *config.Config,
	cache *cache.Cache,
	userRepo userRepo.IRepository,
	userSessionRepo userSessionRepo.IRepository,
	appServiceRepo appServiceRepo.IRepository,
) IUseCase {
	return &usecase{
		cfg:             cfg,
		cache:           cache,
		userRepo:        userRepo,
		userSessionRepo: userSessionRepo,
		appServiceRepo:  appServiceRepo,
	}
}

func (u *usecase) GetMe(ctx context.Context) (models.GetMeResponse, error) {
	userId := pkgCtx.GetUserId(ctx)
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

func (u *usecase) SignUp(ctx context.Context, req models.SignUpRequest) (models.SignUpResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.SignUpResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// check if user already exists
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return models.SignUpResponse{}, err
	}
	if user.ID != "" {
		return models.SignUpResponse{}, pkgErr.InvalidRequest("user already exists")
	}

	// create user
	userID, err := u.userRepo.Create(ctx, userModels.CreateRequest{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		return models.SignUpResponse{}, err
	}

	// set user password
	err = u.userRepo.SetPassword(ctx, userID, req.Password)
	if err != nil {
		return models.SignUpResponse{}, err
	}

	// return response
	return models.SignUpResponse{
		ID: userID,
	}, nil
}

func (u *usecase) Login(ctx context.Context, req models.LoginRequest) (models.LoginResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.LoginResponse{}, pkgErr.InvalidRequest(err.Error())
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
	if !cryp.CompareBcrypt(req.Password, user.Password) {
		return models.LoginResponse{}, pkgErr.InvalidRequest("invalid email or password")
	}

	// generate access tokens
	accessToken, accessTokenClaims, err := u.generateAccessTokens(user.ID, user.Email)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// generate refresh tokens
	refreshToken, _, err := u.generateRefreshTokens(user.ID, user.Email)
	if err != nil {
		return models.LoginResponse{}, err
	}

	// create user session
	_, err = u.createUserSession(ctx, user.ID, accessTokenClaims.GetTokenID(), user.Email, refreshToken, accessTokenClaims.GetExpiredAt())
	if err != nil {
		return models.LoginResponse{}, err
	}

	// update user last login
	err = u.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessTokenClaims.GetExpiredAt().Format(time.RFC3339),
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

	// generate new access tokens
	newAccessToken, accessTokenClaims, err := u.generateAccessTokens(user.ID, user.Email)
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
	userID := pkgCtx.GetUserId(ctx)
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
	if !cryp.CompareBcrypt(req.OldPassword, user.Password) {
		return pkgErr.InvalidRequest("old password is incorrect")
	}

	// update password
	if req.NewPassword != req.OldPassword {
		err = u.userRepo.SetPassword(ctx, userID, req.NewPassword)
		if err != nil {
			return err
		}
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
	userID := pkgCtx.GetUserId(ctx)
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
	decryptedSecret, err := aes.Decrypt(req.AppSecret, u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.RequestLoginResponse{}, err
	}
	decryptedSecret2, err := aes.Decrypt(appService.AppSecret, u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.RequestLoginResponse{}, err
	}
	if decryptedSecret != decryptedSecret2 {
		return models.RequestLoginResponse{}, pkgErr.InvalidRequest("invalid app_secret")
	}

	// generate session ID and set to cache
	sessionID := cryp.ULID()
	u.cache.Set(sessionID, appService.ID, time.Duration(u.cfg.Auth.ExternalAuthSessionTTL)*time.Minute)

	// return response
	return models.RequestLoginResponse{
		RedirectURL: fmt.Sprintf("%s?session_id=%s", constants.AUTH_WEB_ENDPOINT_SSO_LOGIN, sessionID),
	}, nil
}
