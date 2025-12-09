package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"

	"github.com/gofiber/fiber/v2"
)

func SetupAppServiceRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)
	rAppService := router.Group(constants.APP_SERVICE_GROUP_NAME)
	rAppService.Post(constants.APP_SERVICE_ENDPOINT_REGISTER, middleware.AuthMiddleware, RegisterApp)
	rAppService.Post(constants.APP_SERVICE_ENDPOINT_VERIFY, VerifyApp)
	rAppService.Post(constants.APP_SERVICE_ENDPOINT_REFRESH, middleware.AuthMiddleware, RefreshApp)
}
