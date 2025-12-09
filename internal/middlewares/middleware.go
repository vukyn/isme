package middlewares

import (
	"github.com/vukyn/isme/internal/config"
	authUC "github.com/vukyn/isme/internal/domains/auth/usecase"
)

type Middleware struct {
	cfg    *config.Config
	authUC authUC.IUseCase
}

func NewMiddleware(
	cfg *config.Config,
	authUC authUC.IUseCase,
) *Middleware {
	return &Middleware{
		cfg:    cfg,
		authUC: authUC,
	}
}
