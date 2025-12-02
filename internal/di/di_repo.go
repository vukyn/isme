package di

import (
	"isme/internal/constants"
	userRepo "isme/internal/domains/user/repository"
	userSessionRepo "isme/internal/domains/user_session/repository"

	"github.com/sarulabs/di/v2"
	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/log"
)

func defineRepository() []*di.Def {
	return []*di.Def{
		defineUserRepository(),
		defineUserSessionRepository(),
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
