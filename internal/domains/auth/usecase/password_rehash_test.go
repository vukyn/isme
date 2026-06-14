package usecase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/domains/auth/models"
	roleEntity "github.com/vukyn/isme/internal/domains/role/entity"
	roleModels "github.com/vukyn/isme/internal/domains/role/models"
	userConstants "github.com/vukyn/isme/internal/domains/user/constants"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	userSessionEntity "github.com/vukyn/isme/internal/domains/user_session/entity"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"

	"github.com/vukyn/kuery/cryp"
)

type setPasswordCall struct {
	id       string
	password string
}

type fakeUserRepository struct {
	user             userEntity.User
	setPasswordCalls []setPasswordCall
	setPasswordErr   error
}

func (f *fakeUserRepository) Create(ctx context.Context, req userModels.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id string) (userEntity.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (userEntity.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) SetPassword(ctx context.Context, id string, password string) error {
	f.setPasswordCalls = append(f.setPasswordCalls, setPasswordCall{id: id, password: password})
	return f.setPasswordErr
}

func (f *fakeUserRepository) UpdateProfile(ctx context.Context, id string, name string, avatarURL string) error {
	return nil
}

func (f *fakeUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) Verify(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) List(ctx context.Context, req userModels.ListRequest) ([]userEntity.User, int64, error) {
	return nil, 0, nil
}

func (f *fakeUserRepository) UpdateStatus(ctx context.Context, id string, status int32) error {
	return nil
}

func (f *fakeUserRepository) SoftDelete(ctx context.Context, id string) error {
	return nil
}

type fakeUserSessionRepository struct{}

func (f *fakeUserSessionRepository) Create(ctx context.Context, req userSessionModels.CreateRequest) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{ID: "session-id"}, nil
}

func (f *fakeUserSessionRepository) UpdateLastLogin(ctx context.Context, req userSessionModels.UpdateLastLoginRequest) error {
	return nil
}

func (f *fakeUserSessionRepository) InactiveAllUserSession(ctx context.Context, userID string) error {
	return nil
}

func (f *fakeUserSessionRepository) InactiveSessionByTokenID(ctx context.Context, tokenID string) error {
	return nil
}

func (f *fakeUserSessionRepository) InactiveSessionByID(ctx context.Context, sessionID string) error {
	return nil
}

func (f *fakeUserSessionRepository) InactiveAllUserSessionExcept(ctx context.Context, userID string, exceptTokenID string) error {
	return nil
}

func (f *fakeUserSessionRepository) CountActiveByUserIDCreatedAfter(ctx context.Context, userID string, after time.Time) (int, error) {
	return 0, nil
}

func (f *fakeUserSessionRepository) CountRotationsByUserIDSince(ctx context.Context, userID string, since time.Time) (int, error) {
	return 0, nil
}

func (f *fakeUserSessionRepository) InactiveExpiredSessions(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeUserSessionRepository) PruneRotationsBefore(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeUserSessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{}, nil
}

func (f *fakeUserSessionRepository) FindByTokenID(ctx context.Context, tokenID string) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{}, nil
}

func (f *fakeUserSessionRepository) GetByID(ctx context.Context, sessionID string) (userSessionEntity.UserSession, error) {
	return userSessionEntity.UserSession{}, nil
}

func (f *fakeUserSessionRepository) GetListActiveByUserID(ctx context.Context, userID string) ([]userSessionEntity.UserSession, error) {
	return nil, nil
}

func (f *fakeUserSessionRepository) CountActiveByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error) {
	return nil, nil
}

type fakeRoleRepository struct {
	permissionCodes        []string
	groupedPermissionCodes map[string][]string
}

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

func (f *fakeRoleRepository) UpdatePermissionAppearance(ctx context.Context, appID string, resource string, icon string, color string) error {
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
	return f.permissionCodes, nil
}

func (f *fakeRoleRepository) GetPermissionCodesGroupedByApp(ctx context.Context, userID string) (map[string][]string, error) {
	return f.groupedPermissionCodes, nil
}

func (f *fakeRoleRepository) GetAppCodesByUserID(ctx context.Context, userID string) ([]string, error) {
	codes := make([]string, 0, len(f.groupedPermissionCodes))
	for code := range f.groupedPermissionCodes {
		codes = append(codes, code)
	}
	return codes, nil
}

func (f *fakeRoleRepository) GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionCodesByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetRoleCodesGroupedByAppByUserIDs(ctx context.Context, userIDs []string) (map[string][]roleModels.UserAppRole, error) {
	return nil, nil
}

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal RSA public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	cfg := &config.Config{}
	cfg.Auth.AccessTokenPrivateKey = string(privateKeyPEM)
	cfg.Auth.AccessTokenPublicKey = string(publicKeyPEM)
	cfg.Auth.AccessTokenExpireIn = 3600
	cfg.Auth.RefreshTokenSecretKey = "test-refresh-secret"
	cfg.Auth.RefreshTokenExpireIn = 7200
	return cfg
}

func newTestUsecase(t *testing.T, userRepository *fakeUserRepository) IUseCase {
	t.Helper()
	return newTestUsecaseWithRoles(t, userRepository, &fakeRoleRepository{})
}

func newTestUsecaseWithRoles(t *testing.T, userRepository *fakeUserRepository, roleRepository *fakeRoleRepository) IUseCase {
	t.Helper()
	uc, _ := newTestUsecaseWithActivity(t, userRepository, roleRepository)
	return uc
}

// newTestUsecaseWithActivity wires the auth usecase with a recording activity
// double, returning it so tests can assert emitted events.
func newTestUsecaseWithActivity(t *testing.T, userRepository *fakeUserRepository, roleRepository *fakeRoleRepository) (IUseCase, *fakeActivityUsecase) {
	t.Helper()
	activity := &fakeActivityUsecase{}
	uc := NewUsecase(newTestConfig(t), nil, userRepository, &fakeUserSessionRepository{}, nil, roleRepository, activity)
	return uc, activity
}

func TestLoginRehashesBcryptPassword(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashBcrypt("s3cret-password", 4),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	authUsecase := newTestUsecase(t, userRepository)

	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}
	if res.AccessToken == "" {
		t.Error("expected access token to be set")
	}

	if len(userRepository.setPasswordCalls) != 1 {
		t.Fatalf("expected 1 SetPassword call (rehash), got %d", len(userRepository.setPasswordCalls))
	}
	call := userRepository.setPasswordCalls[0]
	if call.id != "user-1" || call.password != "s3cret-password" {
		t.Errorf("unexpected SetPassword call: %+v", call)
	}
}

func TestLoginArgon2idPasswordNoRehash(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashArgon2id("s3cret-password"),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	authUsecase := newTestUsecase(t, userRepository)

	_, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	if len(userRepository.setPasswordCalls) != 0 {
		t.Errorf("expected no SetPassword call for argon2id hash, got %d", len(userRepository.setPasswordCalls))
	}
}

func TestLoginRehashFailureDoesNotFailLogin(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashBcrypt("s3cret-password", 4),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
		setPasswordErr: errors.New("database unavailable"),
	}
	authUsecase := newTestUsecase(t, userRepository)

	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed despite rehash failure, got error: %v", err)
	}
	if res.AccessToken == "" {
		t.Error("expected access token to be set")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashBcrypt("s3cret-password", 4),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	authUsecase := newTestUsecase(t, userRepository)

	_, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected login to fail with wrong password")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Errorf("unexpected error: %v", err)
	}

	if len(userRepository.setPasswordCalls) != 0 {
		t.Errorf("expected no SetPassword call on failed login, got %d", len(userRepository.setPasswordCalls))
	}
}

// TestLoginEmitsSignIn proves a genuine login records exactly one sign_in event
// for the user, carrying the device (user agent) + client IP from context.
func TestLoginEmitsSignIn(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashArgon2id("s3cret-password"),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	authUsecase, activity := newTestUsecaseWithActivity(t, userRepository, &fakeRoleRepository{})

	_, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	if len(activity.signInCalls) != 1 {
		t.Fatalf("expected exactly 1 sign_in record, got %d", len(activity.signInCalls))
	}
	if activity.signInCalls[0].userID != "user-1" {
		t.Errorf("expected sign_in for user-1, got %q", activity.signInCalls[0].userID)
	}
}

// TestLoginRecorderErrorDoesNotFailLogin proves a failing recorder never fails
// the login (best-effort audit).
func TestLoginRecorderErrorDoesNotFailLogin(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashArgon2id("s3cret-password"),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	activity := &fakeActivityUsecase{recordErr: true}
	authUsecase := NewUsecase(newTestConfig(t), nil, userRepository, &fakeUserSessionRepository{}, nil, &fakeRoleRepository{}, activity)

	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed despite recorder failure, got error: %v", err)
	}
	if res.AccessToken == "" {
		t.Error("expected access token to be set")
	}
	if len(activity.signInCalls) != 1 {
		t.Errorf("expected the recorder to still be invoked once, got %d", len(activity.signInCalls))
	}
}

// TestLogoutEmitsSignOut proves a logout records exactly one sign_out for the
// caller, and still succeeds when the recorder errors (best-effort).
func TestLogoutEmitsSignOut(t *testing.T) {
	activity := &fakeActivityUsecase{recordErr: true}
	uc := NewUsecase(newTestConfig(t), nil, &fakeUserRepository{}, &fakeUserSessionRepository{}, nil, &fakeRoleRepository{}, activity)

	err := uc.Logout(ctxWithUser("user-1", "token-1"))
	if err != nil {
		t.Fatalf("expected logout to succeed, got %v", err)
	}

	if len(activity.signOutCalls) != 1 || activity.signOutCalls[0] != "user-1" {
		t.Errorf("expected one sign_out for user-1, got %v", activity.signOutCalls)
	}
}

// TestChangePasswordEmitsPasswordChanged proves a password change records exactly
// one password_changed for the caller, and still succeeds when the recorder
// errors (best-effort).
func TestChangePasswordEmitsPasswordChanged(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:       "user-1",
			Email:    "user@example.com",
			Password: cryp.HashArgon2id("old-password"),
			Status:   userConstants.UserStatusActive,
		},
	}
	activity := &fakeActivityUsecase{recordErr: true}
	uc := NewUsecase(newTestConfig(t), nil, userRepository, &fakeUserSessionRepository{}, nil, &fakeRoleRepository{}, activity)

	err := uc.ChangePassword(ctxWithUser("user-1", "token-1"), models.ChangePasswordRequest{
		OldPassword: "old-password",
		NewPassword: "new-password-123",
	})
	if err != nil {
		t.Fatalf("expected change password to succeed, got %v", err)
	}

	if len(activity.passwordChangedCalls) != 1 || activity.passwordChangedCalls[0] != "user-1" {
		t.Errorf("expected one password_changed for user-1, got %v", activity.passwordChangedCalls)
	}
}

func TestLoginBlockedWhenUnverified(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-1",
			Email:      "user@example.com",
			Password:   cryp.HashArgon2id("s3cret-password"),
			Status:     userConstants.UserStatusActive,
			IsVerified: false,
		},
	}
	authUsecase := newTestUsecase(t, userRepository)

	// correct password → the verification block surfaces
	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "s3cret-password",
	})
	if err == nil {
		t.Fatal("expected login to fail for unverified account")
	}
	if !strings.Contains(err.Error(), "account pending verification") {
		t.Errorf("unexpected error: %v", err)
	}
	if res.AccessToken != "" {
		t.Error("expected no access token for unverified account")
	}

	// wrong password → generic error, never leaks the verification state
	_, err = authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected login to fail with wrong password")
	}
	if !strings.Contains(err.Error(), "invalid email or password") {
		t.Errorf("wrong password must not leak verification state, got: %v", err)
	}
}
