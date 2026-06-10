package di

import (
	"github.com/vukyn/isme/internal/constants"
	"github.com/vukyn/isme/internal/scheduler"

	"github.com/sarulabs/di/v2"
	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/log"
)

// schedulerInstance holds the app-scoped scheduler singleton. It is constructed
// once during the DI build (from the App-scoped DB) and returned by the
// container so the settings usecase can resolve it as an IReloader.
func defineScheduler() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_SCHEDULER,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			cfg := GetConfig(ctn)
			instance, err := scheduler.New(db, cfg.Scheduler.Enabled)
			if err != nil {
				return nil, err
			}
			log.New().Debug("Scheduler initialized")
			return instance, nil
		},
		Close: func(obj any) error {
			instance := obj.(*scheduler.Scheduler)
			instance.Stop()
			log.New().Debug("Scheduler destroyed")
			return nil
		},
	}
	return def
}

func GetScheduler(ctn di.Container) *scheduler.Scheduler {
	return ctn.Get(constants.CONTAINER_NAME_SCHEDULER).(*scheduler.Scheduler)
}
