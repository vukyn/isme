package scheduler

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/log"
)

// JobKey identifies a named scheduled job in the keyed registry. The string
// values are sourced from the settings entity package — the single source of
// truth shared with the repository, usecase and migration.
type JobKey string

const (
	// JobSessionRevoke prunes expired-but-active sessions.
	JobSessionRevoke JobKey = settingsEntity.JobKeySessionRevoke
	// JobRotationCleanup prunes old token_rotation_events rows.
	JobRotationCleanup JobKey = settingsEntity.JobKeyRotationCleanup
	// JobActivityCleanup prunes old activity_events rows.
	JobActivityCleanup JobKey = settingsEntity.JobKeyActivityCleanup
)

// rotationCleanupParams mirrors the params JSON of the rotation-cleanup job.
type rotationCleanupParams struct {
	RetentionHours int64 `json:"retention_hours"`
}

// activityCleanupParams mirrors the params JSON of the activity-cleanup job.
// Retention is in DAYS (not hours like rotation), so the cutoff multiplies by
// 24h per day.
type activityCleanupParams struct {
	RetentionDays int64 `json:"retention_days"`
}

// IReloader is the seam the settings usecase depends on to re-apply a changed
// schedule. It lives in the scheduler package so the usecase imports the
// scheduler (one direction only) and no import cycle forms: the scheduler
// imports the settings + user_session REPOSITORIES, never their usecases.
type IReloader interface {
	// Reload re-applies the schedule for the named job: removes any current job
	// and, when enabled, installs a new cron job for the given expression.
	Reload(ctx context.Context, jobKey JobKey, enabled bool, cronExpr string) error
}

// Scheduler owns a gocron scheduler and a keyed registry of named jobs. It is an
// application-scoped singleton (one per process), constructed directly from the
// *bun.DB so it does not depend on request-scoped DI containers.
type Scheduler struct {
	sched           gocron.Scheduler
	userSessionRepo userSessionRepo.IRepository
	settingsRepo    settingsRepo.IRepository
	activityRepo    activityRepo.IRepository
	enabled         bool

	mu      sync.Mutex
	jobs    map[JobKey]uuid.UUID
	started bool
	stopped bool
}

// New constructs the scheduler singleton. enabled is the master kill-switch
// (env SCHEDULER_ENABLED): when false the scheduler never installs any job.
func New(db *bun.DB, enabled bool) (*Scheduler, error) {
	sched, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		sched:           sched,
		userSessionRepo: userSessionRepo.NewRepository(db),
		settingsRepo:    settingsRepo.NewRepository(db),
		activityRepo:    activityRepo.NewRepository(db),
		enabled:         enabled,
		jobs:            make(map[JobKey]uuid.UUID),
	}, nil
}

// Start reads the persisted config and, if both the env kill-switch and the DB
// flag allow it, installs each job. The gocron scheduler is always started so a
// later Reload can add a job without a restart. Each job loads and installs
// independently — one job's failure logs and continues, never blocking the other.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sched.Start()
	s.started = true

	if !s.enabled {
		log.New().Info("Scheduler disabled by SCHEDULER_ENABLED")
		return
	}

	s.loadSessionRevoke(ctx)
	s.loadRotationCleanup(ctx)
	s.loadActivityCleanup(ctx)
}

// loadSessionRevoke loads + installs the session-revoke job. Callers must hold s.mu.
func (s *Scheduler) loadSessionRevoke(ctx context.Context) {
	config, err := s.settingsRepo.GetSchedule(ctx, string(JobSessionRevoke))
	if err != nil {
		log.New().Errorf("Scheduler: failed to load session-revoke config: %v", err)
		return
	}
	if !config.Enabled || config.Cron == "" {
		return
	}
	if err := s.applyJob(JobSessionRevoke, config.Cron); err != nil {
		log.New().Errorf("Scheduler: failed to install session-revoke job: %v", err)
		return
	}
	log.New().Infof("Session auto-revoke scheduler started with cron %q", config.Cron)
}

// loadRotationCleanup loads + installs the rotation-cleanup job. Callers must hold s.mu.
func (s *Scheduler) loadRotationCleanup(ctx context.Context) {
	config, err := s.settingsRepo.GetSchedule(ctx, string(JobRotationCleanup))
	if err != nil {
		log.New().Errorf("Scheduler: failed to load rotation-cleanup config: %v", err)
		return
	}
	if !config.Enabled || config.Cron == "" {
		return
	}
	if err := s.applyJob(JobRotationCleanup, config.Cron); err != nil {
		log.New().Errorf("Scheduler: failed to install rotation-cleanup job: %v", err)
		return
	}
	log.New().Infof("Rotation cleanup scheduler started with cron %q", config.Cron)
}

// loadActivityCleanup loads + installs the activity-cleanup job. Callers must hold s.mu.
func (s *Scheduler) loadActivityCleanup(ctx context.Context) {
	config, err := s.settingsRepo.GetSchedule(ctx, string(JobActivityCleanup))
	if err != nil {
		log.New().Errorf("Scheduler: failed to load activity-cleanup config: %v", err)
		return
	}
	if !config.Enabled || config.Cron == "" {
		return
	}
	if err := s.applyJob(JobActivityCleanup, config.Cron); err != nil {
		log.New().Errorf("Scheduler: failed to install activity-cleanup job: %v", err)
		return
	}
	log.New().Infof("Activity cleanup scheduler started with cron %q", config.Cron)
}

// Reload re-applies the schedule for the named job live. It removes the current
// job (if any) and, when enabled (and the env kill-switch allows), installs a
// fresh cron job. Retention changes for the cleanup job need no reload — the
// retention window is read fresh on each run.
func (s *Scheduler) Reload(ctx context.Context, jobKey JobKey, enabled bool, cronExpr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil
	}

	s.removeJob(jobKey)

	if !s.enabled || !enabled || cronExpr == "" {
		log.New().Infof("Scheduler reloaded %q: no active job", jobKey)
		return nil
	}
	if err := s.applyJob(jobKey, cronExpr); err != nil {
		return err
	}
	log.New().Infof("Scheduler reloaded %q with cron %q", jobKey, cronExpr)
	return nil
}

// Stop shuts the scheduler down. It is idempotent and bounded by gocron's own
// shutdown — it must never block graceful shutdown indefinitely.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped || !s.started {
		s.stopped = true
		return
	}
	if err := s.sched.Shutdown(); err != nil {
		log.New().Errorf("Scheduler: shutdown error: %v", err)
	}
	s.stopped = true
}

// applyJob installs a new cron job for the given key, dispatching the task by
// key. Callers must hold s.mu.
func (s *Scheduler) applyJob(jobKey JobKey, cronExpr string) error {
	var task gocron.Task
	switch jobKey {
	case JobSessionRevoke:
		task = gocron.NewTask(func() { s.runRevoke(context.Background()) })
	case JobRotationCleanup:
		task = gocron.NewTask(func() { s.runCleanup(context.Background()) })
	case JobActivityCleanup:
		task = gocron.NewTask(func() { s.runActivityCleanup(context.Background()) })
	default:
		return nil
	}

	job, err := s.sched.NewJob(
		gocron.CronJob(cronExpr, false), // 5-field standard cron
		task,
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return err
	}
	s.jobs[jobKey] = job.ID()
	return nil
}

// removeJob removes the job for the given key if present. Callers must hold s.mu.
func (s *Scheduler) removeJob(jobKey JobKey) {
	jobID, ok := s.jobs[jobKey]
	if !ok {
		return
	}
	if err := s.sched.RemoveJob(jobID); err != nil {
		log.New().Errorf("Scheduler: failed to remove job %q: %v", jobKey, err)
	}
	delete(s.jobs, jobKey)
}

// runRevoke is the scheduled task: inactivate expired sessions and record the
// run. Errors are logged, never panicked, so a failed run does not crash the
// process or kill the schedule.
func (s *Scheduler) runRevoke(ctx context.Context) {
	now := time.Now().UTC()
	revoked, err := s.userSessionRepo.InactiveExpiredSessions(ctx, now)
	if err != nil {
		log.New().Errorf("Scheduler: revoke expired sessions failed: %v", err)
		return
	}
	result, err := json.Marshal(map[string]int64{"revoked": revoked})
	if err != nil {
		log.New().Errorf("Scheduler: marshal revoke result failed: %v", err)
		return
	}
	if err := s.settingsRepo.RecordScheduleRun(ctx, string(JobSessionRevoke), now, string(result)); err != nil {
		log.New().Errorf("Scheduler: record run failed: %v", err)
		// the revoke still happened — fall through to log it
	}
	log.New().Infof("Session auto-revoke run complete: %d session(s) revoked", revoked)
}

// runCleanup is the scheduled task: prune token_rotation_events older than the
// configured retention window and record the run. The config (and therefore the
// retention window) is read FRESH on each run, so a retention-only change takes
// effect on the next run without a scheduler reload. Errors are logged, never
// panicked, so a failed run does not crash the process or kill the schedule.
func (s *Scheduler) runCleanup(ctx context.Context) {
	now := time.Now().UTC()
	config, err := s.settingsRepo.GetSchedule(ctx, string(JobRotationCleanup))
	if err != nil {
		log.New().Errorf("Scheduler: load rotation-cleanup config failed: %v", err)
		return
	}
	params := rotationCleanupParams{}
	if config.Params != "" {
		if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
			log.New().Errorf("Scheduler: parse rotation-cleanup params failed: %v", err)
			return
		}
	}
	before := rotationCutoff(now, params.RetentionHours)
	cleaned, err := s.userSessionRepo.PruneRotationsBefore(ctx, before)
	if err != nil {
		log.New().Errorf("Scheduler: prune rotation events failed: %v", err)
		return
	}
	result, err := json.Marshal(map[string]int64{"cleaned": cleaned})
	if err != nil {
		log.New().Errorf("Scheduler: marshal cleanup result failed: %v", err)
		return
	}
	if err := s.settingsRepo.RecordScheduleRun(ctx, string(JobRotationCleanup), now, string(result)); err != nil {
		log.New().Errorf("Scheduler: record cleanup run failed: %v", err)
		// the prune still happened — fall through to log it
	}
	log.New().Infof("Rotation cleanup run complete: %d rotation event(s) pruned", cleaned)
}

// rotationCutoff is the pure cutoff calculation: events with rotated_at before
// this time are eligible for pruning.
func rotationCutoff(now time.Time, retentionHours int64) time.Time {
	return now.Add(-time.Duration(retentionHours) * time.Hour)
}

// runActivityCleanup is the scheduled task: prune activity_events older than the
// configured retention window (in DAYS) and record the run. The config (and
// therefore the retention window) is read FRESH on each run, so a retention-only
// change takes effect on the next run without a scheduler reload. Errors are
// logged, never panicked, so a failed run does not crash the process or kill the
// schedule.
func (s *Scheduler) runActivityCleanup(ctx context.Context) {
	now := time.Now().UTC()
	config, err := s.settingsRepo.GetSchedule(ctx, string(JobActivityCleanup))
	if err != nil {
		log.New().Errorf("Scheduler: load activity-cleanup config failed: %v", err)
		return
	}
	params := activityCleanupParams{}
	if config.Params != "" {
		if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
			log.New().Errorf("Scheduler: parse activity-cleanup params failed: %v", err)
			return
		}
	}
	before := activityCutoff(now, params.RetentionDays)
	pruned, err := s.activityRepo.PruneBefore(ctx, before)
	if err != nil {
		log.New().Errorf("Scheduler: prune activity events failed: %v", err)
		return
	}
	result, err := json.Marshal(map[string]int64{"pruned": pruned})
	if err != nil {
		log.New().Errorf("Scheduler: marshal activity-cleanup result failed: %v", err)
		return
	}
	if err := s.settingsRepo.RecordScheduleRun(ctx, string(JobActivityCleanup), now, string(result)); err != nil {
		log.New().Errorf("Scheduler: record activity-cleanup run failed: %v", err)
		// the prune still happened — fall through to log it
	}
	log.New().Infof("Activity cleanup run complete: %d activity event(s) pruned", pruned)
}

// activityCutoff is the pure cutoff calculation: events created before this time
// are eligible for pruning. Retention is in DAYS, so each day is 24 hours.
func activityCutoff(now time.Time, retentionDays int64) time.Time {
	return now.Add(-time.Duration(retentionDays) * 24 * time.Hour)
}
