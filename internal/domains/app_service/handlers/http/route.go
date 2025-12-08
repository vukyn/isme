package handlers

import (
	iapp "isme/internal/app"
	idi "isme/internal/di"

	"github.com/gofiber/fiber/v2"
)

func SetupAppServiceRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)
	rAppService := router.Group("/app-service")
	rAppService.Post("/register", middleware.AuthMiddleware, RegisterApp)
	rAppService.Post("/verify", VerifyApp)
	rAppService.Post("/refresh", middleware.AuthMiddleware, RefreshApp)
}
