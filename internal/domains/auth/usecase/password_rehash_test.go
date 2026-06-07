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

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/domains/auth/models"
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

func (f *fakeUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) PromoteAdmin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) IsAdmin(ctx context.Context, id string) (bool, error) {
	return false, nil
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

	cfg := &config.Config{}
	cfg.Auth.AccessTokenPrivateKey = string(privateKeyPEM)
	cfg.Auth.AccessTokenExpireIn = 3600
	cfg.Auth.RefreshTokenSecretKey = "test-refresh-secret"
	cfg.Auth.RefreshTokenExpireIn = 7200
	return cfg
}

func newTestUsecase(t *testing.T, userRepository *fakeUserRepository) IUseCase {
	t.Helper()
	return NewUsecase(newTestConfig(t), nil, userRepository, &fakeUserSessionRepository{}, nil)
}

func TestLoginRehashesBcryptPassword(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:       "user-1",
			Email:    "user@example.com",
			Password: cryp.HashBcrypt("s3cret-password", 4),
			Status:   userConstants.UserStatusActive,
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
			ID:       "user-1",
			Email:    "user@example.com",
			Password: cryp.HashArgon2id("s3cret-password"),
			Status:   userConstants.UserStatusActive,
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
			ID:       "user-1",
			Email:    "user@example.com",
			Password: cryp.HashBcrypt("s3cret-password", 4),
			Status:   userConstants.UserStatusActive,
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
			ID:       "user-1",
			Email:    "user@example.com",
			Password: cryp.HashBcrypt("s3cret-password", 4),
			Status:   userConstants.UserStatusActive,
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
