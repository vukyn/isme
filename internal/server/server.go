package server

import (
	iapp "isme/internal/app"
	authHandlers "isme/internal/domains/auth/handlers/http"
	"os"

	pkgCtx "isme/pkg/ctx"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/vukyn/kuery/log"

	pkgRecover "isme/pkg/recover"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() {
	log.New().Info("Starting server")

	s.app = fiber.New()

	// Add CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",                                      // Allow all origins
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS", // Allow all common methods
		AllowHeaders:     "*",                                      // Allow all headers
		AllowCredentials: false,                                    // Must be false when AllowOrigins is "*"
	}))

	// inject di container to fiber ctx
	s.app.Use(diContainerMiddleware)

	// recover from panic
	s.app.Use(pkgRecover.NewFiberRecover())

	// api/v1
	apiV1 := s.app.Group("/api/v1")
	authHandlers.SetupAuthRoutes(apiV1)

	// start server
	go func() {
		err := s.app.Listen(":8080")
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
