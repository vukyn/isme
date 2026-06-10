package app

import (
	"context"

	"github.com/vukyn/kuery/log"

	"github.com/vukyn/isme/internal/config"
	idi "github.com/vukyn/isme/internal/di"

	"github.com/sarulabs/di/v2"
)

var App di.Container
var Config *config.Config
var Scheduler interface {
	Start(ctx context.Context)
	Stop()
}

func Init() {
	app, err := idi.NewBuilder().Build() // build all dependencies
	if err != nil {
		log.New().Fatal("Failed to build app", err)
	}
	App = app
	Config = idi.GetConfig(app)
	err = log.Init(log.Config{
		Mode:  Config.Logger.Mode,
		Level: Config.Logger.Level,
	})
	if err != nil {
		log.New().Fatal("Failed to initialize logger", err)
	}
	log.New().Info("Logger initialized")

	// Force database initialization by accessing it
	_ = idi.GetDB(app)

	// Force scheduler singleton construction (built from the App-scoped DB)
	Scheduler = idi.GetScheduler(app)
}
