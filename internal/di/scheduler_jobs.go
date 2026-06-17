package di

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/vukyn/isme/internal/constants"
	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	"github.com/uptrace/bun"
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

// databaseBackupParams mirrors the params JSON of the database-backup job.
// Retention is a COUNT of backup files to keep (not a time window), so pruning
// deletes the oldest files past that count.
type databaseBackupParams struct {
	RetainCount int64 `json:"retain_count"`
}

// defaultBackupRetainCount is the fallback number of backup files to keep when
// the params blob carries a zero/missing retain_count.
const defaultBackupRetainCount int64 = 10

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

// newDatabaseBackupRun returns the database-backup job body: snapshot the
// file-based SQLite database via VACUUM INTO into db/backups/, then prune the
// oldest backups beyond the configured retain count and record the run. The
// retain count is read FRESH on each run. ALL errors are logged, never
// panicked, so a failed backup does not crash the process or kill the schedule.
func newDatabaseBackupRun(
	db *bun.DB,
	settingsRepository settingsRepo.IRepository,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		now := time.Now().UTC()

		config, err := settingsRepository.GetSchedule(ctx, settingsEntity.JobKeyDatabaseBackup)
		if err != nil {
			log.New().Errorf("Scheduler: load database-backup config failed: %v", err)
			return nil
		}
		params := databaseBackupParams{}
		if config.Params != "" {
			if err := json.Unmarshal([]byte(config.Params), &params); err != nil {
				log.New().Errorf("Scheduler: parse database-backup params failed: %v", err)
				return nil
			}
		}
		if params.RetainCount <= 0 {
			params.RetainCount = defaultBackupRetainCount
		}

		dbDir := filepath.Dir(constants.DB_FILE_PATH)
		backupDir := filepath.Join(dbDir, "backups")
		if err := os.MkdirAll(backupDir, 0o755); err != nil {
			log.New().Errorf("Scheduler: create backup dir failed: %v", err)
			return nil
		}

		target := filepath.Join(backupDir, "app-"+now.Format("20060102-150405")+".db")
		// Try the bound-param form first. The path is a server-generated
		// timestamp (not user input), so inlining it into the SQL string on
		// fallback is injection-safe.
		if _, err := db.ExecContext(ctx, "VACUUM INTO ?", target); err != nil {
			if _, err := db.ExecContext(ctx, "VACUUM INTO '"+target+"'"); err != nil {
				log.New().Errorf("Scheduler: database backup (VACUUM INTO) failed: %v", err)
				return nil
			}
		}

		var bytes int64
		if info, err := os.Stat(target); err == nil {
			bytes = info.Size()
		}

		deleted, kept, err := pruneBackups(backupDir, int(params.RetainCount))
		if err != nil {
			log.New().Errorf("Scheduler: prune backups failed: %v", err)
			return nil
		}

		result, err := json.Marshal(map[string]any{
			"backup_path": target,
			"kept":        kept,
			"deleted":     deleted,
			"bytes":       bytes,
		})
		if err != nil {
			log.New().Errorf("Scheduler: marshal database-backup result failed: %v", err)
			return nil
		}
		if err := settingsRepository.RecordScheduleRun(ctx, settingsEntity.JobKeyDatabaseBackup, now, string(result)); err != nil {
			log.New().Errorf("Scheduler: record database-backup run failed: %v", err)
			// the backup still happened — fall through to log it
		}
		log.New().Infof("Database backup run complete: %s (kept %d, pruned %d)", target, kept, deleted)
		return nil
	}
}

// pruneBackups deletes the oldest backup files in backupDir, keeping at most
// retainCount of the newest. Backups are named app-<timestamp>.db, so a
// descending name sort orders them newest-first; everything past retainCount is
// deleted. Returns the count deleted and the count kept.
func pruneBackups(backupDir string, retainCount int) (deleted int, kept int, err error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return 0, 0, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if matched, _ := filepath.Match("app-*.db", name); matched {
			names = append(names, name)
		}
	}
	// Descending: timestamp names sort lexicographically by recency, so the
	// newest backups come first and stale ones fall to the tail.
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	if retainCount < 0 {
		retainCount = 0
	}
	if retainCount >= len(names) {
		return 0, len(names), nil
	}
	for _, name := range names[retainCount:] {
		if removeErr := os.Remove(filepath.Join(backupDir, name)); removeErr != nil {
			return deleted, retainCount, removeErr
		}
		deleted++
	}
	return deleted, retainCount, nil
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
