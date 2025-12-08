package usecase

import (
	"context"
	"isme/internal/config"
	"isme/internal/domains/app_service/constants"
	"isme/internal/domains/app_service/entity"
	"isme/internal/domains/app_service/models"
	appServiceRepo "isme/internal/domains/app_service/repository"
	"isme/pkg/cryp/aes"
	"isme/pkg/cryp/rand"
	pkgErr "isme/pkg/http/errors"
)

type usecase struct {
	cfg            *config.Config
	appServiceRepo appServiceRepo.IRepository
}

func NewUsecase(
	appServiceRepo appServiceRepo.IRepository,
	cfg *config.Config,
) IUseCase {
	return &usecase{
		cfg:            cfg,
		appServiceRepo: appServiceRepo,
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
	appSecret := rand.RandMixedString(8, true, true)
	if appSecret == "" {
		return models.RegisterResponse{}, pkgErr.InvalidRequest("failed to generate app_secret")
	}

	// encrypt app_secret with AES
	encryptedSecret, err := aes.Encrypt(appSecret, u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.RegisterResponse{}, pkgErr.InvalidRequest("failed to encrypt app_secret: " + err.Error())
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
		AppSecret: encryptedSecret,
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

	// decrypt app_secret from request and database
	decryptedSecret1, err := aes.Decrypt(req.AppSecret, u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.VerifyResponse{
			Ok: false,
		}, nil
	}
	decryptedSecret2, err := aes.Decrypt(appService.AppSecret, u.cfg.AES.Secret, req.CtxInfo)
	if err != nil {
		return models.VerifyResponse{
			Ok: false,
		}, nil
	}
	if decryptedSecret1 != decryptedSecret2 {
		return models.VerifyResponse{
			Ok: false,
		}, pkgErr.InvalidRequest("invalid app_secret")
	}

	return models.VerifyResponse{
		Ok: true,
	}, nil
}
