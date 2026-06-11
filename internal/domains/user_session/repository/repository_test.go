package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	"github.com/vukyn/isme/internal/domains/user_session/constants"
	"github.com/vukyn/isme/internal/domains/user_session/entity"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

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

func insertSession(t *testing.T, db *bun.DB, id string, status int32, expiresAt time.Time) {
	t.Helper()
	session := entity.UserSession{
		ID:        id,
		UserID:    "user_" + id,
		Email:     id + "@example.com",
		Status:    status,
		ExpiresAt: expiresAt,
		ClientIP:  "127.0.0.1",
		TokenID:   "tok_" + id,
	}
	if _, err := db.NewInsert().Model(&session).Exec(context.Background()); err != nil {
		t.Fatalf("insert session %s: %v", id, err)
	}
}

func statusOf(t *testing.T, db *bun.DB, id string) int32 {
	t.Helper()
	session := entity.UserSession{}
	if err := db.NewSelect().Model(&session).Where("id = ?", id).Scan(context.Background()); err != nil {
		t.Fatalf("scan session %s: %v", id, err)
	}
	return session.Status
}

func TestInactiveExpiredSessions(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)
	now := time.Now().UTC()

	// active + expired -> should flip to inactive and be counted
	insertSession(t, db, "expired_active", constants.UserSessionStatusActive, now.Add(-time.Hour))
	// active + future -> untouched
	insertSession(t, db, "future_active", constants.UserSessionStatusActive, now.Add(time.Hour))
	// already inactive + expired -> untouched (not counted)
	insertSession(t, db, "expired_inactive", constants.UserSessionStatusInactive, now.Add(-time.Hour))

	count, err := repository.InactiveExpiredSessions(context.Background(), now)
	if err != nil {
		t.Fatalf("InactiveExpiredSessions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 session revoked, got %d", count)
	}

	if got := statusOf(t, db, "expired_active"); got != constants.UserSessionStatusInactive {
		t.Errorf("expired_active: expected inactive(%d), got %d", constants.UserSessionStatusInactive, got)
	}
	if got := statusOf(t, db, "future_active"); got != constants.UserSessionStatusActive {
		t.Errorf("future_active: expected active(%d), got %d", constants.UserSessionStatusActive, got)
	}
	if got := statusOf(t, db, "expired_inactive"); got != constants.UserSessionStatusInactive {
		t.Errorf("expired_inactive: expected inactive(%d), got %d", constants.UserSessionStatusInactive, got)
	}
}

func TestInactiveExpiredSessionsNoMatches(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)
	now := time.Now().UTC()

	insertSession(t, db, "future_active", constants.UserSessionStatusActive, now.Add(time.Hour))

	count, err := repository.InactiveExpiredSessions(context.Background(), now)
	if err != nil {
		t.Fatalf("InactiveExpiredSessions: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 sessions revoked, got %d", count)
	}
}

func insertRotation(t *testing.T, db *bun.DB, id string, rotatedAt time.Time) {
	t.Helper()
	event := entity.TokenRotationEvent{
		ID:        id,
		UserID:    "user_" + id,
		SessionID: "sess_" + id,
		RotatedAt: rotatedAt,
	}
	if _, err := db.NewInsert().Model(&event).Exec(context.Background()); err != nil {
		t.Fatalf("insert rotation %s: %v", id, err)
	}
}

func rotationIDs(t *testing.T, db *bun.DB) map[string]bool {
	t.Helper()
	var events []entity.TokenRotationEvent
	if err := db.NewSelect().Model(&events).Scan(context.Background()); err != nil {
		t.Fatalf("scan rotations: %v", err)
	}
	ids := make(map[string]bool, len(events))
	for _, event := range events {
		ids[event.ID] = true
	}
	return ids
}

func TestPruneRotationsBefore(t *testing.T) {
	db := newTestDB(t)
	repository := NewRepository(db)
	now := time.Now().UTC()
	cutoff := now.Add(-48 * time.Hour)

	// older than cutoff -> pruned
	insertRotation(t, db, "old_a", now.Add(-72*time.Hour))
	insertRotation(t, db, "old_b", now.Add(-49*time.Hour))
	// at-or-after cutoff -> survive
	insertRotation(t, db, "edge_at_cutoff", cutoff)
	insertRotation(t, db, "recent", now.Add(-1*time.Hour))

	cleaned, err := repository.PruneRotationsBefore(context.Background(), cutoff)
	if err != nil {
		t.Fatalf("PruneRotationsBefore: %v", err)
	}
	if cleaned != 2 {
		t.Fatalf("expected 2 rotation events pruned, got %d", cleaned)
	}

	survivors := rotationIDs(t, db)
	if len(survivors) != 2 {
		t.Fatalf("expected 2 survivors, got %d (%v)", len(survivors), survivors)
	}
	if !survivors["edge_at_cutoff"] || !survivors["recent"] {
		t.Fatalf("unexpected survivors: %v", survivors)
	}
	if survivors["old_a"] || survivors["old_b"] {
		t.Fatalf("pruned events still present: %v", survivors)
	}
}
