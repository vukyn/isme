package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"

	"github.com/gofiber/fiber/v2"
)

func SetupMediaRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)

	r := router.Group(constants.MEDIA_GROUP_NAME)

	// Avatar upload is a self-service action: any signed-in user may upload
	// their own profile photo. AuthMiddleware-gated, no RBAC permission — the
	// returned URL is only persisted on the caller's own user record.
	r.Post(constants.MEDIA_ENDPOINT_UPLOAD, middleware.AuthMiddleware, Upload)
}
