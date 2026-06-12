package di

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	pkgScheduler "github.com/vukyn/kuery/scheduler"
)

func TestRotationCutoff(t *testing.T) {
	now := time.Date(2026, 6, 11, 4, 0, 0, 0, time.UTC)
	cases := []struct {
		name           string
		retentionHours int64
		want           time.Time
	}{
		{"48h", 48, now.Add(-48 * time.Hour)},
		{"24h floor", 24, now.Add(-24 * time.Hour)},
		{"7d", 168, now.Add(-168 * time.Hour)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rotationCutoff(now, tc.retentionHours); !got.Equal(tc.want) {
				t.Fatalf("rotationCutoff(%v) = %v, want %v", tc.retentionHours, got, tc.want)
			}
		})
	}
}

func TestActivityCutoff(t *testing.T) {
	now := time.Date(2026, 6, 11, 5, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		retentionDays int64
		want          time.Time
	}{
		{"30d", 30, now.Add(-30 * 24 * time.Hour)},
		{"7d floor", 7, now.Add(-7 * 24 * time.Hour)},
		{"90d default", 90, now.Add(-90 * 24 * time.Hour)},
		{"365d", 365, now.Add(-365 * 24 * time.Hour)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := activityCutoff(now, tc.retentionDays); !got.Equal(tc.want) {
				t.Fatalf("activityCutoff(%v) = %v, want %v", tc.retentionDays, got, tc.want)
			}
		})
	}
}

// newTestDB opens an in-memory SQLite database and applies every migration.
func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	sqldb.SetMaxOpenConns(1)
	db := bun.NewDB(sqldb, sqlitedialect.New())
	for _, migration := range sqliteHistory.Migrations {
		if err := migration.Up(db); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Name, err)
		}
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// The migration must seed exactly the three job rows the scheduler reads, so a
// Reload/Get can never silently target a non-existent row. The job-key strings
// are a single source of truth (settings entity consts).
func TestJobKeysAreConsistentWithMigration(t *testing.T) {
	db := newTestDB(t)
	for _, jobKey := range []string{settingsEntity.JobKeySessionRevoke, settingsEntity.JobKeyRotationCleanup, settingsEntity.JobKeyActivityCleanup} {
		var count int
		row := db.QueryRow("SELECT COUNT(*) FROM schedule_config WHERE job_key = ?", jobKey)
		if err := row.Scan(&count); err != nil {
			t.Fatalf("query schedule_config for %q: %v", jobKey, err)
		}
		if count != 1 {
			t.Fatalf("expected exactly 1 schedule_config row for %q, got %d", jobKey, count)
		}
	}
}

// With both configs seeded disabled (the migration default), the schedule
// provider reports every job as not-enabled, so the engine installs no job at
// Start. This is the parity check for the old "both disabled => 0 jobs" test.
func TestScheduleProviderReportsSeededJobsDisabled(t *testing.T) {
	db := newTestDB(t)
	provider := newScheduleProvider(settingsRepo.NewRepository(db))

	for _, jobKey := range []pkgScheduler.JobKey{
		pkgScheduler.JobKey(settingsEntity.JobKeySessionRevoke),
		pkgScheduler.JobKey(settingsEntity.JobKeyRotationCleanup),
		pkgScheduler.JobKey(settingsEntity.JobKeyActivityCleanup),
	} {
		enabled, _, err := provider.Load(context.Background(), jobKey)
		if err != nil {
			t.Fatalf("provider.Load(%q): %v", jobKey, err)
		}
		if enabled {
			t.Fatalf("expected job %q seeded disabled, got enabled", jobKey)
		}
	}
}
