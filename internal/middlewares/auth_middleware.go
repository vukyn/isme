package middlewares

import (
	"strings"

	authModels "github.com/vukyn/isme/internal/domains/auth/models"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	"github.com/vukyn/kuery/log"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgHttp "github.com/vukyn/kuery/http/fiber"

	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) AuthMiddleware(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return pkgHttp.Unauthorized(c)
	}

	tokenParts := strings.Split(authorization, " ")
	if len(tokenParts) != 2 {
		return pkgHttp.Unauthorized(c)
	}

	tokenStr := tokenParts[1]

	verifyTokenResponse, err := m.authUC.VerifyToken(pkgCtx.NewContextFromFiberCtx(c), authModels.VerifyTokenRequest{
		Token: tokenStr,
	})
	if err != nil {
		log.New().Debugf("Invalid token: %v", err)
		return pkgHttp.Unauthorized(c)
	}
	if !verifyTokenResponse.Ok {
		return pkgHttp.Unauthorized(c)
	}

	// isme is its own app: extract the perms granted for the "isme" app and make
	// them the request-scoped perms used by rbac.RequirePermission on its routes
	claims := verifyTokenResponse.Claims
	pkgCtx.SetClaimsToFiberCtx(c, claims)
	pkgCtx.SetPermsToFiberCtx(c, claims.GetPermsForApp(roleConstants.APP_CODE_ISME))
	return c.Next()
}
