package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	"github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"
	"github.com/vukyn/isme/internal/testdb"

	"github.com/uptrace/bun"
	pkgTx "github.com/vukyn/kuery/bun/tx"
)

// These tests run SoftDelete against a real Postgres database with a connection
// pool, so a repository write that escapes the transaction (by not reading its
// handle through pkgTx.Conn) is caught: it commits on its own connection and
// survives the rollback, or blocks on a row lock the transaction holds and the
// test times out.

var errInjected = errors.New("injected failure")

// failingSessionRepository runs the real revoke (when asked) and then fails,
// so the rollback must undo both the user delete and any revoked session.
type failingSessionRepository struct {
	userSessionRepo.IRepository
	writeBeforeFailing bool
}

func (f *failingSessionRepository) InactiveAllUserSession(ctx context.Context, userID string) error {
	if f.writeBeforeFailing {
		if err := f.IRepository.InactiveAllUserSession(ctx, userID); err != nil {
			return err
		}
	}
	return errInjected
}

func countActiveSessions(t *testing.T, db *bun.DB, userID string) int {
	t.Helper()
	count, err := db.NewSelect().
		TableExpr("user_sessions").
		Where("user_id = ?", userID).
		Where("status = ?", userSessionConstants.UserSessionStatusActive).
		Count(context.Background())
	if err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	return int(count)
}

func countLiveUsers(t *testing.T, db *bun.DB, userID string) int {
	t.Helper()
	count, err := db.NewSelect().
		TableExpr("users").
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
		Count(context.Background())
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	return int(count)
}

func TestSoftDeleteRollsBackOnRealDatabase(t *testing.T) {
	for _, writeBeforeFailing := range []bool{false, true} {
		name := "revoke fails before writing"
		if writeBeforeFailing {
			name = "revoke fails after writing"
		}
		t.Run(name, func(t *testing.T) {
			db := testdb.Open(t)
			ctx := context.Background()

			userID, err := userRepo.NewRepository(db).Create(ctx, models.CreateRequest{Name: "Target", Email: "target@hasaki.vn"})
			if err != nil {
				t.Fatalf("create user: %v", err)
			}
			_, err = userSessionRepo.NewRepository(db).Create(ctx, userSessionModels.CreateRequest{
				UserID:       userID,
				TokenID:      "token-target",
				Email:        "target@hasaki.vn",
				RefreshToken: "refresh-target",
				ExpiresAt:    time.Now().UTC().Add(time.Hour),
			})
			if err != nil {
				t.Fatalf("create session: %v", err)
			}

			failing := NewUsecase(
				userRepo.NewRepository(db),
				&failingSessionRepository{IRepository: userSessionRepo.NewRepository(db), writeBeforeFailing: writeBeforeFailing},
				roleRepo.NewRepository(db),
				pkgTx.NewRunner(db),
			)
			if err := failing.SoftDelete(ctx, userID); !errors.Is(err, errInjected) {
				t.Fatalf("expected injected failure, got %v", err)
			}
			if countLiveUsers(t, db, userID) != 1 {
				t.Error("expected the user to stay live after rollback")
			}
			if countActiveSessions(t, db, userID) != 1 {
				t.Error("expected the session to stay active after rollback")
			}

			// a retry with healthy repositories deletes the user and revokes the session
			healthy := NewUsecase(
				userRepo.NewRepository(db),
				userSessionRepo.NewRepository(db),
				roleRepo.NewRepository(db),
				pkgTx.NewRunner(db),
			)
			if err := healthy.SoftDelete(ctx, userID); err != nil {
				t.Fatalf("expected retry to succeed, got %v", err)
			}
			if countLiveUsers(t, db, userID) != 0 {
				t.Error("expected the user to be soft-deleted after retry")
			}
			if countActiveSessions(t, db, userID) != 0 {
				t.Error("expected no active session after retry")
			}
		})
	}
}
