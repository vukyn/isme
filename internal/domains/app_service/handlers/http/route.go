package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	"github.com/vukyn/kuery/rbac"

	"github.com/gofiber/fiber/v2"
)

func SetupAppServiceRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)
	rAppService := router.Group(constants.APP_SERVICE_GROUP_NAME)
	rAppService.Post(constants.APP_SERVICE_ENDPOINT_REGISTER, middleware.AuthMiddleware, RegisterApp)
	rAppService.Post(constants.APP_SERVICE_ENDPOINT_VERIFY, VerifyApp)
	rAppService.Post(constants.APP_SERVICE_ENDPOINT_REFRESH, middleware.AuthMiddleware, RefreshApp)
	rAppService.Get(constants.APP_SERVICE_ENDPOINT_ROOT, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_APP_SERVICE_READ), ListApps)
	rAppService.Get(constants.APP_SERVICE_ENDPOINT_DETAIL, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_APP_SERVICE_READ), GetApp)
	rAppService.Patch(constants.APP_SERVICE_ENDPOINT_DETAIL, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_APP_SERVICE_UPDATE), UpdateAppAppearance)
	rAppService.Patch(constants.APP_SERVICE_ENDPOINT_STATUS, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_APP_SERVICE_UPDATE), UpdateAppStatus)
}
