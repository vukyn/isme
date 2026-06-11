package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	activityConstants "github.com/vukyn/isme/internal/domains/activity/constants"
	"github.com/vukyn/isme/internal/domains/activity/entity"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// newTestDB opens an in-memory SQLite database and applies every migration
// (including 026, which creates activity_events).
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

// TestCreateGeneratesIDAndDefaultsMeta proves an empty id is auto-filled with a
// ULID and an empty meta defaults to "{}".
func TestCreateGeneratesIDAndDefaultsMeta(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)

	err := repository.Create(context.Background(), entity.ActivityEvent{
		UserID: "user-1",
		Type:   activityConstants.ActivityTypeSignOut,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	events, err := repository.ListByUserID(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID == "" {
		t.Error("expected a generated ULID id")
	}
	if events[0].Meta != "{}" {
		t.Errorf("expected default meta '{}', got %q", events[0].Meta)
	}
}

// TestListByUserIDNewestFirst proves results are ordered created_at DESC.
func TestListByUserIDNewestFirst(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)

	// Insert in a known order. BeforeAppendModel stamps created_at at insert time;
	// distinct ids let us identify them regardless of equal-second timestamps.
	for _, typ := range []string{
		activityConstants.ActivityTypeSignIn,
		activityConstants.ActivityTypePasswordChanged,
		activityConstants.ActivityTypeSignOut,
	} {
		if err := repository.Create(context.Background(), entity.ActivityEvent{
			ID:     "id-" + typ,
			UserID: "user-1",
			Type:   typ,
			Meta:   "{}",
		}); err != nil {
			t.Fatalf("create %s: %v", typ, err)
		}
	}

	events, err := repository.ListByUserID(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	// newest first: each event's created_at must be >= the next one's.
	for i := 0; i+1 < len(events); i++ {
		if events[i].CreatedAt.Before(events[i+1].CreatedAt) {
			t.Errorf("events not ordered newest-first at index %d", i)
		}
	}
}

// TestListByUserIDLimit proves the limit caps the number of rows returned.
func TestListByUserIDLimit(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)

	for i := 0; i < 5; i++ {
		if err := repository.Create(context.Background(), entity.ActivityEvent{
			UserID: "user-1",
			Type:   activityConstants.ActivityTypeSignIn,
		}); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	events, err := repository.ListByUserID(context.Background(), "user-1", 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected limit of 2, got %d", len(events))
	}
}

// TestListByUserIDScopedToUser proves a user only sees their own events.
func TestListByUserIDScopedToUser(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)

	if err := repository.Create(context.Background(), entity.ActivityEvent{UserID: "user-1", Type: activityConstants.ActivityTypeSignIn}); err != nil {
		t.Fatalf("create user-1: %v", err)
	}
	if err := repository.Create(context.Background(), entity.ActivityEvent{UserID: "user-2", Type: activityConstants.ActivityTypeSignIn}); err != nil {
		t.Fatalf("create user-2: %v", err)
	}

	events, err := repository.ListByUserID(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected only user-1's event, got %d", len(events))
	}
	if events[0].UserID != "user-1" {
		t.Errorf("expected user-1, got %q", events[0].UserID)
	}
}

// TestListByUserIDEmptyNonNil proves an empty result is a non-nil slice so the
// JSON encoder emits [] not null (the FE feed iterates the array).
func TestListByUserIDEmptyNonNil(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)

	events, err := repository.ListByUserID(context.Background(), "nobody", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if events == nil {
		t.Fatal("expected a non-nil empty slice, got nil")
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// seedEventAt inserts an event and forces its created_at to the given time
// (BeforeAppendModel stamps time.Now on insert, so we overwrite the column).
func seedEventAt(t *testing.T, db *bun.DB, id string, createdAt time.Time) {
	t.Helper()
	if err := (&repository{db: db}).Create(context.Background(), entity.ActivityEvent{
		ID:     id,
		UserID: "user-1",
		Type:   activityConstants.ActivityTypeSignIn,
	}); err != nil {
		t.Fatalf("seed create %s: %v", id, err)
	}
	if _, err := db.NewUpdate().
		Model((*entity.ActivityEvent)(nil)).
		Set("created_at = ?", createdAt).
		Where("id = ?", id).
		Exec(context.Background()); err != nil {
		t.Fatalf("seed set created_at %s: %v", id, err)
	}
}

// TestPruneBeforeDeletesOnlyOlderRows proves the cutoff is exclusive on the
// recent side: rows created_at < before are deleted, newer rows survive.
func TestPruneBeforeDeletesOnlyOlderRows(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)
	now := time.Now().UTC()

	seedEventAt(t, db, "old-1", now.Add(-100*24*time.Hour))
	seedEventAt(t, db, "old-2", now.Add(-91*24*time.Hour))
	seedEventAt(t, db, "recent", now.Add(-1*24*time.Hour))

	pruned, err := repository.PruneBefore(context.Background(), now.Add(-90*24*time.Hour))
	if err != nil {
		t.Fatalf("PruneBefore: %v", err)
	}
	if pruned != 2 {
		t.Fatalf("expected 2 rows pruned, got %d", pruned)
	}

	events, err := repository.ListByUserID(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 1 || events[0].ID != "recent" {
		t.Fatalf("expected only the recent event to remain, got %+v", events)
	}
}

// TestPruneBeforeEmptyTable proves pruning an empty table is a no-op returning 0.
func TestPruneBeforeEmptyTable(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)

	pruned, err := repository.PruneBefore(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("PruneBefore: %v", err)
	}
	if pruned != 0 {
		t.Fatalf("expected 0 rows pruned on empty table, got %d", pruned)
	}
}
