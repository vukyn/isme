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
