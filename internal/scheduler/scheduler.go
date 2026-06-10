package scheduler

import (
	"context"
	"sync"
	"time"

	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/log"
)

// IReloader is the seam the settings usecase depends on to re-apply a changed
// schedule. It lives in the scheduler package so the usecase imports the
// scheduler (one direction only) and no import cycle forms: the scheduler
// imports the settings + user_session REPOSITORIES, never their usecases.
type IReloader interface {
	// Reload re-applies the schedule: removes any current job and, when
	// enabled, installs a new cron job for the given expression.
	Reload(ctx context.Context, enabled bool, cronExpr string) error
}

// Scheduler owns a gocron scheduler and the single session-revoke job. It is an
// application-scoped singleton (one per process), constructed directly from the
// *bun.DB so it does not depend on request-scoped DI containers.
type Scheduler struct {
	sched           gocron.Scheduler
	userSessionRepo userSessionRepo.IRepository
	settingsRepo    settingsRepo.IRepository
	enabled         bool

	mu      sync.Mutex
	jobID   uuid.UUID
	hasJob  bool
	started bool
	stopped bool
}

// New constructs the scheduler singleton. enabled is the master kill-switch
// (env SCHEDULER_ENABLED): when false the scheduler never installs a job.
func New(db *bun.DB, enabled bool) (*Scheduler, error) {
	sched, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		sched:           sched,
		userSessionRepo: userSessionRepo.NewRepository(db),
		settingsRepo:    settingsRepo.NewRepository(db),
		enabled:         enabled,
	}, nil
}

// Start reads the persisted config and, if both the env kill-switch and the DB
// flag allow it, installs the cron job. The gocron scheduler is always started
// so a later Reload can add a job without a restart.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sched.Start()
	s.started = true

	if !s.enabled {
		log.New().Info("Session auto-revoke scheduler disabled by SCHEDULER_ENABLED")
		return
	}

	config, err := s.settingsRepo.Get(ctx)
	if err != nil {
		log.New().Errorf("Scheduler: failed to load config: %v", err)
		return
	}
	if !config.Enabled || config.Cron == "" {
		return
	}
	if err := s.applyJob(config.Cron); err != nil {
		log.New().Errorf("Scheduler: failed to install job: %v", err)
		return
	}
	log.New().Infof("Session auto-revoke scheduler started with cron %q", config.Cron)
}

// Reload re-applies the schedule live. It removes the current job (if any) and,
// when enabled (and the env kill-switch allows), installs a fresh cron job.
func (s *Scheduler) Reload(ctx context.Context, enabled bool, cronExpr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil
	}

	s.removeJob()

	if !s.enabled || !enabled || cronExpr == "" {
		log.New().Info("Session auto-revoke scheduler reloaded: no active job")
		return nil
	}
	if err := s.applyJob(cronExpr); err != nil {
		return err
	}
	log.New().Infof("Session auto-revoke scheduler reloaded with cron %q", cronExpr)
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

// applyJob installs a new cron job. Callers must hold s.mu.
func (s *Scheduler) applyJob(cronExpr string) error {
	job, err := s.sched.NewJob(
		gocron.CronJob(cronExpr, false), // 5-field standard cron
		gocron.NewTask(func() { s.runRevoke(context.Background()) }),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return err
	}
	s.jobID = job.ID()
	s.hasJob = true
	return nil
}

// removeJob removes the current job if present. Callers must hold s.mu.
func (s *Scheduler) removeJob() {
	if !s.hasJob {
		return
	}
	if err := s.sched.RemoveJob(s.jobID); err != nil {
		log.New().Errorf("Scheduler: failed to remove job: %v", err)
	}
	s.hasJob = false
	s.jobID = uuid.UUID{}
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
	if err := s.settingsRepo.RecordRun(ctx, now, revoked); err != nil {
		log.New().Errorf("Scheduler: record run failed: %v", err)
		// the revoke still happened — fall through to log it
	}
	log.New().Infof("Session auto-revoke run complete: %d session(s) revoked", revoked)
}
