package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	"github.com/vukyn/kuery/rbac"

	"github.com/gofiber/fiber/v2"
)

func SetupSettingsRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)

	rSettings := router.Group(constants.SETTINGS_GROUP_NAME, middleware.AuthMiddleware)
	rSettings.Get(constants.SETTINGS_ENDPOINT_SESSION_REVOKE, rbac.RequirePermission(roleConstants.PERM_SETTINGS_READ), GetSessionRevokeConfig)
	rSettings.Put(constants.SETTINGS_ENDPOINT_SESSION_REVOKE, rbac.RequirePermission(roleConstants.PERM_SETTINGS_UPDATE), UpdateSessionRevokeConfig)
}
