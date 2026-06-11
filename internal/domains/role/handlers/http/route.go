package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	"github.com/vukyn/kuery/rbac"

	"github.com/gofiber/fiber/v2"
)

func SetupRoleRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)

	rRole := router.Group(constants.ROLE_GROUP_NAME, middleware.AuthMiddleware)
	rRole.Get(constants.ROLE_ENDPOINT_ROOT, rbac.RequirePermission(roleConstants.PERM_ROLE_READ), ListRoles)
	rRole.Post(constants.ROLE_ENDPOINT_ROOT, rbac.RequirePermission(roleConstants.PERM_ROLE_CREATE), CreateRole)
	rRole.Get(constants.ROLE_ENDPOINT_DETAIL, rbac.RequirePermission(roleConstants.PERM_ROLE_READ), GetRoleDetail)
	rRole.Put(constants.ROLE_ENDPOINT_DETAIL, rbac.RequirePermission(roleConstants.PERM_ROLE_UPDATE), UpdateRole)
	rRole.Delete(constants.ROLE_ENDPOINT_DETAIL, rbac.RequirePermission(roleConstants.PERM_ROLE_DELETE), DeleteRole)
	rRole.Put(constants.ROLE_ENDPOINT_PERMISSIONS, rbac.RequirePermission(roleConstants.PERM_ROLE_UPDATE), SetRolePermissions)
	rRole.Get(constants.ROLE_ENDPOINT_MEMBERS, rbac.RequirePermission(roleConstants.PERM_ROLE_READ), ListRoleMembers)
	rRole.Post(constants.ROLE_ENDPOINT_MEMBERS, rbac.RequirePermission(roleConstants.PERM_ROLE_ASSIGN), AddRoleMembers)
	rRole.Delete(constants.ROLE_ENDPOINT_MEMBER_DETAIL, rbac.RequirePermission(roleConstants.PERM_ROLE_ASSIGN), RemoveRoleMember)

	router.Get(constants.PERMISSION_ENDPOINT_CATALOG, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_ROLE_READ), ListPermissions)
	router.Post(constants.PERMISSION_ENDPOINT_CATALOG, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_ROLE_CREATE), CreatePermissions)
	router.Put(constants.PERMISSION_ENDPOINT_APPEARANCE, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_ROLE_UPDATE), UpdatePermissionAppearance)
	router.Delete(constants.PERMISSION_ENDPOINT_DETAIL, middleware.AuthMiddleware, rbac.RequirePermission(roleConstants.PERM_ROLE_DELETE), DeletePermission)
}
