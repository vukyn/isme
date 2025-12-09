package handlers

import (
	iapp "isme/internal/app"
	"isme/internal/constants"
	idi "isme/internal/di"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)
	rAuth := router.Group(constants.AUTH_GROUP_NAME)
	rAuth.Post(constants.AUTH_ENDPOINT_SIGNUP, SignUp)
	rAuth.Post(constants.AUTH_ENDPOINT_LOGIN, Login)
	rAuth.Post(constants.AUTH_ENDPOINT_REFRESH, RefreshToken)
	rAuth.Get(constants.AUTH_ENDPOINT_ME, middleware.AuthMiddleware, GetMe)
	rAuth.Post(constants.AUTH_ENDPOINT_CHANGE_PASSWORD, middleware.AuthMiddleware, ChangePassword)
	rAuth.Post(constants.AUTH_ENDPOINT_LOGOUT, middleware.AuthMiddleware, Logout)
	rAuth.Post(constants.AUTH_ENDPOINT_REQUEST_LOGIN, RequestLogin)
	rAuth.Post(constants.AUTH_ENDPOINT_EXCHANGE_CODE, ExchangeCode)
}
