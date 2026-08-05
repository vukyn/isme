package middlewares

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/vukyn/isme/internal/config"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	pkgCtx "github.com/vukyn/kuery/ctx"
	"github.com/vukyn/kuery/rbac"
	pkgRecover "github.com/vukyn/kuery/recover"

	"github.com/gofiber/fiber/v2"
	"github.com/sarulabs/di/v2"
)

// releaseProbeName is the DI definition the container-lifetime tests resolve on
// every request. Its only job is to own a Close function, because a Close counter is
// the ONLY way a test can see a DOUBLE release: sarulabs/di's second Delete on an
// already-closed container returns nil and changes no observable state — it just
// runs every registered Close a second time. isme's real Close funcs are all debug
// log lines, so a leftover `defer ctn.Delete()` in a handler is invisible to
// IsClosed() and visible here.
const releaseProbeName = "release-probe"

// releaseHarnessOptions configures the harness. skipResolve mounts the probe
// middleware WITHOUT building the probe object, which is the "request that resolves
// nothing at all" case — /assets, /favicon.svg and the SPA catch-all never touch the
// container.
type releaseHarnessOptions struct {
	skipResolve bool
}

// releaseHarness is the real middleware stack internal/server mounts — the DI
// container injection first, panic recovery INSIDE it — in front of one route per way
// a request can end. It records every request's sub-container plus how many times a
// Close ran, so both halves of "released exactly once" are assertable.
//
// The mutex is not decoration: app.Test serves each request on another goroutine, so
// the probe middleware and the Close callback both write from there while the test
// reads from here.
type releaseHarness struct {
	app          *fiber.App
	appContainer di.Container

	mutex      sync.Mutex
	containers []di.Container
	closeCalls int
}

func newReleaseHarness(t *testing.T, options releaseHarnessOptions) *releaseHarness {
	t.Helper()

	harness := &releaseHarness{containers: make([]di.Container, 0, 8)}

	builder, err := di.NewBuilder()
	if err != nil {
		t.Fatalf("di builder: %v", err)
	}
	if err := builder.Add(di.Def{
		Name:  releaseProbeName,
		Scope: di.Request,
		Build: func(ctn di.Container) (any, error) { return new(int), nil },
		Close: func(obj any) error {
			harness.mutex.Lock()
			defer harness.mutex.Unlock()
			harness.closeCalls++
			return nil
		},
	}); err != nil {
		t.Fatalf("add release probe definition: %v", err)
	}
	harness.appContainer = builder.Build()
	t.Cleanup(func() {
		// Only if an assertion did not already close it — a second delete would run
		// the probe Closes again and make the counter lie for any later reader.
		if !harness.appContainer.IsClosed() {
			_ = harness.appContainer.DeleteWithSubContainers()
		}
	})

	// The real Middleware, with a nil auth usecase: AuthMiddleware answers 401 on a
	// missing/malformed Authorization header BEFORE it ever reaches the usecase, and a
	// tokenless request is exactly the 401 under test here. A test that needed a
	// verified token would have to inject a real usecase.
	middleware := NewMiddleware(&config.Config{}, nil)

	app := fiber.New()
	// The order internal/server uses. DiContainerMiddleware is outermost, so the
	// recover middleware runs INSIDE it: a panic is caught below and the container's
	// defer runs on the way out either way.
	app.Use(DiContainerMiddleware(harness.appContainer))
	app.Use(pkgRecover.NewFiberRecover())
	app.Use(func(c *fiber.Ctx) error {
		container := pkgCtx.GetDiContainerRequestFromFiberCtx(c)
		harness.mutex.Lock()
		harness.containers = append(harness.containers, container)
		harness.mutex.Unlock()
		if !options.skipResolve {
			// Build the probe INSIDE this container so a Close is registered against
			// it — di only closes objects it actually built.
			if _, err := container.SafeGet(releaseProbeName); err != nil {
				return err
			}
		}
		return c.Next()
	})

	// One route per ending. Registration order is match order in Fiber, so the SPA
	// catch-all stays last.
	app.Get("/ok", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	app.Get("/error", func(c *fiber.Ctx) error { return fiber.NewError(http.StatusTeapot, "handler failed") })
	app.Get("/panic", func(c *fiber.Ctx) error { panic("handler exploded") })
	// The real middlewares, not stubs: a tokenless request is a 401 straight out of
	// AuthMiddleware, and rbac.RequirePermission 403s a caller with no perms in its
	// context — the gate every isme management route is mounted behind.
	app.Get("/protected", middleware.AuthMiddleware, func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	app.Get("/permission-gated", rbac.RequirePermission(roleConstants.PERM_USER_READ), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	app.Get("/*", func(c *fiber.Ctx) error { return c.SendString("spa") })

	harness.app = app
	return harness
}

// get performs one GET and returns its status.
func (h *releaseHarness) get(t *testing.T, path string) int {
	t.Helper()
	response, err := h.app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

// requireStatus performs one GET and fails unless the status matches, so a subtest
// proves it exercised the path it claims to (an unexpected 200 from /protected would
// otherwise silently turn a short-circuit test into a handler test).
func (h *releaseHarness) requireStatus(t *testing.T, path string, want int) {
	t.Helper()
	if got := h.get(t, path); got != want {
		t.Fatalf("GET %s = %d, want %d", path, got, want)
	}
}

// assertReleased is the whole property, in two halves. Every container the middleware
// created must be CLOSED (no leak), and Close must have run EXACTLY wantCloses times
// (no double release). wantCloses is normally the request count; it is 0 for the
// skipResolve harness, where nothing was ever built to close.
func (h *releaseHarness) assertReleased(t *testing.T, wantRequests, wantCloses int) {
	t.Helper()
	h.mutex.Lock()
	containers := slices.Clone(h.containers)
	closeCalls := h.closeCalls
	h.mutex.Unlock()

	if len(containers) != wantRequests {
		t.Fatalf("captured %d request containers, want %d", len(containers), wantRequests)
	}
	for index, container := range containers {
		if !container.IsClosed() {
			t.Fatalf("request %d's container is still OPEN: DiContainerMiddleware created it and must release it on every path, including the ones that never reach a handler",
				index)
		}
	}
	if closeCalls != wantCloses {
		t.Fatalf("Close ran %d times for %d requests, want %d — more than one per request means the container is being deleted twice (a leftover `defer ctn.Delete()` in a handler: di's second Delete returns nil but re-runs every Close)",
			closeCalls, wantRequests, wantCloses)
	}
}

// assertNoRetainedChildren proves the parent retains ZERO request sub-containers —
// the leak itself rather than a proxy for it, and without reflection.
//
// sarulabs/di does not export the children map, but Delete() reads it
// (containerSlayer.go:22-34): with any child present it does NOT close the container,
// it only sets deleteIfNoChild and returns nil. So "the app container closes on its
// first Delete" IS "its children map is empty". Destroys the app container, so this
// is the last thing a subtest does.
func (h *releaseHarness) assertNoRetainedChildren(t *testing.T) {
	t.Helper()
	if err := h.appContainer.Delete(); err != nil {
		t.Fatalf("delete app container: %v", err)
	}
	if !h.appContainer.IsClosed() {
		t.Fatal("the app container did NOT close on its first Delete, so it still retains request sub-containers — that is the leak: sarulabs/di holds every sub-container in its parent's children map until it is deleted, so each leaked one is retained for the life of the process")
	}
}

// TestDiContainerMiddlewareReleasesRequestContainer covers the ownership rule that
// replaced the handlers' `defer ctn.Delete()` (55 of them): the middleware that
// CREATES the request sub-container releases it, on every path.
//
// The leaking paths were the cheap unauthenticated ones an attacker can issue for
// free — a tokenless 401, a 403, /assets, /favicon.svg, a 404, every SPA render — so
// the leak needed no privileges to trigger and grew for the life of the process.
func TestDiContainerMiddlewareReleasesRequestContainer(t *testing.T) {
	t.Run("a request that reaches its handler", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		harness.requireStatus(t, "/ok", http.StatusOK)
		harness.assertReleased(t, 1, 1)
		harness.assertNoRetainedChildren(t)
	})

	// The two short-circuits that carry the real production volume: an expired
	// session and a caller reaching for a route its role does not grant.
	t.Run("a 401 from AuthMiddleware, before any handler", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		harness.requireStatus(t, "/protected", http.StatusUnauthorized)
		harness.assertReleased(t, 1, 1)
		harness.assertNoRetainedChildren(t)
	})

	t.Run("a 403 from rbac.RequirePermission, before any handler", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		harness.requireStatus(t, "/permission-gated", http.StatusForbidden)
		harness.assertReleased(t, 1, 1)
		harness.assertNoRetainedChildren(t)
	})

	t.Run("a handler that returns an error", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		harness.requireStatus(t, "/error", http.StatusTeapot)
		harness.assertReleased(t, 1, 1)
		harness.assertNoRetainedChildren(t)
	})

	// A defer runs while a panic unwinds, so this SHOULD hold — but "should" is the
	// reason to test it: the release lives in the middleware OUTSIDE the recover
	// middleware, and if it were ever moved inside (or turned into a plain call after
	// c.Next()) a panicking handler would leak silently, and a panic is exactly the
	// state nobody watches memory during.
	t.Run("a panic recovered by the recover middleware", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		harness.requireStatus(t, "/panic", http.StatusInternalServerError)
		harness.assertReleased(t, 1, 1)
		harness.assertNoRetainedChildren(t)
	})

	t.Run("the SPA catch-all, whose handler never touches the container", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		harness.requireStatus(t, "/apps/detail", http.StatusOK)
		harness.assertReleased(t, 1, 1)
		harness.assertNoRetainedChildren(t)
	})

	// The /assets and /favicon.svg shape: the middleware still builds a container
	// (it is global, mounted ahead of the routes) and nothing resolves anything from
	// it. Nothing to Close, but the container itself must still go.
	t.Run("a request that resolves nothing at all", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{skipResolve: true})
		harness.requireStatus(t, "/assets/index-abc123.js", http.StatusOK)
		harness.assertReleased(t, 1, 0)
		harness.assertNoRetainedChildren(t)
	})

	// The regression test for the leak as it was measured: a flood of the cheapest
	// possible request — unauthenticated, no DB work — used to retain one
	// sub-container each, permanently.
	t.Run("a flood of short-circuited requests retains nothing", func(t *testing.T) {
		const requests = 500
		harness := newReleaseHarness(t, releaseHarnessOptions{})
		for range requests {
			harness.requireStatus(t, "/protected", http.StatusUnauthorized)
		}
		harness.assertReleased(t, requests, requests)
		harness.assertNoRetainedChildren(t)
	})

	// Every ending mixed in one process, which is the only shape that would catch a
	// release that works per-path but is somehow order-dependent.
	t.Run("every path in one process", func(t *testing.T) {
		harness := newReleaseHarness(t, releaseHarnessOptions{})

		paths := []struct {
			path string
			want int
		}{
			{"/ok", http.StatusOK},
			{"/protected", http.StatusUnauthorized},
			{"/permission-gated", http.StatusForbidden},
			{"/error", http.StatusTeapot},
			{"/panic", http.StatusInternalServerError},
			{"/spa/route", http.StatusOK},
		}
		for _, item := range paths {
			harness.requireStatus(t, item.path, item.want)
		}

		harness.assertReleased(t, len(paths), len(paths))
		harness.assertNoRetainedChildren(t)
	})
}
