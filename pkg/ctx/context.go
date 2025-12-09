package ctx

import (
	"context"

	pkgClaims "github.com/vukyn/isme/pkg/claims"

	"github.com/gofiber/fiber/v2"
	"github.com/sarulabs/di/v2"
)

type ContextKey string

var (
	UserIDKey             ContextKey = "user_id"
	EmailKey              ContextKey = "email"
	TokenIDKey            ContextKey = "token_id"
	ClientIPKey           ContextKey = "client_ip"
	UserAgentKey          ContextKey = "user_agent"
	DiContainerRequestKey ContextKey = "di_container_request"
)

func SetClaimsToFiberCtx(ctx *fiber.Ctx, claims pkgClaims.Claims) {
	ctx.Locals(string(UserIDKey), claims.GetUserID())
	ctx.Locals(string(EmailKey), claims.GetEmail())
	ctx.Locals(string(TokenIDKey), claims.GetTokenID())
}

func NewContextFromFiberCtx(fiberCtx *fiber.Ctx) context.Context {
	userID := GetUserIdFromFiberCtx(fiberCtx)
	email := GetUserEmailFromFiberCtx(fiberCtx)
	tokenID := GetTokenIDFromFiberCtx(fiberCtx)
	userAgent := GetUserAgentFromFiberCtx(fiberCtx)
	clientIP := GetClientIPFromFiberCtx(fiberCtx)

	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, EmailKey, email)
	ctx = context.WithValue(ctx, TokenIDKey, tokenID)
	ctx = context.WithValue(ctx, UserAgentKey, userAgent)
	ctx = context.WithValue(ctx, ClientIPKey, clientIP)
	return ctx
}

func GetUserId(ctx context.Context) string {
	userID := ctx.Value(UserIDKey)
	if userID == nil {
		return ""
	}
	if userID, ok := userID.(string); ok {
		return userID
	}
	return ""
}

func GetUserEmail(ctx context.Context) string {
	email := ctx.Value(EmailKey)
	if email == nil {
		return ""
	}
	if email, ok := email.(string); ok {
		return email
	}
	return ""
}

func GetTokenID(ctx context.Context) string {
	tokenID := ctx.Value(TokenIDKey)
	if tokenID == nil {
		return ""
	}
	if tokenID, ok := tokenID.(string); ok {
		return tokenID
	}
	return ""
}

func GetClientIP(ctx context.Context) string {
	clientIP := ctx.Value(ClientIPKey)
	if clientIP == nil {
		return ""
	}
	return clientIP.(string)
}

func GetUserAgent(ctx context.Context) string {
	userAgent := ctx.Value(UserAgentKey)
	if userAgent == nil {
		return ""
	}
	return userAgent.(string)
}

func GetUserIdFromFiberCtx(ctx *fiber.Ctx) string {
	val := ctx.Locals(string(UserIDKey))
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetUserEmailFromFiberCtx(ctx *fiber.Ctx) string {
	val := ctx.Locals(string(EmailKey))
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetTokenIDFromFiberCtx(ctx *fiber.Ctx) string {
	val := ctx.Locals(string(TokenIDKey))
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetUserAgentFromFiberCtx(ctx *fiber.Ctx) string {
	return ctx.Get("User-Agent")
}

func GetClientIPFromFiberCtx(ctx *fiber.Ctx) string {
	return ctx.IP()
}

func SetDiContainerRequestToFiberCtx(ctx *fiber.Ctx, request di.Container) {
	ctx.Locals(string(DiContainerRequestKey), request)
}

func GetDiContainerRequestFromFiberCtx(ctx *fiber.Ctx) di.Container {
	container := ctx.Locals(string(DiContainerRequestKey))
	if container == nil {
		return di.Container{}
	}
	if container, ok := container.(di.Container); ok {
		return container
	}
	return di.Container{}
}
