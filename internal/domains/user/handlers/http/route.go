package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	"github.com/vukyn/kuery/rbac"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)

	rUser := router.Group(constants.USER_GROUP_NAME, middleware.AuthMiddleware)
	rUser.Get(constants.USER_ENDPOINT_ROOT, rbac.RequirePermission(roleConstants.PERM_USER_READ), ListUsers)
	rUser.Patch(constants.USER_ENDPOINT_STATUS, rbac.RequirePermission(roleConstants.PERM_USER_UPDATE), UpdateUserStatus)
	rUser.Delete(constants.USER_ENDPOINT_DETAIL, rbac.RequirePermission(roleConstants.PERM_USER_DELETE), DeleteUser)
	rUser.Get(constants.USER_ENDPOINT_SESSIONS, rbac.RequirePermission(roleConstants.PERM_USER_SESSION_READ), ListUserSessions)
	rUser.Post(constants.USER_ENDPOINT_SESSION_REVOKE, rbac.RequirePermission(roleConstants.PERM_USER_SESSION_REVOKE), RevokeUserSession)
}
