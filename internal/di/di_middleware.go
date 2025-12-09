package di

import (
	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/constants"
	"github.com/vukyn/isme/internal/middlewares"

	"github.com/sarulabs/di/v2"
	"github.com/vukyn/kuery/log"
)

func defineMiddleware() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_MIDDLEWARE,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			subCtn, err := ctn.SubContainer()
			if err != nil {
				return nil, err
			}
			authUC, err := GetAuthUsecase(subCtn)
			if err != nil {
				return nil, err
			}
			log.New().Info("Middleware initialized")
			return middlewares.NewMiddleware(cfg, authUC), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Middleware destroyed")
			return nil
		},
	}
	return def
}

func GetMiddleware(ctn di.Container) *middlewares.Middleware {
	return ctn.Get(constants.CONTAINER_NAME_MIDDLEWARE).(*middlewares.Middleware)
}
