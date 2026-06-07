package usecase

import (
	"context"
	"slices"
	"testing"

	"github.com/vukyn/isme/internal/domains/auth/models"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	userConstants "github.com/vukyn/isme/internal/domains/user/constants"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"

	"github.com/vukyn/kuery/cryp"
	"github.com/vukyn/kuery/jwt"
)

func TestLoginAccessTokenCarriesAdminClaim(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:       "user-admin",
			Email:    "admin@example.com",
			Password: cryp.HashArgon2id("s3cret-password"),
			Status:   userConstants.UserStatusActive,
			IsAdmin:  true,
		},
	}
	cfg := newTestConfig(t)
	authUsecase := NewUsecase(cfg, nil, userRepository, &fakeUserSessionRepository{}, nil, &fakeRoleRepository{})

	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "admin@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	claims, err := jwt.ValidateJWTWithRSAPublicKey(res.AccessToken, cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if !claims.GetIsAdmin() {
		t.Error("expected adm claim to be true for an admin user")
	}
}

func TestLoginAccessTokenCarriesPermissionClaims(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:       "user-member",
			Email:    "member@example.com",
			Password: cryp.HashArgon2id("s3cret-password"),
			Status:   userConstants.UserStatusActive,
		},
	}
	roleRepository := &fakeRoleRepository{
		permissionCodes: []string{roleConstants.PERM_USER_READ, roleConstants.PERM_ROLE_READ},
	}
	cfg := newTestConfig(t)
	authUsecase := NewUsecase(cfg, nil, userRepository, &fakeUserSessionRepository{}, nil, roleRepository)

	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "member@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	claims, err := jwt.ValidateJWTWithRSAPublicKey(res.AccessToken, cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if claims.GetIsAdmin() {
		t.Error("expected adm claim to be false for a non-admin user")
	}
	perms := claims.GetPerms()
	if !slices.Contains(perms, roleConstants.PERM_USER_READ) || !slices.Contains(perms, roleConstants.PERM_ROLE_READ) {
		t.Errorf("expected perms claim to contain assigned permission codes, got %v", perms)
	}
}
