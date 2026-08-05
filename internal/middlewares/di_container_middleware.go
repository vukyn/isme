package middlewares

import (
	pkgCtx "github.com/vukyn/kuery/ctx"
	"github.com/vukyn/kuery/log"

	"github.com/gofiber/fiber/v2"
	"github.com/sarulabs/di/v2"
)

// DiContainerMiddleware creates a request-scoped sub-container off the given app
// container, stores it in the Fiber locals so handlers can resolve request-scoped
// dependencies, and — the load-bearing part — RELEASES it when the request is done.
//
// ⚠️ THE MIDDLEWARE THAT CREATES THE CONTAINER OWNS ITS LIFETIME. Handlers must NOT
// call `ctn.Delete()`; they only read the container out of the locals. This used to
// be the other way round (every handler carried a `defer ctn.Delete()` — 55 of them)
// and it leaked, unavoidably: this middleware is mounted globally, ahead of the
// routes, so it has already built a sub-container by the time anything decides the
// request is not going to reach a handler. `sarulabs/di` keeps every sub-container in
// its parent's `children` map until it is deleted, so each of those requests retained
// its container for the life of the PROCESS — and the leaking paths are the cheap,
// unauthenticated ones an attacker picks: a 401 from AuthMiddleware, a 403 from
// rbac.RequirePermission, /assets, /favicon.svg, a 404, and every SPA catch-all
// render. A handler-owned lifetime cannot cover a request that never reaches a
// handler; only the creator can.
//
// The release is a `defer`, so it also runs while a panic unwinds (the recover
// middleware is mounted INSIDE this one) and on any error return.
//
// DeleteWithSubContainers, not Delete: Delete is conditional — with any child
// present it merely sets `deleteIfNoChild` and returns nil, leaving the container in
// the parent's children map, which is precisely the leak. The single owner of a
// lifetime needs an unconditional release. Safe because nothing here uses the
// container after its handler returns: isme has no streaming responses (no
// SendStream/SetBodyStreamWriter/StreamRequestBody) and no goroutine that resolves
// from a request container — the scheduler builds its repositories straight off the
// App-scoped DB for exactly that reason.
func DiContainerMiddleware(app di.Container) fiber.Handler {
	return func(c *fiber.Ctx) error {
		request, err := app.SubContainer()
		if err != nil {
			return err
		}
		defer func() {
			if err := request.DeleteWithSubContainers(); err != nil {
				log.New().Errorf("release request di container: %v", err)
			}
		}()
		pkgCtx.SetDiContainerRequestToFiberCtx(c, request)
		return c.Next()
	}
}
