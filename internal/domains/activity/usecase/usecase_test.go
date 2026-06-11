package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/domains/activity/constants"
	"github.com/vukyn/isme/internal/domains/activity/entity"
)

// fakeRepository captures created events and can be made to fail, so tests can
// assert both the emitted type/meta and the best-effort no-propagation contract.
type fakeRepository struct {
	created   []entity.ActivityEvent
	createErr error
}

func (f *fakeRepository) Create(ctx context.Context, event entity.ActivityEvent) error {
	f.created = append(f.created, event)
	return f.createErr
}

func (f *fakeRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]entity.ActivityEvent, error) {
	return nil, nil
}

func (f *fakeRepository) PruneBefore(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func decodeMeta(t *testing.T, raw string) map[string]any {
	t.Helper()
	meta := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		t.Fatalf("unmarshal meta %q: %v", raw, err)
	}
	return meta
}

func TestRecordSignInBuildsTypeAndMeta(t *testing.T) {
	repo := &fakeRepository{}
	uc := NewUsecase(repo)

	uc.RecordSignIn(context.Background(), "user-1", "Chrome on macOS", "127.0.0.1")

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 created event, got %d", len(repo.created))
	}
	event := repo.created[0]
	if event.Type != constants.ActivityTypeSignIn {
		t.Errorf("expected type %q, got %q", constants.ActivityTypeSignIn, event.Type)
	}
	if event.UserID != "user-1" {
		t.Errorf("expected user-1, got %q", event.UserID)
	}
	meta := decodeMeta(t, event.Meta)
	if meta["device"] != "Chrome on macOS" {
		t.Errorf("expected device in meta, got %v", meta["device"])
	}
	if meta["client_ip"] != "127.0.0.1" {
		t.Errorf("expected client_ip in meta, got %v", meta["client_ip"])
	}
}

func TestRecordSignOutBuildsTypeAndEmptyMeta(t *testing.T) {
	repo := &fakeRepository{}
	uc := NewUsecase(repo)

	uc.RecordSignOut(context.Background(), "user-1")

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 created event, got %d", len(repo.created))
	}
	if repo.created[0].Type != constants.ActivityTypeSignOut {
		t.Errorf("expected type %q, got %q", constants.ActivityTypeSignOut, repo.created[0].Type)
	}
	if len(decodeMeta(t, repo.created[0].Meta)) != 0 {
		t.Errorf("expected empty meta, got %q", repo.created[0].Meta)
	}
}

func TestRecordPasswordChangedBuildsType(t *testing.T) {
	repo := &fakeRepository{}
	uc := NewUsecase(repo)

	uc.RecordPasswordChanged(context.Background(), "user-1")

	if len(repo.created) != 1 || repo.created[0].Type != constants.ActivityTypePasswordChanged {
		t.Fatalf("expected one password_changed event, got %+v", repo.created)
	}
}

func TestRecordInvitationSentBuildsTypeAndMeta(t *testing.T) {
	repo := &fakeRepository{}
	uc := NewUsecase(repo)

	uc.RecordInvitationSent(context.Background(), "inviter-1", "new@example.com", []string{"Member", "Editor"})

	if len(repo.created) != 1 {
		t.Fatalf("expected 1 created event, got %d", len(repo.created))
	}
	event := repo.created[0]
	if event.Type != constants.ActivityTypeInvitationSent {
		t.Errorf("expected type %q, got %q", constants.ActivityTypeInvitationSent, event.Type)
	}
	if event.UserID != "inviter-1" {
		t.Errorf("expected inviter-1, got %q", event.UserID)
	}
	meta := decodeMeta(t, event.Meta)
	if meta["email"] != "new@example.com" {
		t.Errorf("expected email in meta, got %v", meta["email"])
	}
	roles, ok := meta["roles"].([]any)
	if !ok || len(roles) != 2 {
		t.Errorf("expected 2 roles in meta, got %v", meta["roles"])
	}
}

// TestRecordSwallowsRepoError proves a repository failure never propagates — the
// Record* methods return nothing and the audited action is unaffected.
func TestRecordSwallowsRepoError(t *testing.T) {
	repo := &fakeRepository{createErr: errors.New("database unavailable")}
	uc := NewUsecase(repo)

	// none of these panic or propagate; they return void.
	uc.RecordSignIn(context.Background(), "user-1", "device", "ip")
	uc.RecordSignOut(context.Background(), "user-1")
	uc.RecordPasswordChanged(context.Background(), "user-1")
	uc.RecordInvitationSent(context.Background(), "inviter-1", "e@example.com", []string{"Member"})

	if len(repo.created) != 4 {
		t.Errorf("expected 4 Create attempts despite errors, got %d", len(repo.created))
	}
}

func TestListMapsEventsToItems(t *testing.T) {
	repo := &listRepo{events: []entity.ActivityEvent{
		{ID: "id-1", UserID: "user-1", Type: constants.ActivityTypeSignIn, Meta: `{"device":"Chrome","client_ip":"127.0.0.1"}`},
		{ID: "id-2", UserID: "user-1", Type: constants.ActivityTypeSignOut, Meta: "{}"},
	}}
	uc := NewUsecase(repo)

	items, err := uc.List(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Type != constants.ActivityTypeSignIn {
		t.Errorf("expected first item sign_in, got %q", items[0].Type)
	}
	if items[0].Meta["device"] != "Chrome" {
		t.Errorf("expected device parsed into meta map, got %v", items[0].Meta["device"])
	}
}

// TestListEmptyNonNil proves an empty feed maps to a non-nil slice.
func TestListEmptyNonNil(t *testing.T) {
	uc := NewUsecase(&listRepo{events: []entity.ActivityEvent{}})

	items, err := uc.List(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil empty slice")
	}
}

// listRepo returns a fixed set of events for List tests.
type listRepo struct {
	events []entity.ActivityEvent
}

func (l *listRepo) Create(ctx context.Context, event entity.ActivityEvent) error { return nil }

func (l *listRepo) ListByUserID(ctx context.Context, userID string, limit int) ([]entity.ActivityEvent, error) {
	return l.events, nil
}

func (l *listRepo) PruneBefore(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}
