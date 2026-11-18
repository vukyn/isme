package di

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	settingsEntity "github.com/vukyn/isme/internal/domains/settings/entity"
	settingsRepo "github.com/vukyn/isme/internal/domains/settings/repository"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	"github.com/vukyn/isme/internal/domains/user_invitation/constants"
	userInvitationEntity "github.com/vukyn/isme/internal/domains/user_invitation/entity"
	userInvitationRepo "github.com/vukyn/isme/internal/domains/user_invitation/repository"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	"github.com/uptrace/bun"
	kueryDb "github.com/vukyn/kuery/bun/db"
	pkgTx "github.com/vukyn/kuery/bun/tx"
)

// newFileTestDB opens a file SQLite database with the service's pool settings,
// so a write that escapes a transaction fails instead of hanging on a single
// shared connection.
func newFileTestDB(t *testing.T) *bun.DB {
	t.Helper()
	db, err := kueryDb.Open(kueryDb.Config{
		Driver:     kueryDb.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "jobs.db"),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	for _, migration := range sqliteHistory.Migrations {
		if err := migration.Up(db); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Name, err)
		}
	}
	return db
}

func lastResult(t *testing.T, settingsRepository settingsRepo.IRepository, jobKey string) map[string]int64 {
	t.Helper()
	config, err := settingsRepository.GetSchedule(context.Background(), jobKey)
	if err != nil {
		t.Fatalf("get schedule %s: %v", jobKey, err)
	}
	if config.LastResult == nil || config.LastRunAt == nil {
		t.Fatalf("expected job %s to record a run", jobKey)
	}
	result := map[string]int64{}
	if err := json.Unmarshal([]byte(*config.LastResult), &result); err != nil {
		t.Fatalf("parse result: %v", err)
	}
	return result
}

func TestInvitationCleanupRunExpiresOnlyPendingPastExpiry(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	invitationRepository := userInvitationRepo.NewRepository(db)
	settingsRepository := settingsRepo.NewRepository(db)

	create := func(email string, expiresAt time.Time) string {
		id, err := invitationRepository.Create(ctx, userInvitationEntity.UserInvitation{
			Email: email, TokenHash: "hash-" + email, ExpiresAt: expiresAt,
		}, []userInvitationEntity.UserInvitationRole{{RoleID: "rol_member", AppServiceID: "app_isme"}})
		if err != nil {
			t.Fatalf("create invitation %s: %v", email, err)
		}
		return id
	}
	pastID := create("past@hasaki.vn", time.Now().UTC().Add(-time.Hour))
	liveID := create("live@hasaki.vn", time.Now().UTC().Add(time.Hour))
	acceptedID := create("accepted@hasaki.vn", time.Now().UTC().Add(time.Hour))
	if _, err := invitationRepository.MarkAccepted(ctx, acceptedID); err != nil {
		t.Fatalf("accept: %v", err)
	}

	if err := newInvitationCleanupRun(invitationRepository, settingsRepository)(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}

	for id, want := range map[string]int32{
		pastID:     int32(constants.InvitationStatusExpired),
		liveID:     int32(constants.InvitationStatusPending),
		acceptedID: int32(constants.InvitationStatusAccepted),
	} {
		stored, err := invitationRepository.GetByID(ctx, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		if stored.Status != want {
			t.Errorf("invitation %s: expected status %d, got %d", id, want, stored.Status)
		}
	}
	if got := lastResult(t, settingsRepository, settingsEntity.JobKeyInvitationCleanup)["expired"]; got != 1 {
		t.Errorf("expected recorded expired count 1, got %d", got)
	}
}

type cleanupFixture struct {
	db                    *bun.DB
	userRepository        userRepo.IRepository
	userSessionRepository userSessionRepo.IRepository
	roleRepository        roleRepo.IRepository
	settingsRepository    settingsRepo.IRepository
}

// seedUser creates a user aged by age, optionally verified, with one active
// session and one role membership.
func (f cleanupFixture) seedUser(t *testing.T, email string, age time.Duration, verified bool) string {
	t.Helper()
	ctx := context.Background()
	userID, err := f.userRepository.Create(ctx, userModels.CreateRequest{Name: email, Email: email})
	if err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	if verified {
		if err := f.userRepository.Verify(ctx, userID); err != nil {
			t.Fatalf("verify %s: %v", email, err)
		}
	}
	if _, err := f.db.NewUpdate().TableExpr("users").Set("created_at = ?", time.Now().UTC().Add(-age)).Where("id = ?", userID).Exec(ctx); err != nil {
		t.Fatalf("age user %s: %v", email, err)
	}
	if _, err := f.userSessionRepository.Create(ctx, userSessionModels.CreateRequest{
		UserID: userID, TokenID: "token-" + email, Email: email, RefreshToken: "refresh-" + email,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}); err != nil {
		t.Fatalf("create session %s: %v", email, err)
	}
	appServiceID := "app_isme"
	if err := f.roleRepository.AddMembers(ctx, "rol_member", []string{userID}, &appServiceID); err != nil {
		t.Fatalf("add member %s: %v", email, err)
	}
	return userID
}

func (f cleanupFixture) state(t *testing.T, userID string) (live bool, activeSessions int64, memberships int64) {
	t.Helper()
	ctx := context.Background()
	liveCount, err := f.db.NewSelect().TableExpr("users").Where("id = ?", userID).Where("deleted_at IS NULL").Count(ctx)
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	activeSessions, err = f.db.NewSelect().TableExpr("user_sessions").Where("user_id = ?", userID).
		Where("status = ?", userSessionConstants.UserSessionStatusActive).Count(ctx)
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	memberships, err = f.db.NewSelect().TableExpr("user_roles").Where("user_id = ?", userID).Count(ctx)
	if err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	return liveCount == 1, activeSessions, memberships
}

func newCleanupFixture(t *testing.T) cleanupFixture {
	db := newFileTestDB(t)
	return cleanupFixture{
		db:                    db,
		userRepository:        userRepo.NewRepository(db),
		userSessionRepository: userSessionRepo.NewRepository(db),
		roleRepository:        roleRepo.NewRepository(db),
		settingsRepository:    settingsRepo.NewRepository(db),
	}
}

func (f cleanupFixture) run(t *testing.T) {
	t.Helper()
	run := newUnverifiedUserCleanupRun(f.userRepository, f.userSessionRepository, f.roleRepository, pkgTx.NewRunner(f.db), f.settingsRepository)
	if err := run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestUnverifiedUserCleanupRunDeletesOnlyOldUnverified(t *testing.T) {
	fixture := newCleanupFixture(t)
	day := 24 * time.Hour

	oldUnverified := fixture.seedUser(t, "old-unverified@hasaki.vn", 40*day, false)
	youngUnverified := fixture.seedUser(t, "young-unverified@hasaki.vn", 2*day, false)
	oldVerified := fixture.seedUser(t, "old-verified@hasaki.vn", 40*day, true)

	fixture.run(t)

	if live, sessions, memberships := fixture.state(t, oldUnverified); live || sessions != 0 || memberships != 0 {
		t.Errorf("old unverified: expected deleted with no sessions/memberships, got live=%v sessions=%d memberships=%d", live, sessions, memberships)
	}
	for name, userID := range map[string]string{"young unverified": youngUnverified, "old verified": oldVerified} {
		if live, sessions, memberships := fixture.state(t, userID); !live || sessions != 1 || memberships != 1 {
			t.Errorf("%s: expected untouched, got live=%v sessions=%d memberships=%d", name, live, sessions, memberships)
		}
	}

	var deletedBy string
	if err := fixture.db.NewSelect().TableExpr("users").Column("deleted_by").Where("id = ?", oldUnverified).Scan(context.Background(), &deletedBy); err != nil {
		t.Fatalf("read deleted_by: %v", err)
	}
	if deletedBy != unverifiedUserCleanupActor {
		t.Errorf("expected deleted_by %q, got %q", unverifiedUserCleanupActor, deletedBy)
	}
	if got := lastResult(t, fixture.settingsRepository, settingsEntity.JobKeyUnverifiedUserCleanup)["deleted"]; got != 1 {
		t.Errorf("expected recorded deleted count 1, got %d", got)
	}

	// a second run finds nothing left to delete
	fixture.run(t)
	if got := lastResult(t, fixture.settingsRepository, settingsEntity.JobKeyUnverifiedUserCleanup)["deleted"]; got != 0 {
		t.Errorf("expected recorded deleted count 0 on the second run, got %d", got)
	}
}

func TestUnverifiedUserCleanupRunSkipsBelowFloor(t *testing.T) {
	fixture := newCleanupFixture(t)
	oldUnverified := fixture.seedUser(t, "old-unverified@hasaki.vn", 40*24*time.Hour, false)

	if _, err := fixture.db.NewUpdate().TableExpr("schedule_config").Set("params = ?", `{"min_age_days":0}`).
		Where("job_key = ?", settingsEntity.JobKeyUnverifiedUserCleanup).Exec(context.Background()); err != nil {
		t.Fatalf("set params: %v", err)
	}

	fixture.run(t)

	if live, _, _ := fixture.state(t, oldUnverified); !live {
		t.Error("expected a below-floor config to skip the run and delete nothing")
	}
}
