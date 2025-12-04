package main

import (
	"context"
	"isme/internal/app"
	"isme/internal/server"
	"isme/pkg/graceful"
	"time"

	"github.com/vukyn/kuery/log"
)

func init() {
	app.Init()
}

func main() {
	// start server
	server := server.NewServer(app.Config)
	server.Start()

	// graceful shutdown
	shutdown := func(ctx context.Context) error {
		log.New().Info("Shutting down server")
		err := server.Stop()
		if err != nil {
			log.New().Errorf("Failed to stop server: %v", err)
			return err
		}
		log.New().Debug("Server stopped")
		err = app.App.DeleteWithSubContainers()
		if err != nil {
			log.New().Errorf("Failed to delete app: %v", err)
			return err
		}
		log.New().Debug("App deleted")
		return nil
	}

	graceful.ShutdownWithCallback(
		shutdown,
		&graceful.ShutdownOptions{
			Signals:   graceful.DefaultSignals,
			Verbose:   app.Config.Graceful.Verbose,
			StepDelay: time.Duration(app.Config.Graceful.StepDelay) * time.Millisecond,
			Timeout:   time.Duration(app.Config.Graceful.ServerShutdownTimeout) * time.Millisecond,
			Logger:    log.New(),
		},
	) /// ---> block here
}
