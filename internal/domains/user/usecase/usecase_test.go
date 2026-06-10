package usecase

import (
	"context"
	"slices"
	"testing"
	"time"

	roleEntity "github.com/vukyn/isme/internal/domains/role/entity"
	roleModels "github.com/vukyn/isme/internal/domains/role/models"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	"github.com/vukyn/isme/internal/domains/user/entity"
	"github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	userSessionEntity "github.com/vukyn/isme/internal/domains/user_session/entity"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	pkgCtx "github.com/vukyn/kuery/ctx"
)

// === Fakes ===

type fakeUserRepository struct {
	usersByID       map[string]entity.User
	updatedStatuses map[string]int32
	softDeletedIDs  []string
	verifiedIDs     []string
}

var _ userRepo.IRepository = (*fakeUserRepository)(nil)

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		usersByID:       map[string]entity.User{},
		updatedStatuses: map[string]int32{},
	}
}

func (f *fakeUserRepository) Create(ctx context.Context, req models.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id string) (entity.User, error) {
	return f.usersByID[id], nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	return entity.User{}, nil
}

func (f *fakeUserRepository) SetPassword(ctx context.Context, id string, password string) error {
	return nil
}

func (f *fakeUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) Verify(ctx context.Context, id string) error {
	f.verifiedIDs = append(f.verifiedIDs, id)
	return nil
}

func (f *fakeUserRepository) List(ctx context.Context, req models.ListRequest) ([]entity.User, int64, error) {
	return nil, 0, nil
}

func (f *fakeUserRepository) UpdateStatus(ctx context.Context, id string, status int32) error {
	f.updatedStatuses[id] = status
	return nil
}

func (f *fakeUserRepository) SoftDelete(ctx context.Context, id string) error {
	f.softDeletedIDs = append(f.softDeletedIDs, id)
	return nil
}

type fakeUserSessionRepository struct {
	sessionsByID        map[string]userSessionEntity.UserSession
	inactivatedSessions []string
	inactivatedUserAlls []string
}

var _ userSessionRepo.IRepository = (*fakeUserSessionRepository)(nil)

func newFakeUserSessionRepository() *fakeUserSessionRepository {
	return &fakeUserSessionRepository{
		sessionsByID: map[string]userSessionEntity.UserSession{},
	}
}

func (f *fakeUserSessionRepository) Create(ctx context.Context, req userSessionModels.CreateRequest) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{}, nil
}

func (f *fakeUserSessionRepository) UpdateLastLogin(ctx context.Context, req userSessionModels.UpdateLastLoginRequest) error {
	return nil
}

func (f *fakeUserSessionRepository) InactiveAllUserSession(ctx context.Context, userID string) error {
	f.inactivatedUserAlls = append(f.inactivatedUserAlls, userID)
	return nil
}

func (f *fakeUserSessionRepository) InactiveSessionByTokenID(ctx context.Context, tokenID string) error {
	return nil
}

func (f *fakeUserSessionRepository) InactiveSessionByID(ctx context.Context, sessionID string) error {
	f.inactivatedSessions = append(f.inactivatedSessions, sessionID)
	return nil
}

func (f *fakeUserSessionRepository) InactiveAllUserSessionExcept(ctx context.Context, userID string, exceptTokenID string) error {
	return nil
}

func (f *fakeUserSessionRepository) CountActiveByUserIDCreatedAfter(ctx context.Context, userID string, after time.Time) (int, error) {
	return 0, nil
}

func (f *fakeUserSessionRepository) InactiveExpiredSessions(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeUserSessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{}, nil
}

func (f *fakeUserSessionRepository) FindByTokenID(ctx context.Context, tokenID string) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{}, nil
}

func (f *fakeUserSessionRepository) GetByID(ctx context.Context, sessionID string) (userSessionEntity.UserSession, error) {
	return f.sessionsByID[sessionID], nil
}

func (f *fakeUserSessionRepository) GetListActiveByUserID(ctx context.Context, userID string) ([]userSessionEntity.UserSession, error) {
	return nil, nil
}

func (f *fakeUserSessionRepository) CountActiveByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error) {
	return map[string]int{}, nil
}

type fakeRoleRepository struct{}

var _ roleRepo.IRepository = (*fakeRoleRepository)(nil)

func (f *fakeRoleRepository) Create(ctx context.Context, req roleModels.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeRoleRepository) GetByID(ctx context.Context, id string) (roleEntity.Role, error) {
	return roleEntity.Role{}, nil
}

func (f *fakeRoleRepository) GetByAppAndCode(ctx context.Context, appID string, code string) (roleEntity.Role, error) {
	return roleEntity.Role{}, nil
}

func (f *fakeRoleRepository) List(ctx context.Context, req roleModels.ListRequest) ([]roleModels.RoleListItem, error) {
	return nil, nil
}

func (f *fakeRoleRepository) Update(ctx context.Context, id string, req roleModels.UpdateRequest) error {
	return nil
}

func (f *fakeRoleRepository) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRoleRepository) ListPermissions(ctx context.Context, req roleModels.ListPermissionsRequest) ([]roleEntity.Permission, error) {
	return nil, nil
}

func (f *fakeRoleRepository) CreatePermissions(ctx context.Context, appID string, perms []roleModels.PermissionItem) (map[string]int64, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionByID(ctx context.Context, permissionID int64) (roleEntity.Permission, error) {
	return roleEntity.Permission{}, nil
}

func (f *fakeRoleRepository) DeletePermission(ctx context.Context, permissionID int64) error {
	return nil
}

func (f *fakeRoleRepository) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]roleEntity.Permission, error) {
	return nil, nil
}

func (f *fakeRoleRepository) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []int64) error {
	return nil
}

func (f *fakeRoleRepository) ListMembers(ctx context.Context, roleID string, req roleModels.ListMembersRequest) ([]roleModels.MemberItem, int, error) {
	return nil, 0, nil
}

func (f *fakeRoleRepository) CountMembersByRoleID(ctx context.Context, roleID string) (int, error) {
	return 0, nil
}

func (f *fakeRoleRepository) AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error {
	return nil
}

func (f *fakeRoleRepository) RemoveMember(ctx context.Context, roleID string, userID string, appServiceID *string) error {
	return nil
}

func (f *fakeRoleRepository) GetPermissionCodesByUserID(ctx context.Context, userID string, appID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionCodesGroupedByApp(ctx context.Context, userID string) (map[string][]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetAppCodesByUserID(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionCodesByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	return map[string][]string{}, nil
}

func (f *fakeRoleRepository) GetRoleCodesGroupedByAppByUserIDs(ctx context.Context, userIDs []string) (map[string][]roleModels.UserAppRole, error) {
	return map[string][]roleModels.UserAppRole{}, nil
}

// === Tests ===

func TestRevokeSession(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		sessionID   string
		wantErr     string
		wantRevoked bool
	}{
		{
			name:        "owned session revoked",
			userID:      "user-a",
			sessionID:   "session-1",
			wantRevoked: true,
		},
		{
			name:      "session of another user rejected",
			userID:    "user-b",
			sessionID: "session-1",
			wantErr:   "session not found",
		},
		{
			name:      "missing session rejected",
			userID:    "user-a",
			sessionID: "session-unknown",
			wantErr:   "session not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := newFakeUserRepository()
			fakeUserSession := newFakeUserSessionRepository()
			fakeUserSession.sessionsByID["session-1"] = userSessionEntity.UserSession{ID: "session-1", UserID: "user-a"}
			testUsecase := NewUsecase(fakeUser, fakeUserSession, &fakeRoleRepository{})

			err := testUsecase.RevokeSession(context.Background(), tt.userID, tt.sessionID)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				if len(fakeUserSession.inactivatedSessions) != 0 {
					t.Error("session was revoked despite rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("RevokeSession() error = %v", err)
			}
			if tt.wantRevoked && !slices.Contains(fakeUserSession.inactivatedSessions, tt.sessionID) {
				t.Errorf("session %q was not revoked", tt.sessionID)
			}
		})
	}
}

func TestSoftDelete(t *testing.T) {
	tests := []struct {
		name         string
		callerUserID string
		targetUserID string
		wantErr      string
	}{
		{
			name:         "self-delete rejected",
			callerUserID: "user-a",
			targetUserID: "user-a",
			wantErr:      "cannot delete your own account",
		},
		{
			name:         "missing user rejected",
			callerUserID: "user-a",
			targetUserID: "user-unknown",
			wantErr:      "user not found",
		},
		{
			name:         "delete another user succeeds",
			callerUserID: "user-a",
			targetUserID: "user-b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := newFakeUserRepository()
			fakeUser.usersByID["user-a"] = entity.User{ID: "user-a"}
			fakeUser.usersByID["user-b"] = entity.User{ID: "user-b"}
			fakeUserSession := newFakeUserSessionRepository()
			testUsecase := NewUsecase(fakeUser, fakeUserSession, &fakeRoleRepository{})

			ctx := context.WithValue(context.Background(), pkgCtx.UserIDKey, tt.callerUserID)
			err := testUsecase.SoftDelete(ctx, tt.targetUserID)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				if len(fakeUser.softDeletedIDs) != 0 {
					t.Error("user was deleted despite rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("SoftDelete() error = %v", err)
			}
			if !slices.Contains(fakeUser.softDeletedIDs, tt.targetUserID) {
				t.Errorf("user %q was not deleted", tt.targetUserID)
			}
			if !slices.Contains(fakeUserSession.inactivatedUserAlls, tt.targetUserID) {
				t.Errorf("sessions of user %q were not revoked", tt.targetUserID)
			}
		})
	}
}

func TestVerifyUser(t *testing.T) {
	tests := []struct {
		name         string
		targetUserID string
		wantErr      string
	}{
		{"unverified user verified", "user-unverified", ""},
		{"missing user rejected", "user-unknown", "user not found"},
		{"already verified rejected", "user-verified", "user already verified"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := newFakeUserRepository()
			fakeUser.usersByID["user-unverified"] = entity.User{ID: "user-unverified"}
			fakeUser.usersByID["user-verified"] = entity.User{ID: "user-verified", IsVerified: true}
			testUsecase := NewUsecase(fakeUser, newFakeUserSessionRepository(), &fakeRoleRepository{})

			err := testUsecase.VerifyUser(context.Background(), tt.targetUserID)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				if len(fakeUser.verifiedIDs) != 0 {
					t.Error("user was verified despite rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("VerifyUser() error = %v", err)
			}
			if !slices.Contains(fakeUser.verifiedIDs, tt.targetUserID) {
				t.Errorf("user %q was not verified", tt.targetUserID)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name         string
		targetUserID string
		status       int32
		wantErr      string
	}{
		{"activate succeeds", "user-a", 1, ""},
		{"deactivate succeeds", "user-a", 2, ""},
		{"status zero rejected", "user-a", 0, "invalid status, must be 1 (active) or 2 (inactive)"},
		{"status out of range rejected", "user-a", 3, "invalid status, must be 1 (active) or 2 (inactive)"},
		{"missing user rejected", "user-unknown", 1, "user not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeUser := newFakeUserRepository()
			fakeUser.usersByID["user-a"] = entity.User{ID: "user-a"}
			testUsecase := NewUsecase(fakeUser, newFakeUserSessionRepository(), &fakeRoleRepository{})

			err := testUsecase.UpdateStatus(context.Background(), tt.targetUserID, models.UpdateStatusRequest{Status: tt.status})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				if len(fakeUser.updatedStatuses) != 0 {
					t.Error("status was updated despite rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateStatus() error = %v", err)
			}
			if got := fakeUser.updatedStatuses[tt.targetUserID]; got != tt.status {
				t.Errorf("updated status = %d, want %d", got, tt.status)
			}
		})
	}
}
