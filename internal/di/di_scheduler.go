package di

import (
	"github.com/vukyn/isme/internal/constants"
	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	"github.com/sarulabs/di/v2"
	"github.com/uptrace/bun"
	pkgScheduler "github.com/vukyn/kuery/scheduler"

	"github.com/vukyn/kuery/log"
)

// defineScheduler builds the app-scoped scheduler engine singleton. It is
// constructed once during the DI build from the App-scoped DB: it registers the
// four isme jobs (session-revoke, rotation-cleanup, activity-cleanup,
// database-backup) with their job bodies as closures over freshly built
// repositories. No WithLocation option
// is passed, so the engine evaluates schedules in the process's local time —
// matching the pre-migration engine exactly (parity).
func defineScheduler() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_SCHEDULER,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			cfg := GetConfig(ctn)

			// build the repositories the job bodies close over, directly from the
			// App-scoped DB so the engine does not depend on request-scoped containers.
			userSessionRepository := userSessionRepo.NewRepository(db)
			settingsRepository := settingsRepo.NewRepository(db)
			activityRepository := activityRepo.NewRepository(db)

			// NO WithLocation — gocron defaults to local time, matching the original engine.
			engine, err := pkgScheduler.New(cfg.Scheduler.Enabled)
			if err != nil {
				return nil, err
			}

			engine.Register(pkgScheduler.Job{
				Key: pkgScheduler.JobKey(settingsEntity.JobKeySessionRevoke),
				Run: newSessionRevokeRun(userSessionRepository, settingsRepository),
			})
			engine.Register(pkgScheduler.Job{
				Key: pkgScheduler.JobKey(settingsEntity.JobKeyRotationCleanup),
				Run: newRotationCleanupRun(userSessionRepository, settingsRepository),
			})
			engine.Register(pkgScheduler.Job{
				Key: pkgScheduler.JobKey(settingsEntity.JobKeyActivityCleanup),
				Run: newActivityCleanupRun(activityRepository, settingsRepository),
			})
			engine.Register(pkgScheduler.Job{
				Key: pkgScheduler.JobKey(settingsEntity.JobKeyDatabaseBackup),
				Run: newDatabaseBackupRun(db, settingsRepository),
			})

			log.New().Debug("Scheduler initialized")
			return engine, nil
		},
		Close: func(obj any) error {
			engine := obj.(*pkgScheduler.Engine)
			engine.Stop()
			log.New().Debug("Scheduler destroyed")
			return nil
		},
	}
	return def
}

// defineScheduleProvider builds the app-scoped schedule provider singleton. It
// is the engine's initial-load path at Start: it reads each registered job's
// persisted schedule from the settings repository (built from the App-scoped DB).
func defineScheduleProvider() *di.Def {
	def := &di.Def{
		Name:  constants.CONTAINER_NAME_SCHEDULE_PROVIDER,
		Scope: di.App,
		Build: func(ctn di.Container) (any, error) {
			db := ctn.Get(constants.CONTAINER_NAME_DB).(*bun.DB)
			settingsRepository := settingsRepo.NewRepository(db)
			log.New().Debug("Schedule provider initialized")
			return newScheduleProvider(settingsRepository), nil
		},
		Close: func(obj any) error {
			log.New().Debug("Schedule provider destroyed")
			return nil
		},
	}
	return def
}

func GetScheduler(ctn di.Container) *pkgScheduler.Engine {
	return ctn.Get(constants.CONTAINER_NAME_SCHEDULER).(*pkgScheduler.Engine)
}

func GetScheduleProvider(ctn di.Container) pkgScheduler.ScheduleProvider {
	return ctn.Get(constants.CONTAINER_NAME_SCHEDULE_PROVIDER).(pkgScheduler.ScheduleProvider)
}
