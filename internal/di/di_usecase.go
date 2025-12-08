package di

import (
	"isme/cache"
	"isme/internal/config"
	"isme/internal/constants"
	appServiceUsecase "isme/internal/domains/app_service/usecase"
	authUsecase "isme/internal/domains/auth/usecase"

	"github.com/sarulabs/di/v2"
	"github.com/vukyn/kuery/log"
)

func defineUsecase() []*di.Def {
	return []*di.Def{
		defineAuthUsecase(),
		defineAppServiceUsecase(),
	}
}

func defineAuthUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_AUTH_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			userRepo, err := GetUserRepository(ctn)
			if err != nil {
				return nil, err
			}
			userSessionRepo, err := GetUserSessionRepository(ctn)
			if err != nil {
				return nil, err
			}
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			cache := ctn.Get(constants.CONTAINER_NAME_CACHE).(*cache.Cache)
			appServiceRepo, err := GetAppServiceRepository(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("Auth usecase initialized")
			return authUsecase.NewUsecase(cfg, cache, userRepo, userSessionRepo, appServiceRepo), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Auth usecase destroyed")
			return nil
		},
	}
	return def
}

func GetAuthUsecase(ctn di.Container) (authUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_AUTH_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(authUsecase.IUseCase), nil
}

func defineAppServiceUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_APP_SERVICE_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			appServiceRepo, err := GetAppServiceRepository(ctn)
			if err != nil {
				return nil, err
			}
			userRepo, err := GetUserRepository(ctn)
			if err != nil {
				return nil, err
			}
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			log.New().Debug("App service usecase initialized")
			return appServiceUsecase.NewUsecase(appServiceRepo, userRepo, cfg), nil
		},
		Close: func(obj any) error {
			log.New().Debug("App service usecase destroyed")
			return nil
		},
	}
	return def
}

func GetAppServiceUsecase(ctn di.Container) (appServiceUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_APP_SERVICE_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(appServiceUsecase.IUseCase), nil
}
