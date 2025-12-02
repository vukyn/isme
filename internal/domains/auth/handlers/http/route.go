package handlers

import (
	iapp "isme/internal/app"
	idi "isme/internal/di"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)
	rAuth := router.Group("/auth")
	rAuth.Post("/signup", SignUp)
	rAuth.Post("/login", Login)
	rAuth.Post("/refresh", RefreshToken)
	rAuth.Get("/me", middleware.AuthMiddleware, GetMe)
	rAuth.Post("/change-password", middleware.AuthMiddleware, ChangePassword)
	rAuth.Post("/logout", middleware.AuthMiddleware, Logout)
}
