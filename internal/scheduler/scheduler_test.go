package scheduler

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"

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
