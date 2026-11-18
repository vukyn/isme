package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	"github.com/vukyn/isme/internal/domains/user_invitation/constants"
	"github.com/vukyn/isme/internal/domains/user_invitation/entity"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// newTestDB opens an in-memory SQLite database and applies every migration
// from db/history/sqlite.
func newTestDB(t *testing.T) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	// a single connection keeps every query on the same in-memory database
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

func newInvitation(email string) entity.UserInvitation {
	return entity.UserInvitation{
		Email:     email,
		TokenHash: "hash-" + email,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
}

func memberAssignments() []entity.UserInvitationRole {
	return []entity.UserInvitationRole{
		{RoleID: "rol_member", AppServiceID: "app_isme"},
	}
}

func TestMarkAcceptedConditionalClaim(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	invitationRepository := NewRepository(db)

	invitationID, err := invitationRepository.Create(ctx, newInvitation("claim@example.com"), memberAssignments())
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}

	// the assignment child row is persisted and loadable
	assignments, err := invitationRepository.GetAssignmentsByInvitationID(ctx, invitationID)
	if err != nil {
		t.Fatalf("get assignments: %v", err)
	}
	if len(assignments) != 1 || assignments[0].RoleID != "rol_member" || assignments[0].AppServiceID != "app_isme" {
		t.Errorf("expected one member assignment, got %+v", assignments)
	}

	// first claim wins
	accepted, err := invitationRepository.MarkAccepted(ctx, invitationID)
	if err != nil {
		t.Fatalf("first MarkAccepted: %v", err)
	}
	if !accepted {
		t.Fatal("expected first MarkAccepted to claim the invitation")
	}

	// second claim must lose — the row is no longer pending
	accepted, err = invitationRepository.MarkAccepted(ctx, invitationID)
	if err != nil {
		t.Fatalf("second MarkAccepted: %v", err)
	}
	if accepted {
		t.Fatal("expected second MarkAccepted to return false")
	}

	stored, err := invitationRepository.GetByID(ctx, invitationID)
	if err != nil {
		t.Fatalf("get invitation: %v", err)
	}
	if stored.Status != int32(constants.InvitationStatusAccepted) {
		t.Errorf("expected accepted status, got %d", stored.Status)
	}
	if stored.AcceptedAt.IsZero() {
		t.Error("expected accepted_at to be set")
	}
}

func TestPendingEmailPartialUniqueIndex(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	invitationRepository := NewRepository(db)

	firstID, err := invitationRepository.Create(ctx, newInvitation("unique@example.com"), memberAssignments())
	if err != nil {
		t.Fatalf("create first invitation: %v", err)
	}

	// a second pending invitation for the same email violates the partial index
	duplicate := newInvitation("unique@example.com")
	duplicate.TokenHash = "hash-other"
	if _, err := invitationRepository.Create(ctx, duplicate, memberAssignments()); err == nil {
		t.Fatal("expected second pending invitation for the same email to fail")
	} else if !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Errorf("expected a unique-constraint error, got: %v", err)
	}

	// after revoking the pending row, a new invitation for the email is allowed
	revoked, err := invitationRepository.MarkRevoked(ctx, firstID)
	if err != nil {
		t.Fatalf("revoke first invitation: %v", err)
	}
	if !revoked {
		t.Fatal("expected revoke to succeed")
	}

	reissued := newInvitation("unique@example.com")
	reissued.TokenHash = "hash-reissue"
	if _, err := invitationRepository.Create(ctx, reissued, memberAssignments()); err != nil {
		t.Fatalf("expected re-issue after revoke to succeed, got: %v", err)
	}
}

func TestMarkExpiredOnlyFlipsPendingPastExpiry(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	invitationRepository := NewRepository(db)

	live, err := invitationRepository.Create(ctx, newInvitation("live@example.com"), memberAssignments())
	if err != nil {
		t.Fatalf("create live invitation: %v", err)
	}
	pastInvitation := newInvitation("past@example.com")
	pastInvitation.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	past, err := invitationRepository.Create(ctx, pastInvitation, memberAssignments())
	if err != nil {
		t.Fatalf("create past invitation: %v", err)
	}

	expired, err := invitationRepository.MarkExpired(ctx, live)
	if err != nil || expired {
		t.Fatalf("expected a live pending invitation not to expire, got expired=%v err=%v", expired, err)
	}

	expired, err = invitationRepository.MarkExpired(ctx, past)
	if err != nil || !expired {
		t.Fatalf("expected a pending invitation past expiry to expire, got expired=%v err=%v", expired, err)
	}
	stored, err := invitationRepository.GetByID(ctx, past)
	if err != nil {
		t.Fatalf("get past invitation: %v", err)
	}
	if stored.Status != int32(constants.InvitationStatusExpired) {
		t.Errorf("expected status %d, got %d", constants.InvitationStatusExpired, stored.Status)
	}

	// a second call finds nothing pending
	expired, err = invitationRepository.MarkExpired(ctx, past)
	if err != nil || expired {
		t.Fatalf("expected an already-expired invitation not to flip again, got expired=%v err=%v", expired, err)
	}
}

func TestExpirePendingBeforeFreesTheEmail(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	invitationRepository := NewRepository(db)

	pastOne := newInvitation("again@example.com")
	pastOne.ExpiresAt = time.Now().UTC().Add(-2 * time.Hour)
	pastOneID, err := invitationRepository.Create(ctx, pastOne, memberAssignments())
	if err != nil {
		t.Fatalf("create past invitation: %v", err)
	}
	pastTwo := newInvitation("other@example.com")
	pastTwo.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	if _, err := invitationRepository.Create(ctx, pastTwo, memberAssignments()); err != nil {
		t.Fatalf("create second past invitation: %v", err)
	}
	liveID, err := invitationRepository.Create(ctx, newInvitation("live@example.com"), memberAssignments())
	if err != nil {
		t.Fatalf("create live invitation: %v", err)
	}
	revokedPast := newInvitation("revoked@example.com")
	revokedPast.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	revokedID, err := invitationRepository.Create(ctx, revokedPast, memberAssignments())
	if err != nil {
		t.Fatalf("create revoked invitation: %v", err)
	}
	if _, err := invitationRepository.MarkRevoked(ctx, revokedID); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	count, err := invitationRepository.ExpirePendingBefore(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("expire pending: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 invitations expired, got %d", count)
	}

	for id, want := range map[string]int32{
		pastOneID: int32(constants.InvitationStatusExpired),
		liveID:    int32(constants.InvitationStatusPending),
		revokedID: int32(constants.InvitationStatusRevoked),
	} {
		stored, err := invitationRepository.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		if stored.Status != want {
			t.Errorf("invitation %s: expected status %d, got %d", id, want, stored.Status)
		}
	}

	// the partial pending-email index no longer holds the expired row
	fresh := newInvitation("again@example.com")
	fresh.TokenHash = "hash-again-fresh"
	if _, err := invitationRepository.Create(ctx, fresh, memberAssignments()); err != nil {
		t.Errorf("expected a new pending invitation for the expired email, got %v", err)
	}
}
