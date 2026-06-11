package di

import (
	"github.com/vukyn/isme/internal/constants"
	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	userInvitationRepo "github.com/vukyn/isme/internal/domains/user_invitation/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	"github.com/sarulabs/di/v2"
	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/log"
)

func defineRepository() []*di.Def {
	return []*di.Def{
		defineUserRepository(),
		defineUserSessionRepository(),
		defineAppServiceRepository(),
		defineRoleRepository(),
		defineUserInvitationRepository(),
		defineSettingsRepository(),
		defineActivityRepository(),
	}
}

func defineUserRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_USER_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("User repository initialized")
			return userRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("User repository destroyed")
			return nil
		},
	}
	return def
}

func GetUserRepository(ctn di.Container) (userRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_USER_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(userRepo.IRepository), nil
}

func defineUserSessionRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_USER_SESSION_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("User session repository initialized")
			return userSessionRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("User session repository destroyed")
			return nil
		},
	}
	return def
}

func GetUserSessionRepository(ctn di.Container) (userSessionRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_USER_SESSION_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(userSessionRepo.IRepository), nil
}

func defineAppServiceRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_APP_SERVICE_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("App service repository initialized")
			return appServiceRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("App service repository destroyed")
			return nil
		},
	}
	return def
}

func GetAppServiceRepository(ctn di.Container) (appServiceRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_APP_SERVICE_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(appServiceRepo.IRepository), nil
}

func defineRoleRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_ROLE_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("Role repository initialized")
			return roleRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Role repository destroyed")
			return nil
		},
	}
	return def
}

func GetRoleRepository(ctn di.Container) (roleRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_ROLE_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(roleRepo.IRepository), nil
}

func defineUserInvitationRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_USER_INVITATION_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("User invitation repository initialized")
			return userInvitationRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("User invitation repository destroyed")
			return nil
		},
	}
	return def
}

func GetUserInvitationRepository(ctn di.Container) (userInvitationRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_USER_INVITATION_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(userInvitationRepo.IRepository), nil
}

func defineSettingsRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_SETTINGS_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("Settings repository initialized")
			return settingsRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Settings repository destroyed")
			return nil
		},
	}
	return def
}

func GetSettingsRepository(ctn di.Container) (settingsRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_SETTINGS_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(settingsRepo.IRepository), nil
}

func defineActivityRepository() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_ACTIVITY_REPOSITORY,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			log.New().Debug("Activity repository initialized")
			return activityRepo.NewRepository(db), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Activity repository destroyed")
			return nil
		},
	}
	return def
}

func GetActivityRepository(ctn di.Container) (activityRepo.IRepository, error) {
	repo, err := ctn.SafeGet(constants.CONTAINER_NAME_ACTIVITY_REPOSITORY)
	if err != nil {
		return nil, err
	}
	return repo.(activityRepo.IRepository), nil
}
