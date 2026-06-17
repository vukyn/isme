package di

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
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

// makeBackups creates n dummy app-<timestamp>.db files in dir with ascending
// timestamps, so a descending-name sort orders them newest-first. It returns
// the sorted (newest-first) list of names it created.
func makeBackups(t *testing.T, dir string, n int) []string {
	t.Helper()
	names := make([]string, 0, n)
	base := time.Date(2026, 6, 17, 3, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		name := "app-" + base.Add(time.Duration(i)*time.Minute).Format("20060102-150405") + ".db"
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write backup %s: %v", name, err)
		}
		names = append(names, name)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names
}

func TestPruneBackups(t *testing.T) {
	cases := []struct {
		name        string
		fileCount   int
		retain      int
		wantDeleted int
		wantKept    int
	}{
		{"empty dir", 0, 10, 0, 0},
		{"fewer than retain", 3, 10, 0, 3},
		{"equal to retain", 10, 10, 0, 10},
		{"more than retain", 13, 10, 3, 10},
		{"retain one", 5, 1, 4, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			created := makeBackups(t, dir, tc.fileCount)

			deleted, kept, err := pruneBackups(dir, tc.retain)
			if err != nil {
				t.Fatalf("pruneBackups: %v", err)
			}
			if deleted != tc.wantDeleted {
				t.Fatalf("deleted = %d, want %d", deleted, tc.wantDeleted)
			}
			if kept != tc.wantKept {
				t.Fatalf("kept = %d, want %d", kept, tc.wantKept)
			}

			// the newest tc.wantKept files (created is newest-first) must survive.
			for i, name := range created {
				_, statErr := os.Stat(filepath.Join(dir, name))
				if i < tc.wantKept {
					if statErr != nil {
						t.Fatalf("expected newest backup %s kept, got %v", name, statErr)
					}
				} else if !os.IsNotExist(statErr) {
					t.Fatalf("expected stale backup %s deleted, stat err = %v", name, statErr)
				}
			}
		})
	}
}

// pruneBackups must ignore non-backup files and never delete them.
func TestPruneBackupsIgnoresNonBackupFiles(t *testing.T) {
	dir := t.TempDir()
	makeBackups(t, dir, 5)
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("keep me"), 0o644); err != nil {
		t.Fatalf("write non-backup file: %v", err)
	}

	deleted, kept, err := pruneBackups(dir, 2)
	if err != nil {
		t.Fatalf("pruneBackups: %v", err)
	}
	if deleted != 3 || kept != 2 {
		t.Fatalf("deleted/kept = %d/%d, want 3/2", deleted, kept)
	}
	if _, err := os.Stat(filepath.Join(dir, "README.txt")); err != nil {
		t.Fatalf("non-backup file must survive prune, got %v", err)
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

// The migration must seed exactly the four job rows the scheduler reads, so a
// Reload/Get can never silently target a non-existent row. The job-key strings
// are a single source of truth (settings entity consts).
func TestJobKeysAreConsistentWithMigration(t *testing.T) {
	db := newTestDB(t)
	for _, jobKey := range []string{settingsEntity.JobKeySessionRevoke, settingsEntity.JobKeyRotationCleanup, settingsEntity.JobKeyActivityCleanup, settingsEntity.JobKeyDatabaseBackup} {
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
		pkgScheduler.JobKey(settingsEntity.JobKeyDatabaseBackup),
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
