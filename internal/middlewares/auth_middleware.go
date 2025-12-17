package middlewares

import (
	"strings"

	authModels "github.com/vukyn/isme/internal/domains/auth/models"

	"github.com/vukyn/kuery/log"

	pkgCtx "github.com/vukyn/isme/pkg/ctx"
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

	pkgCtx.SetClaimsToFiberCtx(c, verifyTokenResponse.Claims)
	return c.Next()
}
