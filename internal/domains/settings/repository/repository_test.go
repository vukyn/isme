package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	"github.com/vukyn/isme/internal/domains/settings/entity"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// newTestDB opens an in-memory SQLite database and applies every migration so
// the schedule_config table + seeded job rows exist.
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

// TestGetScheduleMissingRowReturnsEmpty confirms an unknown job key yields an
// empty struct, not an error.
func TestGetScheduleMissingRowReturnsEmpty(t *testing.T) {
	repo := NewRepository(newTestDB(t))

	config, err := repo.GetSchedule(context.Background(), "does_not_exist")
	if err != nil {
		t.Fatalf("GetSchedule: %v", err)
	}
	if config.JobKey != "" || config.Enabled || config.Cron != "" {
		t.Fatalf("expected empty config for missing row, got %+v", config)
	}
}

// TestUpdateAndRecordRoundTripThroughJSONColumns persists enabled/cron/params
// then a run result, and reads both back through the generic schedule_config
// row — exercising the real SQL path and the params/last_result JSON columns.
func TestUpdateAndRecordRoundTripThroughJSONColumns(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(newTestDB(t))

	if err := repo.UpdateSchedule(ctx, entity.JobKeyRotationCleanup, true, "0 4 * * *", `{"retention_hours":96}`, "usr_admin"); err != nil {
		t.Fatalf("UpdateSchedule: %v", err)
	}

	ranAt := time.Unix(1700000000, 0).UTC()
	if err := repo.RecordScheduleRun(ctx, entity.JobKeyRotationCleanup, ranAt, `{"cleaned":1284}`); err != nil {
		t.Fatalf("RecordScheduleRun: %v", err)
	}

	config, err := repo.GetSchedule(ctx, entity.JobKeyRotationCleanup)
	if err != nil {
		t.Fatalf("GetSchedule: %v", err)
	}
	if !config.Enabled || config.Cron != "0 4 * * *" {
		t.Fatalf("enabled/cron mismatch: %+v", config)
	}
	if config.Params != `{"retention_hours":96}` {
		t.Fatalf("params did not round-trip: %q", config.Params)
	}
	if config.UpdatedBy != "usr_admin" {
		t.Fatalf("updated_by mismatch: %q", config.UpdatedBy)
	}
	if config.LastRunAt == nil || !config.LastRunAt.Equal(ranAt) {
		t.Fatalf("last_run_at mismatch: %+v", config.LastRunAt)
	}
	if config.LastResult == nil || *config.LastResult != `{"cleaned":1284}` {
		t.Fatalf("last_result did not round-trip: %+v", config.LastResult)
	}
}
