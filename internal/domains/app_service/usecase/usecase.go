package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/domains/app_service/constants"
	"github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/app_service/models"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	"github.com/vukyn/kuery/cryp/aes"
	pkgCtx "github.com/vukyn/isme/pkg/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
)

type usecase struct {
	cfg            *config.Config
	appServiceRepo appServiceRepo.IRepository
	userRepo       userRepo.IRepository
}

func NewUsecase(
	appServiceRepo appServiceRepo.IRepository,
	userRepo userRepo.IRepository,
	cfg *config.Config,
) IUseCase {
	return &usecase{
		cfg:            cfg,
		appServiceRepo: appServiceRepo,
		userRepo:       userRepo,
	}
}

func (u *usecase) RegisterApp(ctx context.Context, req models.RegisterRequest) (models.RegisterResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.RegisterResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// check if app_code already exists
	existingApp, err := u.appServiceRepo.GetByCode(ctx, req.AppCode)
	if err != nil {
		return models.RegisterResponse{}, err
	}
	if existingApp.ID != "" {
		return models.RegisterResponse{}, pkgErr.InvalidRequest("app_code already exists")
	}

	// validate ctx_info
	if _, ok := constants.AllowedCtxInfos[req.CtxInfo]; !ok {
		return models.RegisterResponse{}, pkgErr.InvalidRequest("invalid ctx_info")
	}

	// generate app_secret
	appSecret, encryptedSecret, err := generateAndEncryptAppSecret(u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.RegisterResponse{}, err
	}

	// create app service
	_, err = u.appServiceRepo.Create(ctx, entity.CreateRequest{
		AppCode:     req.AppCode,
		AppName:     req.AppName,
		AppSecret:   encryptedSecret,
		RedirectURL: req.RedirectURL,
		CtxInfo:     req.CtxInfo,
		Status:      constants.AppServiceStatusActive,
	})
	if err != nil {
		return models.RegisterResponse{}, err
	}

	return models.RegisterResponse{
		AppSecret: appSecret,
	}, nil
}

func (u *usecase) VerifyApp(ctx context.Context, req models.VerifyRequest) (models.VerifyResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.VerifyResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// get app service by code
	appService, err := u.appServiceRepo.GetByCode(ctx, req.AppCode)
	if err != nil {
		return models.VerifyResponse{}, err
	}

	// check if app_code is valid
	if appService.ID == "" {
		return models.VerifyResponse{
			Ok: false,
		}, nil
	}

	// verify ctx_info matches
	if req.CtxInfo != appService.CtxInfo {
		return models.VerifyResponse{
			Ok: false,
		}, nil
	}

	// decrypt app_secret from database
	decryptedAppSecret, err := aes.Decrypt(appService.AppSecret, u.cfg.AES.Secret, appService.CtxInfo)
	if err != nil {
		return models.VerifyResponse{
			Ok: false,
		}, nil
	}
	if decryptedAppSecret != req.AppSecret {
		return models.VerifyResponse{
			Ok: false,
		}, nil
	}

	return models.VerifyResponse{
		Ok: true,
	}, nil
}

func (u *usecase) RefreshApp(ctx context.Context, req models.RefreshRequest) (models.RefreshResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.RefreshResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// get user ID from context
	userID := pkgCtx.GetUserId(ctx)
	if userID == "" {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("user not authenticated")
	}

	// get app service by code
	appService, err := u.appServiceRepo.GetByCode(ctx, req.AppCode)
	if err != nil {
		return models.RefreshResponse{}, err
	}

	// check if app_code is valid
	if appService.ID == "" {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("app_code not found")
	}

	// verify ctx_info matches
	if req.CtxInfo != appService.CtxInfo {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("invalid ctx_info")
	}

	// decrypt app_secret from request and database
	ok, err := compareAppSecret(req.AppSecret, appService.AppSecret, u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.RefreshResponse{}, err
	}
	if !ok {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("invalid app_secret")
	}

	// check authorization: user must be admin OR creator
	isAdmin, err := u.userRepo.IsAdmin(ctx, userID)
	if err != nil {
		return models.RefreshResponse{}, err
	}
	if !isAdmin && userID != appService.CreatedBy {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("unauthorized: only admin or creator can refresh app secret")
	}

	// generate new app_secret
	appSecret, encryptedSecret, err := generateAndEncryptAppSecret(u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.RefreshResponse{}, err
	}

	// update app service with new secret
	err = u.appServiceRepo.Update(ctx, entity.UpdateRequest{
		ID:        appService.ID,
		AppSecret: &encryptedSecret,
	})
	if err != nil {
		return models.RefreshResponse{}, err
	}

	return models.RefreshResponse{
		AppSecret: appSecret,
	}, nil
}
