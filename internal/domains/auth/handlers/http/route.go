package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)
	r := router.Group(constants.AUTH_GROUP_NAME)
	r.Post(constants.AUTH_ENDPOINT_LOGIN, Login)
	r.Post(constants.AUTH_ENDPOINT_REFRESH, RefreshToken)
	r.Get(constants.AUTH_ENDPOINT_ME, middleware.AuthMiddleware, GetMe)
	r.Post(constants.AUTH_ENDPOINT_CHANGE_PASSWORD, middleware.AuthMiddleware, ChangePassword)
	r.Post(constants.AUTH_ENDPOINT_LOGOUT, middleware.AuthMiddleware, Logout)
	r.Post(constants.AUTH_ENDPOINT_REQUEST_LOGIN, RequestLogin)
	r.Post(constants.AUTH_ENDPOINT_EXCHANGE_CODE, ExchangeCode)
	// Public: AuthMiddleware → VerifyToken would reject an expired access token
	// before the handler runs, breaking the refresh-token probe branch. These
	// endpoints validate the tokens passed in the body themselves.
	r.Post(constants.AUTH_ENDPOINT_SSO_CHECK, SSOCheck)
	r.Post(constants.AUTH_ENDPOINT_SSO_CONSENT, SSOConsent)
}
