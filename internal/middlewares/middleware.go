package middlewares

import (
	"isme/internal/config"
	authUC "isme/internal/domains/auth/usecase"
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
