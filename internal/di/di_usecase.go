package di

import (
	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/constants"
	activityUsecase "github.com/vukyn/isme/internal/domains/activity/usecase"
	appServiceUsecase "github.com/vukyn/isme/internal/domains/app_service/usecase"
	authUsecase "github.com/vukyn/isme/internal/domains/auth/usecase"
	mediaUsecase "github.com/vukyn/isme/internal/domains/media/usecase"
	roleUsecase "github.com/vukyn/isme/internal/domains/role/usecase"
	settingsUsecase "github.com/vukyn/isme/internal/domains/settings/usecase"
	userUsecase "github.com/vukyn/isme/internal/domains/user/usecase"
	userInvitationUsecase "github.com/vukyn/isme/internal/domains/user_invitation/usecase"

	"github.com/sarulabs/di/v2"
	"github.com/vukyn/kuery/log"
	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

func defineUsecase() []*di.Def {
	return []*di.Def{
		defineActivityUsecase(),
		defineAuthUsecase(),
		defineAppServiceUsecase(),
		defineUserUsecase(),
		defineRoleUsecase(),
		defineUserInvitationUsecase(),
		defineSettingsUsecase(),
		defineMediaUsecase(),
	}
}

func defineActivityUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_ACTIVITY_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			activityRepo, err := GetActivityRepository(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("Activity usecase initialized")
			return activityUsecase.NewUsecase(activityRepo), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Activity usecase destroyed")
			return nil
		},
	}
	return def
}

func GetActivityUsecase(ctn di.Container) (activityUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_ACTIVITY_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(activityUsecase.IUseCase), nil
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
			cache := GetCache(ctn)
			appServiceRepo, err := GetAppServiceRepository(ctn)
			if err != nil {
				return nil, err
			}
			roleRepo, err := GetRoleRepository(ctn)
			if err != nil {
				return nil, err
			}
			activityUsecase, err := GetActivityUsecase(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("Auth usecase initialized")
			return authUsecase.NewUsecase(cfg, cache, userRepo, userSessionRepo, appServiceRepo, roleRepo, activityUsecase), nil
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
			roleUsecase, err := GetRoleUsecase(ctn)
			if err != nil {
				return nil, err
			}
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			log.New().Debug("App service usecase initialized")
			return appServiceUsecase.NewUsecase(appServiceRepo, userRepo, roleUsecase, cfg), nil
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

func defineUserUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_USER_USECASE,
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
			roleRepo, err := GetRoleRepository(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("User usecase initialized")
			return userUsecase.NewUsecase(userRepo, userSessionRepo, roleRepo), nil
		},
		Close: func(obj any) error {
			log.New().Debug("User usecase destroyed")
			return nil
		},
	}
	return def
}

func GetUserUsecase(ctn di.Container) (userUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_USER_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(userUsecase.IUseCase), nil
}

func defineRoleUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_ROLE_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			roleRepo, err := GetRoleRepository(ctn)
			if err != nil {
				return nil, err
			}
			userRepo, err := GetUserRepository(ctn)
			if err != nil {
				return nil, err
			}
			appServiceRepo, err := GetAppServiceRepository(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("Role usecase initialized")
			return roleUsecase.NewUsecase(roleRepo, userRepo, appServiceRepo), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Role usecase destroyed")
			return nil
		},
	}
	return def
}

func GetRoleUsecase(ctn di.Container) (roleUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_ROLE_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(roleUsecase.IUseCase), nil
}

func defineUserInvitationUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_USER_INVITATION_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			userInvitationRepo, err := GetUserInvitationRepository(ctn)
			if err != nil {
				return nil, err
			}
			userRepo, err := GetUserRepository(ctn)
			if err != nil {
				return nil, err
			}
			roleRepo, err := GetRoleRepository(ctn)
			if err != nil {
				return nil, err
			}
			appServiceRepo, err := GetAppServiceRepository(ctn)
			if err != nil {
				return nil, err
			}
			activityUsecase, err := GetActivityUsecase(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("User invitation usecase initialized")
			return userInvitationUsecase.NewUsecase(cfg, userInvitationRepo, userRepo, roleRepo, appServiceRepo, activityUsecase), nil
		},
		Close: func(obj any) error {
			log.New().Debug("User invitation usecase destroyed")
			return nil
		},
	}
	return def
}

func GetUserInvitationUsecase(ctn di.Container) (userInvitationUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_USER_INVITATION_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(userInvitationUsecase.IUseCase), nil
}

func defineSettingsUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_SETTINGS_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			settingsRepo, err := GetSettingsRepository(ctn)
			if err != nil {
				return nil, err
			}
			// the app-scoped scheduler engine singleton acts as the IReloader
			reloader := ctn.Get(constants.CONTAINER_NAME_SCHEDULER).(*pkgScheduler.Engine)
			log.New().Debug("Settings usecase initialized")
			return settingsUsecase.NewUsecase(settingsRepo, reloader), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Settings usecase destroyed")
			return nil
		},
	}
	return def
}

func GetSettingsUsecase(ctn di.Container) (settingsUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_SETTINGS_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(settingsUsecase.IUseCase), nil
}

func defineMediaUsecase() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_MEDIA_USECASE,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			// medioaClient may be a typed-nil when MEDIOA_API_KEY is unset; the
			// usecase guards on it and returns a 502.
			medioaClient, err := GetMedioaClient(ctn)
			if err != nil {
				return nil, err
			}
			log.New().Debug("Media usecase initialized")
			return mediaUsecase.NewUsecase(cfg, medioaClient), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Media usecase destroyed")
			return nil
		},
	}
	return def
}

func GetMediaUsecase(ctn di.Container) (mediaUsecase.IUseCase, error) {
	uc, err := ctn.SafeGet(constants.CONTAINER_NAME_MEDIA_USECASE)
	if err != nil {
		return nil, err
	}
	return uc.(mediaUsecase.IUseCase), nil
}
