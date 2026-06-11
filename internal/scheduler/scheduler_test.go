package scheduler

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
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

// With both configs seeded disabled (the migration default), Start must install
// no jobs even though the env kill-switch is on.
func TestStartWithBothDisabledInstallsNoJobs(t *testing.T) {
	db := newTestDB(t)
	s, err := New(db, true)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(s.Stop)

	s.Start(context.Background())

	if len(s.jobs) != 0 {
		t.Fatalf("expected 0 jobs installed, got %d (%v)", len(s.jobs), s.jobs)
	}
}

// The job-key strings are a single source of truth (settings entity consts).
// The scheduler JobKey consts and the migration-seeded job_key rows must all
// agree, otherwise a Reload/Get would silently target a non-existent row.
func TestJobKeysAreConsistentAcrossPackages(t *testing.T) {
	if string(JobSessionRevoke) != settingsEntity.JobKeySessionRevoke {
		t.Fatalf("JobSessionRevoke=%q != entity.JobKeySessionRevoke=%q", JobSessionRevoke, settingsEntity.JobKeySessionRevoke)
	}
	if string(JobRotationCleanup) != settingsEntity.JobKeyRotationCleanup {
		t.Fatalf("JobRotationCleanup=%q != entity.JobKeyRotationCleanup=%q", JobRotationCleanup, settingsEntity.JobKeyRotationCleanup)
	}
	if string(JobActivityCleanup) != settingsEntity.JobKeyActivityCleanup {
		t.Fatalf("JobActivityCleanup=%q != entity.JobKeyActivityCleanup=%q", JobActivityCleanup, settingsEntity.JobKeyActivityCleanup)
	}

	// the migration must seed exactly the three job rows the scheduler reads.
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
