package di

import (
	"context"
	"encoding/json"
	"time"

	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	pkgScheduler "github.com/vukyn/kuery/scheduler"

	"github.com/vukyn/kuery/log"
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

// scheduleProvider adapts the settings repository to the kuery scheduler's
// ScheduleProvider seam: at Start the engine asks it for each registered job's
// persisted schedule. A job is installed only when its DB row is enabled and
// carries a non-empty cron expression.
type scheduleProvider struct {
	settingsRepo settingsRepo.IRepository
}

// Load returns the persisted enabled flag and schedule for the named job. The
// effective-enabled gate (enabled && cron != "") matches the original engine's
// per-job load checks exactly.
func (p *scheduleProvider) Load(ctx context.Context, key pkgScheduler.JobKey) (bool, pkgScheduler.Schedule, error) {
	config, err := p.settingsRepo.GetSchedule(ctx, string(key))
	if err != nil {
		return false, pkgScheduler.Schedule{}, err
	}
	enabled := config.Enabled && config.Cron != ""
	return enabled, pkgScheduler.Cron(config.Cron), nil
}

// newScheduleProvider builds the schedule provider over the settings repository.
func newScheduleProvider(settingsRepository settingsRepo.IRepository) pkgScheduler.ScheduleProvider {
	return &scheduleProvider{settingsRepo: settingsRepository}
}

// newSessionRevokeRun returns the session-revoke job body: inactivate expired
// sessions and record the run. Errors are logged, never panicked, so a failed
// run does not crash the process or kill the schedule.
func newSessionRevokeRun(
	userSessionRepository userSessionRepo.IRepository,
	settingsRepository settingsRepo.IRepository,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		now := time.Now().UTC()
		revoked, err := userSessionRepository.InactiveExpiredSessions(ctx, now)
		if err != nil {
			log.New().Errorf("Scheduler: revoke expired sessions failed: %v", err)
			return nil
		}
		result, err := json.Marshal(map[string]int64{"revoked": revoked})
		if err != nil {
			log.New().Errorf("Scheduler: marshal revoke result failed: %v", err)
			return nil
		}
		if err := settingsRepository.RecordScheduleRun(ctx, settingsEntity.JobKeySessionRevoke, now, string(result)); err != nil {
			log.New().Errorf("Scheduler: record run failed: %v", err)
			// the revoke still happened — fall through to log it
		}
		log.New().Infof("Session auto-revoke run complete: %d session(s) revoked", revoked)
		return nil
	}
}

// newRotationCleanupRun returns the rotation-cleanup job body: prune
// token_rotation_events older than the configured retention window and record
// the run. The config (and therefore the retention window) is read FRESH on
// each run, so a retention-only change takes effect on the next run without a
// scheduler reload. Errors are logged, never panicked.
func newRotationCleanupRun(
	userSessionRepository userSessionRepo.IRepository,
	settingsRepository settingsRepo.IRepository,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		now := time.Now().UTC()
		config, err := settingsRepository.GetSchedule(ctx, settingsEntity.JobKeyRotationCleanup)
		if err != nil {
			log.New().Errorf("Scheduler: load rotation-cleanup config failed: %v", err)
			return nil
		}
		params := rotationCleanupParams{}
		if config.Params != "" {
			if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
				log.New().Errorf("Scheduler: parse rotation-cleanup params failed: %v", err)
				return nil
			}
		}
		before := rotationCutoff(now, params.RetentionHours)
		cleaned, err := userSessionRepository.PruneRotationsBefore(ctx, before)
		if err != nil {
			log.New().Errorf("Scheduler: prune rotation events failed: %v", err)
			return nil
		}
		result, err := json.Marshal(map[string]int64{"cleaned": cleaned})
		if err != nil {
			log.New().Errorf("Scheduler: marshal cleanup result failed: %v", err)
			return nil
		}
		if err := settingsRepository.RecordScheduleRun(ctx, settingsEntity.JobKeyRotationCleanup, now, string(result)); err != nil {
			log.New().Errorf("Scheduler: record cleanup run failed: %v", err)
			// the prune still happened — fall through to log it
		}
		log.New().Infof("Rotation cleanup run complete: %d rotation event(s) pruned", cleaned)
		return nil
	}
}

// newActivityCleanupRun returns the activity-cleanup job body: prune
// activity_events older than the configured retention window (in DAYS) and
// record the run. The retention window is read FRESH on each run, so a
// retention-only change takes effect on the next run without a scheduler
// reload. Errors are logged, never panicked.
func newActivityCleanupRun(
	activityRepository activityRepo.IRepository,
	settingsRepository settingsRepo.IRepository,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		now := time.Now().UTC()
		config, err := settingsRepository.GetSchedule(ctx, settingsEntity.JobKeyActivityCleanup)
		if err != nil {
			log.New().Errorf("Scheduler: load activity-cleanup config failed: %v", err)
			return nil
		}
		params := activityCleanupParams{}
		if config.Params != "" {
			if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
				log.New().Errorf("Scheduler: parse activity-cleanup params failed: %v", err)
				return nil
			}
		}
		before := activityCutoff(now, params.RetentionDays)
		pruned, err := activityRepository.PruneBefore(ctx, before)
		if err != nil {
			log.New().Errorf("Scheduler: prune activity events failed: %v", err)
			return nil
		}
		result, err := json.Marshal(map[string]int64{"pruned": pruned})
		if err != nil {
			log.New().Errorf("Scheduler: marshal activity-cleanup result failed: %v", err)
			return nil
		}
		if err := settingsRepository.RecordScheduleRun(ctx, settingsEntity.JobKeyActivityCleanup, now, string(result)); err != nil {
			log.New().Errorf("Scheduler: record activity-cleanup run failed: %v", err)
			// the prune still happened — fall through to log it
		}
		log.New().Infof("Activity cleanup run complete: %d activity event(s) pruned", pruned)
		return nil
	}
}

// rotationCutoff is the pure cutoff calculation: events with rotated_at before
// this time are eligible for pruning.
func rotationCutoff(now time.Time, retentionHours int64) time.Time {
	return now.Add(-time.Duration(retentionHours) * time.Hour)
}

// activityCutoff is the pure cutoff calculation: events created before this time
// are eligible for pruning. Retention is in DAYS, so each day is 24 hours.
func activityCutoff(now time.Time, retentionDays int64) time.Time {
	return now.Add(-time.Duration(retentionDays) * 24 * time.Hour)
}
