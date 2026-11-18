package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	iapp "github.com/vukyn/isme/internal/app"
	"github.com/vukyn/isme/internal/config"
	appServiceHandlers "github.com/vukyn/isme/internal/domains/app_service/handlers/http"
	authHandlers "github.com/vukyn/isme/internal/domains/auth/handlers/http"
	mediaHandlers "github.com/vukyn/isme/internal/domains/media/handlers/http"
	roleHandlers "github.com/vukyn/isme/internal/domains/role/handlers/http"
	settingsHandlers "github.com/vukyn/isme/internal/domains/settings/handlers/http"
	userHandlers "github.com/vukyn/isme/internal/domains/user/handlers/http"
	userInvitationHandlers "github.com/vukyn/isme/internal/domains/user_invitation/handlers/http"
	"github.com/vukyn/isme/internal/middlewares"
	"github.com/vukyn/isme/internal/web"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v3"
	pkgFiber "github.com/vukyn/kuery/http/fiberv3"
	pkgHealth "github.com/vukyn/kuery/http/health"
	"github.com/vukyn/kuery/log"
	pkgRecover "github.com/vukyn/kuery/recoverv3"
)

type Server struct {
	app *fiber.App
	cfg *config.Config
}

func NewServer(
	cfg *config.Config,
) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Start() {
	log.New().Info("Starting server")

	// The built UI is embedded in the binary (internal/web) — served from that FS
	// rather than a working-directory-relative path, so the binary is self-contained.
	uiFS := web.FS()
	engine := html.NewFileSystem(http.FS(uiFS), ".html")
	s.app = fiber.New(fiber.Config{
		AppName: s.cfg.App.Name,
		Views:   engine,
	})

	// Middlewares
	//
	// CORS stays allow-all on purpose: it is what the service ran with on Fiber v2
	// (a bare cors.New()), and this migration changes no policy. Tightening it to
	// an explicit origin list is an open question, not an oversight. Note that on
	// v3 an EMPTY AllowOrigins slice is also allow-all, so a future allow-list
	// must never be allowed to resolve to zero entries.
	s.app.Use(cors.New())

	// Request logging. Fiber v3 has no fiberzerolog equivalent (gofiber/contrib
	// never published a v3 module of it — its whole API is *fiber.Ctx), so the
	// built-in logger middleware is pointed at kuery's zerolog through a
	// LoggerFunc. The logger runs the app's error handler itself before calling
	// LoggerFunc, so the status read off the response is the final one.
	s.app.Use(fiberLogger.New(fiberLogger.Config{
		LoggerFunc: logRequest,
	}))

	// Inject the request-scoped di container into the fiber ctx — and release it
	// when the request ends. The middleware owns the whole lifetime (handlers only
	// borrow it), so the release covers every path, including the ones that never
	// reach a handler. Mounted ahead of the recover middleware below so its defer
	// also runs while a panic unwinds.
	s.app.Use(middlewares.DiContainerMiddleware(iapp.App))

	// recover from panic
	s.app.Use(pkgRecover.NewFiberRecover())

	// api/v1
	apiV1 := s.app.Group("/api/v1")

	// Deployment verification. Unauthenticated and deliberately I/O-free: it
	// reports the module versions the running binary was BUILT with, read from
	// the embedded build info, so "which build is live?" is one curl rather
	// than a guess from an incidental clue in the logs.
	//
	// It touches no database on purpose. This app is scale-to-zero over a
	// suspending Postgres, so a health check that pinged the DB would wake it
	// on every call — and "the process is up" is a different question from
	// "its database is reachable".
	apiV1.Get("/__version", pkgFiber.Health(pkgHealth.New(
		"github.com/vukyn/kuery",
		"github.com/gofiber/fiber/v3",
	)))
	authHandlers.SetupAuthRoutes(apiV1)
	appServiceHandlers.SetupAppServiceRoutes(apiV1)
	// before user routes so /users/invites is matched ahead of /users/:userID
	userInvitationHandlers.SetupUserInvitationRoutes(apiV1)
	userHandlers.SetupUserRoutes(apiV1)
	roleHandlers.SetupRoleRoutes(apiV1)
	settingsHandlers.SetupSettingsRoutes(apiV1)
	mediaHandlers.SetupMediaRoutes(apiV1)

	// web routes: /assets, root files and the SPA catch-all
	if err := mountWeb(s.app, uiFS, s.renderHomePage); err != nil {
		log.New().Errorf("Failed to open embedded assets: %v", err)
		os.Exit(1)
	}

	// start server
	go func() {
		err := s.app.Listen(fmt.Sprintf(":%d", iapp.Config.App.Port))
		if err != nil {
			log.New().Errorf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()
}

func (s *Server) Stop() error {
	return s.app.Shutdown()
}

// logRequest writes one structured line per request to kuery's zerolog: Error
// for a 5xx, Warn for a 4xx, Info otherwise.
func logRequest(c fiber.Ctx, data *fiberLogger.Data, _ *fiberLogger.Config) error {
	logger := log.New().Zerolog()
	// logger.Data carries no status — read it off the response.
	status := c.Response().StatusCode()
	event := logger.Info()
	switch {
	case status >= fiber.StatusInternalServerError:
		event = logger.Error()
	case status >= fiber.StatusBadRequest:
		event = logger.Warn()
	}
	if data.ChainErr != nil {
		event = event.Err(data.ChainErr)
	}
	event.
		Str("method", c.Method()).
		Str("path", c.Path()).
		Int("status", status).
		Dur("latency", data.Stop.Sub(data.Start)).
		Str("ip", c.IP()).
		Msg("request")
	return nil
}

func (s *Server) renderHomePage(c fiber.Ctx) error {
	apiBaseURL := s.cfg.Vite.BaseURL
	if apiBaseURL == "" {
		apiBaseURL = s.cfg.Vite.BaseURL
	}
	return c.Render("index", fiber.Map{
		"APIBaseURL": apiBaseURL,
	})
}

// mountWeb registers the embedded UI: the hashed bundle under /assets, the
// root-level files index.html references directly, and the SPA catch-all. It
// must be called after every API route, since the catch-all matches anything.
// Split out of Start so tests can mount exactly these routes over a fake FS.
func mountWeb(app *fiber.App, uiFS fs.FS, renderHomePage fiber.Handler) error {
	// Static files - serve from root paths to match HTML references. Assets live
	// under the embedded FS's assets/ subtree. Fiber v3 replaced
	// middleware/filesystem with middleware/static, whose Config.FS takes an
	// fs.FS directly.
	assetsFS, err := fs.Sub(uiFS, "assets")
	if err != nil {
		return err
	}
	app.Use("/assets", func(c fiber.Ctx) error {
		// The embed FS has a zero modtime. Older fasthttp releases stamped it as
		// `Last-Modified: Mon, 01 Jan 0001` and answered ANY If-Modified-Since
		// with 304 (the one in go.mod today skips a zero modtime, but nothing pins
		// that). A date validator against a zero modtime means nothing, so it is
		// dropped on the way in and stripped on the way out; a stale validator
		// can never pin an old file.
		c.Request().Header.Del(fiber.HeaderIfModifiedSince)
		return c.Next()
	}, static.New("", static.Config{
		FS: assetsFS,
		ModifyResponse: func(c fiber.Ctx) error {
			c.Response().Header.Del(fiber.HeaderLastModified)
			return nil
		},
		// v3's static middleware falls through to the next route on a miss,
		// which here is the SPA catch-all: a missing chunk would come back as
		// index.html at 200. A missing asset is a plain 404 instead.
		NotFoundHandler: func(c fiber.Ctx) error {
			return c.SendStatus(fiber.StatusNotFound)
		},
	}))

	// Root-level static assets that index.html references directly (outside
	// /assets). Without this, the SPA catch-all below renders index.html for
	// /favicon.svg, so the browser gets HTML instead of the icon and falls back
	// to its generic globe. Read from the FS rather than sent with SendFile,
	// which can stamp the embed's zero modtime as Last-Modified (see above).
	app.Get("/favicon.svg", func(c fiber.Ctx) error {
		favicon, err := fs.ReadFile(uiFS, "favicon.svg")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Type("svg")
		return c.Send(favicon)
	})

	app.Get("/*", renderHomePage)
	return nil
}
