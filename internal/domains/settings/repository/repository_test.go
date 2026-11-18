package repository

import (
	"context"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/domains/settings/entity"
	"github.com/vukyn/isme/internal/testdb"
)

// TestGetScheduleMissingRowReturnsEmpty confirms an unknown job key yields an
// empty struct, not an error.
func TestGetScheduleMissingRowReturnsEmpty(t *testing.T) {
	repo := NewRepository(testdb.Open(t))

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
	repo := NewRepository(testdb.Open(t))

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
