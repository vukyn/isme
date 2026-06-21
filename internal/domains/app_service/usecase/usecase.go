package usecase

import (
	"context"
	"time"

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/domains/app_service/constants"
	"github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/app_service/models"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleUsecase "github.com/vukyn/isme/internal/domains/role/usecase"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	"github.com/vukyn/kuery/cryp/aes"
	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
)

type usecase struct {
	cfg            *config.Config
	appServiceRepo appServiceRepo.IRepository
	userRepo       userRepo.IRepository
	roleUsecase    roleUsecase.IUseCase
}

func NewUsecase(
	appServiceRepo appServiceRepo.IRepository,
	userRepo userRepo.IRepository,
	roleUsecase roleUsecase.IUseCase,
	cfg *config.Config,
) IUseCase {
	return &usecase{
		cfg:            cfg,
		appServiceRepo: appServiceRepo,
		userRepo:       userRepo,
		roleUsecase:    roleUsecase,
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
	appServiceID, err := u.appServiceRepo.Create(ctx, entity.CreateRequest{
		AppCode:     req.AppCode,
		AppName:     req.AppName,
		AppSecret:   encryptedSecret,
		RedirectURL: req.RedirectURL,
		CtxInfo:     req.CtxInfo,
		Status:      constants.AppServiceStatusActive,
		Icon:        req.Icon,
		Color:       req.Color,
	})
	if err != nil {
		return models.RegisterResponse{}, err
	}

	// auto-provision the default per-app role set (admin role + CRUD perm seed)
	if err := u.roleUsecase.ProvisionDefaultRoles(ctx, appServiceID); err != nil {
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

	// only active app services can be verified
	if appService.Status != constants.AppServiceStatusActive {
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
	userID := pkgCtx.GetUserID(ctx)
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

	// the isme platform app is read-only — its secret cannot be rotated
	if constants.IsPlatformApp(appService.ID) {
		return models.RefreshResponse{}, pkgErr.Forbidden("the isme platform app is read-only and cannot be modified")
	}

	// terminated app services cannot be refreshed
	if appService.Status == constants.AppServiceStatusTerminated {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("app service is terminated")
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

	// check authorization: caller must be the app creator. Permission-level
	// access (app_service:rotate_secret) is enforced at the route guard; this
	// ownership check additionally restricts rotation to the creator.
	if userID != appService.CreatedBy {
		return models.RefreshResponse{}, pkgErr.InvalidRequest("unauthorized: only the creator can refresh app secret")
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

func (u *usecase) ListApps(ctx context.Context, req models.ListRequest) (models.ListResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.ListResponse{}, pkgErr.InvalidRequest(err.Error())
	}
	req.Normalize()

	appServices, total, err := u.appServiceRepo.List(ctx, req)
	if err != nil {
		return models.ListResponse{}, err
	}

	// resolve creator emails once per distinct id
	creatorEmails := make(map[string]string)
	for _, appService := range appServices {
		if appService.CreatedBy == "" {
			continue
		}
		if _, resolved := creatorEmails[appService.CreatedBy]; resolved {
			continue
		}
		creator, err := u.userRepo.GetByID(ctx, appService.CreatedBy)
		if err != nil {
			return models.ListResponse{}, err
		}
		creatorEmails[appService.CreatedBy] = creator.Email
	}

	items := make([]models.AppServiceListItem, 0, len(appServices))
	for _, appService := range appServices {
		updatedAt := ""
		if !appService.UpdatedAt.IsZero() {
			updatedAt = appService.UpdatedAt.Format(time.RFC3339)
		}
		createdAt := ""
		if !appService.CreatedAt.IsZero() {
			createdAt = appService.CreatedAt.Format(time.RFC3339)
		}
		items = append(items, models.AppServiceListItem{
			ID:             appService.ID,
			AppCode:        appService.AppCode,
			AppName:        appService.AppName,
			RedirectURL:    appService.RedirectURL,
			CtxInfo:        appService.CtxInfo,
			Status:         appService.Status,
			Icon:           appService.Icon,
			Color:          appService.Color,
			CreatedAt:      createdAt,
			CreatedBy:      appService.CreatedBy,
			CreatedByEmail: creatorEmails[appService.CreatedBy],
			UpdatedAt:      updatedAt,
			UpdatedBy:      appService.UpdatedBy,
		})
	}

	return models.ListResponse{
		Items: items,
		Total: total,
		Page:  req.Page,
	}, nil
}

func (u *usecase) UpdateStatus(ctx context.Context, id string, req models.UpdateStatusRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// check app service exists
	appService, err := u.appServiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if appService.ID == "" {
		return pkgErr.NotFound("app service not found")
	}

	// the isme platform app is read-only — block at the usecase so the API/URL
	// can't bypass the disabled UI controls
	if constants.IsPlatformApp(appService.ID) {
		return pkgErr.Forbidden("the isme platform app is read-only and cannot be modified")
	}

	// terminated is a terminal state
	if appService.Status == constants.AppServiceStatusTerminated {
		return pkgErr.InvalidRequest("app service is terminated")
	}

	// idempotent no-op when status is unchanged
	if appService.Status == req.Status {
		return nil
	}

	return u.appServiceRepo.UpdateStatus(ctx, id, req.Status)
}

func (u *usecase) GetApp(ctx context.Context, id string) (models.AppServiceListItem, error) {
	appService, err := u.appServiceRepo.GetByID(ctx, id)
	if err != nil {
		return models.AppServiceListItem{}, err
	}
	if appService.ID == "" {
		return models.AppServiceListItem{}, pkgErr.NotFound("app service not found")
	}

	// resolve creator email (best-effort, mirrors ListApps)
	creatorEmail := ""
	if appService.CreatedBy != "" {
		creator, err := u.userRepo.GetByID(ctx, appService.CreatedBy)
		if err != nil {
			return models.AppServiceListItem{}, err
		}
		creatorEmail = creator.Email
	}

	updatedAt := ""
	if !appService.UpdatedAt.IsZero() {
		updatedAt = appService.UpdatedAt.Format(time.RFC3339)
	}
	createdAt := ""
	if !appService.CreatedAt.IsZero() {
		createdAt = appService.CreatedAt.Format(time.RFC3339)
	}

	return models.AppServiceListItem{
		ID:             appService.ID,
		AppCode:        appService.AppCode,
		AppName:        appService.AppName,
		RedirectURL:    appService.RedirectURL,
		CtxInfo:        appService.CtxInfo,
		Status:         appService.Status,
		Icon:           appService.Icon,
		Color:          appService.Color,
		CreatedAt:      createdAt,
		CreatedBy:      appService.CreatedBy,
		CreatedByEmail: creatorEmail,
		UpdatedAt:      updatedAt,
		UpdatedBy:      appService.UpdatedBy,
	}, nil
}

func (u *usecase) UpdateAppearance(ctx context.Context, id string, req models.UpdateAppearanceRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// check app service exists (appearance edits are allowed regardless of status)
	appService, err := u.appServiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if appService.ID == "" {
		return pkgErr.NotFound("app service not found")
	}

	// the isme platform app is read-only — block at the usecase so the API/URL
	// can't bypass the disabled UI controls
	if constants.IsPlatformApp(appService.ID) {
		return pkgErr.Forbidden("the isme platform app is read-only and cannot be modified")
	}

	return u.appServiceRepo.Update(ctx, entity.UpdateRequest{
		ID:          id,
		AppName:     req.AppName,
		RedirectURL: req.RedirectURL,
		Icon:        req.Icon,
		Color:       req.Color,
	})
}
