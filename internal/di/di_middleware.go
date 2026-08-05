package di

import (
	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/constants"
	"github.com/vukyn/isme/internal/middlewares"

	"github.com/sarulabs/di/v2"
	"github.com/vukyn/kuery/log"
)

// defineMiddleware builds the app-scoped middleware singleton (the auth
// middleware's home). It holds the auth usecase for the whole process, which is
// why it needs a sub-container of its own.
//
// ⚠️ THE SUB-CONTAINER BELOW IS DELIBERATELY NEVER DELETED — do not "fix" it.
// The auth usecase is a `di.Request`-scoped definition, and a Request-scoped
// object cannot be resolved straight out of an `App` container; the sub-container
// exists only to make that resolution legal. The Middleware then RETAINS the
// resolved usecase for the life of the process, so deleting the container it came
// from would run that usecase's Close and tear down an object the middleware keeps
// calling on every authenticated request. Its lifetime is intentionally the
// process's: exactly ONE permanently-retained container, created once at boot.
//
// That is a different thing from the per-request leak fixed in
// middlewares.DiContainerMiddleware, which created one sub-container PER REQUEST
// and released none. A single boot-time container is bounded; one per request is
// not. `cmd/main.go` shuts the tree down with DeleteWithSubContainers precisely
// because this child means a plain `Delete()` on the app container would be a
// no-op (sarulabs/di only closes a parent whose children map is empty).
func defineMiddleware() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_MIDDLEWARE,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(constants.CONTAINER_NAME_CONFIG).(*config.Config)
			// Intentionally retained for the process lifetime — see the note above.
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
