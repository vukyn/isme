package usecase

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	"github.com/vukyn/isme/internal/domains/user_invitation/constants"
	"github.com/vukyn/isme/internal/domains/user_invitation/entity"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"
	invitationRepo "github.com/vukyn/isme/internal/domains/user_invitation/repository"

	"github.com/uptrace/bun"
	kueryDb "github.com/vukyn/kuery/bun/db"
	pkgTx "github.com/vukyn/kuery/bun/tx"
	"github.com/vukyn/kuery/cryp"
)

// These tests run Accept against a real file-backed SQLite database opened the
// way the service opens it, so a repository write that escapes the
// transaction (by not reading its handle through pkgTx.Conn) is caught: it
// either commits on its own or blocks on the writer lock and fails.

var errInjected = errors.New("injected failure")

// newFileTestDB opens a file SQLite database under t.TempDir() with the
// service's pool settings and applies every migration.
func newFileTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := kueryDb.Open(kueryDb.Config{
		Driver:     kueryDb.DriverSQLite,
		SQLitePath: filepath.Join(t.TempDir(), "accept.db"),
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

// failingUserRepository fails the selected write after the real ones before it ran.
type failingUserRepository struct {
	userRepo.IRepository
	failSetPassword bool
	failVerify      bool
}

func (f *failingUserRepository) SetPassword(ctx context.Context, id string, password string) error {
	if f.failSetPassword {
		return errInjected
	}
	return f.IRepository.SetPassword(ctx, id, password)
}

func (f *failingUserRepository) Verify(ctx context.Context, id string) error {
	if f.failVerify {
		return errInjected
	}
	return f.IRepository.Verify(ctx, id)
}

type failingRoleRepository struct {
	roleRepo.IRepository
	failAddMembers bool
}

func (f *failingRoleRepository) AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error {
	if f.failAddMembers {
		return errInjected
	}
	return f.IRepository.AddMembers(ctx, roleID, userIDs, appServiceID)
}

func newRealAcceptUsecase(db *bun.DB, users userRepo.IRepository, roles roleRepo.IRepository) IUseCase {
	return NewUsecase(
		newTestConfig(),
		invitationRepo.NewRepository(db),
		users,
		roles,
		appServiceRepo.NewRepository(db),
		nil,
		pkgTx.NewRunner(db),
	)
}

// seedRealInvitation stores a pending invitation with two role assignments and
// returns its id and raw token.
func seedRealInvitation(t *testing.T, db *bun.DB, email string) (string, string) {
	t.Helper()
	rawToken := "raw-token-" + email
	invitationID, err := invitationRepo.NewRepository(db).Create(context.Background(), entity.UserInvitation{
		Email:     email,
		TokenHash: cryp.HashSHA256(rawToken),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}, []entity.UserInvitationRole{
		{RoleID: "rol_member", AppServiceID: "app_isme"},
		{RoleID: "rol_viewer", AppServiceID: "app_rainy"},
	})
	if err != nil {
		t.Fatalf("seed invitation: %v", err)
	}
	return invitationID, rawToken
}

// countUsersByEmail counts every row with the email, soft-deleted ones included.
func countUsersByEmail(t *testing.T, db *bun.DB, email string) int {
	t.Helper()
	count, err := db.NewSelect().TableExpr("users").Where("email = ?", email).Count(context.Background())
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	return count
}

func assertInvitationPending(t *testing.T, db *bun.DB, invitationID string) {
	t.Helper()
	var status int32
	var acceptedAt sql.NullString
	err := db.NewSelect().
		TableExpr("user_invitations").
		ColumnExpr("status, accepted_at").
		Where("id = ?", invitationID).
		Scan(context.Background(), &status, &acceptedAt)
	if err != nil {
		t.Fatalf("read invitation: %v", err)
	}
	if status != int32(constants.InvitationStatusPending) {
		t.Errorf("expected invitation status %d after rollback, got %d", constants.InvitationStatusPending, status)
	}
	if acceptedAt.Valid {
		t.Errorf("expected accepted_at NULL after rollback, got %q", acceptedAt.String)
	}
}

func TestAcceptRollsBackOnRealDatabase(t *testing.T) {
	cases := []struct {
		name  string
		users func(base userRepo.IRepository) userRepo.IRepository
		roles func(base roleRepo.IRepository) roleRepo.IRepository
	}{
		{
			name: "set password fails",
			users: func(base userRepo.IRepository) userRepo.IRepository {
				return &failingUserRepository{IRepository: base, failSetPassword: true}
			},
			roles: func(base roleRepo.IRepository) roleRepo.IRepository { return base },
		},
		{
			name: "verify fails",
			users: func(base userRepo.IRepository) userRepo.IRepository {
				return &failingUserRepository{IRepository: base, failVerify: true}
			},
			roles: func(base roleRepo.IRepository) roleRepo.IRepository { return base },
		},
		{
			name:  "second role assignment fails",
			users: func(base userRepo.IRepository) userRepo.IRepository { return base },
			roles: func(base roleRepo.IRepository) roleRepo.IRepository {
				return &secondCallFailingRoleRepository{IRepository: base}
			},
		},
		{
			name:  "first role assignment fails",
			users: func(base userRepo.IRepository) userRepo.IRepository { return base },
			roles: func(base roleRepo.IRepository) roleRepo.IRepository {
				return &failingRoleRepository{IRepository: base, failAddMembers: true}
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			db := newFileTestDB(t)
			ctx := context.Background()
			email := "rollback@hasaki.vn"
			invitationID, rawToken := seedRealInvitation(t, db, email)
			request := models.AcceptRequest{Token: rawToken, Name: "Rollback", Password: fixturePassword}

			failing := newRealAcceptUsecase(db,
				testCase.users(userRepo.NewRepository(db)),
				testCase.roles(roleRepo.NewRepository(db)))
			if err := failing.Accept(ctx, request); !errors.Is(err, errInjected) {
				t.Fatalf("expected injected failure, got %v", err)
			}

			if count := countUsersByEmail(t, db, email); count != 0 {
				t.Errorf("expected no user row after rollback, got %d", count)
			}
			roleCount, err := db.NewSelect().TableExpr("user_roles").Count(ctx)
			if err != nil {
				t.Fatalf("count user_roles: %v", err)
			}
			if roleCount != 0 {
				t.Errorf("expected no user_roles after rollback, got %d", roleCount)
			}
			assertInvitationPending(t, db, invitationID)

			// the link stays usable: a retry with healthy repositories succeeds
			healthy := newRealAcceptUsecase(db, userRepo.NewRepository(db), roleRepo.NewRepository(db))
			if err := healthy.Accept(ctx, request); err != nil {
				t.Fatalf("expected retry to succeed, got %v", err)
			}
			user, err := userRepo.NewRepository(db).GetByEmail(ctx, email)
			if err != nil || user.ID == "" {
				t.Fatalf("expected the user to exist after retry, got %+v err=%v", user, err)
			}
			if !user.IsVerified {
				t.Error("expected the accepted user to be verified")
			}
			memberships, err := db.NewSelect().TableExpr("user_roles").Where("user_id = ?", user.ID).Count(ctx)
			if err != nil {
				t.Fatalf("count memberships: %v", err)
			}
			if memberships != 2 {
				t.Errorf("expected 2 role memberships after retry, got %d", memberships)
			}
		})
	}
}

// secondCallFailingRoleRepository lets the first assignment write, then fails,
// so the rollback must undo a role row that was really inserted.
type secondCallFailingRoleRepository struct {
	roleRepo.IRepository
	calls int
}

func (f *secondCallFailingRoleRepository) AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error {
	f.calls++
	if f.calls > 1 {
		return errInjected
	}
	return f.IRepository.AddMembers(ctx, roleID, userIDs, appServiceID)
}
