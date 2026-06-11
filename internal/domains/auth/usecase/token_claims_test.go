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

func TestFirstPartyLoginAccessTokenCarriesResourceAccess(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-member",
			Email:      "member@example.com",
			Password:   cryp.HashArgon2id("s3cret-password"),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	roleRepository := &fakeRoleRepository{
		groupedPermissionCodes: map[string][]string{
			roleConstants.APP_CODE_ISME: {roleConstants.PERM_USER_READ, roleConstants.PERM_ROLE_READ},
		},
	}
	cfg := newTestConfig(t)
	authUsecase := NewUsecase(cfg, nil, userRepository, &fakeUserSessionRepository{}, nil, roleRepository, &fakeActivityUsecase{})

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

	// first-party login carries the isme app in the audience
	if !slices.Contains(claims.GetAudience(), roleConstants.APP_CODE_ISME) {
		t.Errorf("expected audience to contain %q, got %v", roleConstants.APP_CODE_ISME, claims.GetAudience())
	}

	// resource_access[isme].perms carries the user's isme-app permission codes
	perms := claims.GetPermsForApp(roleConstants.APP_CODE_ISME)
	if !slices.Contains(perms, roleConstants.PERM_USER_READ) || !slices.Contains(perms, roleConstants.PERM_ROLE_READ) {
		t.Errorf("expected isme perms to contain the assigned codes, got %v", perms)
	}
}

func TestFirstPartyLoginIncludesAllAppsInAudience(t *testing.T) {
	userRepository := &fakeUserRepository{
		user: userEntity.User{
			ID:         "user-multi",
			Email:      "multi@example.com",
			Password:   cryp.HashArgon2id("s3cret-password"),
			Status:     userConstants.UserStatusActive,
			IsVerified: true,
		},
	}
	roleRepository := &fakeRoleRepository{
		groupedPermissionCodes: map[string][]string{
			roleConstants.APP_CODE_ISME: {roleConstants.PERM_USER_READ},
			"medioa2":                   {"object:read", "object:create"},
		},
	}
	cfg := newTestConfig(t)
	authUsecase := NewUsecase(cfg, nil, userRepository, &fakeUserSessionRepository{}, nil, roleRepository, &fakeActivityUsecase{})

	res, err := authUsecase.Login(context.Background(), models.LoginRequest{
		Email:    "multi@example.com",
		Password: "s3cret-password",
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	claims, err := jwt.ValidateJWTWithRSAPublicKey(res.AccessToken, cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	audience := claims.GetAudience()
	for _, want := range []string{roleConstants.APP_CODE_ISME, "medioa2"} {
		if !slices.Contains(audience, want) {
			t.Errorf("expected audience to contain %q, got %v", want, audience)
		}
	}
	if perms := claims.GetPermsForApp("medioa2"); !slices.Contains(perms, "object:create") {
		t.Errorf("expected medioa2 perms to contain object:create, got %v", perms)
	}
}
