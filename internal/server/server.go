package server

import (
	"fmt"
	iapp "isme/internal/app"
	"isme/internal/config"
	appServiceHandlers "isme/internal/domains/app_service/handlers/http"
	authHandlers "isme/internal/domains/auth/handlers/http"
	"net/http"
	"os"

	pkgCtx "isme/pkg/ctx"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/vukyn/kuery/log"

	pkgRecover "isme/pkg/recover"

	"github.com/gofiber/fiber/v2"
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

	engine := html.New("internal/ui", ".html")
	s.app = fiber.New(fiber.Config{
		AppName: s.cfg.App.Name,
		Views:   engine,
	})

	// Middlewares
	s.app.Use(cors.New())
	zerologLogger := log.New().Zerolog()
	s.app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &zerologLogger,
	}))

	// inject di container to fiber ctx
	s.app.Use(diContainerMiddleware)

	// recover from panic
	s.app.Use(pkgRecover.NewFiberRecover())

	// Static files - serve from root paths to match HTML references
	s.app.Use("/assets", filesystem.New(filesystem.Config{
		Root: http.Dir("internal/ui/assets"),
	}))

	// api/v1
	apiV1 := s.app.Group("/api/v1")
	authHandlers.SetupAuthRoutes(apiV1)
	appServiceHandlers.SetupAppServiceRoutes(apiV1)

	// web routes
	s.webRoutes(s.app)

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

func diContainerMiddleware(c *fiber.Ctx) error {
	request, err := iapp.App.SubContainer()
	if err != nil {
		return err
	}
	pkgCtx.SetDiContainerRequestToFiberCtx(c, request)
	return c.Next()
}

func (s *Server) webRoutes(app *fiber.App) {
	renderHomePage := func(c *fiber.Ctx) error {
		apiBaseURL := s.cfg.Vite.BaseURL
		if apiBaseURL == "" {
			apiBaseURL = s.cfg.Vite.BaseURL
		}
		return c.Render("index", fiber.Map{
			"APIBaseURL": apiBaseURL,
		})
	}

	app.Get("/*", renderHomePage)
}
