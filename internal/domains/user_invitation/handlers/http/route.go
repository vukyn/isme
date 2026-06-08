package handlers

import (
	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/constants"
	idi "github.com/vukyn/isme/internal/di"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	"github.com/vukyn/kuery/rbac"

	"github.com/gofiber/fiber/v2"
)

func SetupUserInvitationRoutes(router fiber.Router) {
	middleware := idi.GetMiddleware(iapp.App)

	// admin endpoints under /users — must register before the user routes so
	// /invites is matched ahead of /:userID params
	rUser := router.Group(constants.USER_GROUP_NAME, middleware.AuthMiddleware)
	rUser.Post(constants.USER_ENDPOINT_INVITES, rbac.RequirePermission(roleConstants.PERM_USER_CREATE), CreateInvitation)
	rUser.Get(constants.USER_ENDPOINT_INVITES, rbac.RequirePermission(roleConstants.PERM_USER_READ), ListInvitations)
	rUser.Post(constants.USER_ENDPOINT_INVITE_REVOKE, rbac.RequirePermission(roleConstants.PERM_USER_CREATE), RevokeInvitation)

	// public endpoints under /auth — invite resolution + accept
	rAuth := router.Group(constants.AUTH_GROUP_NAME)
	rAuth.Get(constants.AUTH_ENDPOINT_INVITE_DETAIL, GetInvitationByToken)
	rAuth.Post(constants.AUTH_ENDPOINT_ACCEPT_INVITE, AcceptInvitation)
}
