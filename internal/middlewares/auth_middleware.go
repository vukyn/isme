package middlewares

import (
	authModels "isme/internal/domains/auth/models"
	"strings"

	"github.com/vukyn/kuery/log"

	pkgCtx "isme/pkg/ctx"

	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) AuthMiddleware(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	tokenParts := strings.Split(authorization, " ")
	if len(tokenParts) != 2 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	tokenStr := tokenParts[1]

	verifyTokenResponse, err := m.authUC.VerifyToken(pkgCtx.NewContextFromFiberCtx(c), authModels.VerifyTokenRequest{
		Token: tokenStr,
	})
	if err != nil {
		log.New().Debugf("Invalid token: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	if !verifyTokenResponse.Ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	pkgCtx.SetClaimsToFiberCtx(c, verifyTokenResponse.Claims)
	return c.Next()
}
