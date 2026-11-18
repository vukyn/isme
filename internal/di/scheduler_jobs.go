package di

import (
	"context"
	"encoding/json"
	"time"

	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsModels "github.com/vukyn/isme/internal/domains/settings/models"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	userInvitationRepo "github.com/vukyn/isme/internal/domains/user_invitation/repository"
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

// unverifiedUserCleanupParams mirrors the params JSON of the
// unverified-user-cleanup job. The minimum age is in DAYS.
type unverifiedUserCleanupParams struct {
	MinAgeDays int64 `json:"min_age_days"`
}

// unverifiedUserCleanupActor is stamped as deleted_by on accounts the
// unverified-user-cleanup job removes.
const unverifiedUserCleanupActor = "system:unverified_user_cleanup"

// transactionRunner runs fn inside one database transaction that every
// repository call made with fn's context joins (kuery/bun/tx).
type transactionRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
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

// newInvitationCleanupRun returns the invitation-cleanup job body: mark every
// pending invitation past its expiry as expired and record the run. Errors are
// logged, never panicked.
func newInvitationCleanupRun(
	invitationRepository userInvitationRepo.IRepository,
	settingsRepository settingsRepo.IRepository,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		now := time.Now().UTC()
		expired, err := invitationRepository.ExpirePendingBefore(ctx, now)
		if err != nil {
			log.New().Errorf("Scheduler: expire pending invitations failed: %v", err)
			return nil
		}
		result, err := json.Marshal(map[string]int64{"expired": expired})
		if err != nil {
			log.New().Errorf("Scheduler: marshal invitation-cleanup result failed: %v", err)
			return nil
		}
		if err := settingsRepository.RecordScheduleRun(ctx, settingsEntity.JobKeyInvitationCleanup, now, string(result)); err != nil {
			log.New().Errorf("Scheduler: record invitation-cleanup run failed: %v", err)
			// the expiry still happened — fall through to log it
		}
		log.New().Infof("Invitation cleanup run complete: %d invitation(s) expired", expired)
		return nil
	}
}

// newUnverifiedUserCleanupRun returns the unverified-user-cleanup job body:
// soft-delete live accounts that never finished sign-up and are older than the
// configured minimum age, revoke their sessions and drop their role
// memberships, all in one transaction, then record the run. The minimum age is
// read FRESH on each run; a value below the floor skips the run rather than
// widening the deletion. Errors are logged, never panicked.
func newUnverifiedUserCleanupRun(
	userRepository userRepo.IRepository,
	userSessionRepository userSessionRepo.IRepository,
	roleRepository roleRepo.IRepository,
	transactions transactionRunner,
	settingsRepository settingsRepo.IRepository,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		now := time.Now().UTC()
		config, err := settingsRepository.GetSchedule(ctx, settingsEntity.JobKeyUnverifiedUserCleanup)
		if err != nil {
			log.New().Errorf("Scheduler: load unverified-user-cleanup config failed: %v", err)
			return nil
		}
		params := unverifiedUserCleanupParams{}
		if config.Params != "" {
			if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
				log.New().Errorf("Scheduler: parse unverified-user-cleanup params failed: %v", err)
				return nil
			}
		}
		if params.MinAgeDays < settingsModels.UnverifiedUserMinAgeFloorDays {
			log.New().Errorf("Scheduler: unverified-user-cleanup min_age_days %d is below the floor, run skipped", params.MinAgeDays)
			return nil
		}

		userIDs, err := userRepository.ListUnverifiedCreatedBefore(ctx, activityCutoff(now, params.MinAgeDays))
		if err != nil {
			log.New().Errorf("Scheduler: list unverified users failed: %v", err)
			return nil
		}

		var deleted int64
		if len(userIDs) > 0 {
			err = transactions.RunInTx(ctx, func(ctx context.Context) error {
				count, err := userRepository.SoftDeleteByIDs(ctx, userIDs, unverifiedUserCleanupActor)
				if err != nil {
					return err
				}
				for _, userID := range userIDs {
					if err := userSessionRepository.InactiveAllUserSession(ctx, userID); err != nil {
						return err
					}
				}
				if err := roleRepository.RemoveAllMembershipsByUserIDs(ctx, userIDs); err != nil {
					return err
				}
				deleted = count
				return nil
			})
			if err != nil {
				log.New().Errorf("Scheduler: delete unverified users failed: %v", err)
				return nil
			}
		}

		result, err := json.Marshal(map[string]int64{"deleted": deleted})
		if err != nil {
			log.New().Errorf("Scheduler: marshal unverified-user-cleanup result failed: %v", err)
			return nil
		}
		if err := settingsRepository.RecordScheduleRun(ctx, settingsEntity.JobKeyUnverifiedUserCleanup, now, string(result)); err != nil {
			log.New().Errorf("Scheduler: record unverified-user-cleanup run failed: %v", err)
			// the deletion still happened — fall through to log it
		}
		log.New().Infof("Unverified user cleanup run complete: %d account(s) deleted %v", deleted, userIDs)
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
