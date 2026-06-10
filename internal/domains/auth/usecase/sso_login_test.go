package usecase

import (
	"context"
	"slices"
	"testing"
	"time"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/auth/models"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	userConstants "github.com/vukyn/isme/internal/domains/user/constants"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionEntity "github.com/vukyn/isme/internal/domains/user_session/entity"

	pkgCache "github.com/vukyn/kuery/cache"
	"github.com/vukyn/kuery/cryp"
	"github.com/vukyn/kuery/jwt"
)

// newSSOLoginFixture wires a usecase for exercising the SSO branch of Login. It
// reuses the in-package fakes (fakeUserRepository, ssoUserSessionRepo,
// ssoAppServiceRepo, fakeRoleRepository) so created sessions can be inspected.
// An empty sessionID exercises the non-SSO branch.
func newSSOLoginFixture(t *testing.T, sessionID string, grouped map[string][]string) (*usecase, *ssoUserSessionRepo, string) {
	t.Helper()

	const userID = "user-sso"
	const email = "sso@example.com"
	const password = "s3cret-password"

	cfg := newTestConfig(t)
	cache := pkgCache.NewCache[string, string]()

	user := userEntity.User{
		ID:         userID,
		Name:       "Thao Nguyen",
		Email:      email,
		Password:   cryp.HashArgon2id(password),
		Status:     userConstants.UserStatusActive,
		IsVerified: true,
	}
	userRepository := &fakeUserRepository{user: user}

	sessionRepo := &ssoUserSessionRepo{
		session: userSessionEntity.UserSession{
			ID:           "session-id",
			UserID:       userID,
			Email:        email,
			Status:       userSessionConstants.UserSessionStatusActive,
			ExpiresAt:    time.Now().Add(time.Hour),
			AppServiceID: "app-1",
		},
	}

	appRepo := &ssoAppServiceRepo{
		app: appServiceEntity.AppService{
			ID:          "app-1",
			AppCode:     "medioa2",
			AppName:     "medioa2",
			RedirectURL: "https://app.medioa.local/callback",
		},
	}

	roleRepo := &fakeRoleRepository{groupedPermissionCodes: grouped}

	uc := NewUsecase(cfg, cache, userRepository, sessionRepo, appRepo, roleRepo).(*usecase)

	if sessionID != "" {
		cache.Set(sessionID, "app-1", time.Minute)
	}

	return uc, sessionRepo, password
}

// TestSSOLoginReturnsIdPTokensAndCode is the regression guard: an SSO login must
// now return BOTH the authorization code (app handoff) AND non-empty IdP tokens
// (browser cookies). Previously the response tokens were sanitized to empty.
func TestSSOLoginReturnsIdPTokensAndCode(t *testing.T) {
	const sessionID = "sess-1"
	uc, _, password := newSSOLoginFixture(t, sessionID, map[string][]string{
		"medioa2": {"storage:read"},
	})

	resp, err := uc.Login(context.Background(), models.LoginRequest{
		Email:     "sso@example.com",
		Password:  password,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("expected SSO login to succeed, got %v", err)
	}

	if resp.AuthorizationCode == "" {
		t.Error("expected a non-empty authorization_code (app handoff)")
	}
	if resp.RedirectURL == "" {
		t.Error("expected a non-empty redirect_url")
	}
	if resp.AccessToken == "" {
		t.Error("expected a non-empty access_token (IdP browser session) — regression")
	}
	if resp.RefreshToken == "" {
		t.Error("expected a non-empty refresh_token (IdP browser session) — regression")
	}
	if resp.ExpiresAt == "" {
		t.Error("expected a non-empty expires_at (IdP browser session) — regression")
	}
}

// TestSSOLoginIdPTokenIsFullScope proves the two token pairs are distinct and not
// crossed: the response access token is full-scope (audience includes isme), and
// the authorization-code-exchanged token is aud-restricted to the requesting app.
func TestSSOLoginIdPTokenIsFullScope(t *testing.T) {
	const sessionID = "sess-2"
	uc, _, password := newSSOLoginFixture(t, sessionID, map[string][]string{
		"medioa2": {"storage:read", "storage:write"},
		"rainy":   {"playlist:read"},
	})

	resp, err := uc.Login(context.Background(), models.LoginRequest{
		Email:     "sso@example.com",
		Password:  password,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("expected SSO login to succeed, got %v", err)
	}

	// (b) the RESPONSE access token is full-scope: audience contains isme
	idpClaims, err := jwt.ValidateJWTWithRSAPublicKey(resp.AccessToken, uc.cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate IdP access token: %v", err)
	}
	idpAud := idpClaims.GetAudience()
	if !slices.Contains(idpAud, roleConstants.APP_CODE_ISME) {
		t.Errorf("expected IdP token audience to contain isme (full scope), got %v", idpAud)
	}
	if len(idpAud) == 1 && idpAud[0] == "medioa2" {
		t.Errorf("IdP token must NOT be aud-restricted to the app, got %v", idpAud)
	}

	// (a) the authorization code resolves to AUD-RESTRICTED tokens for the app
	exchange, err := uc.ExchangeCode(context.Background(), models.ExchangeCodeRequest{
		AuthorizationCode: resp.AuthorizationCode,
	})
	if err != nil {
		t.Fatalf("expected exchange to succeed, got %v", err)
	}
	appClaims, err := jwt.ValidateJWTWithRSAPublicKey(exchange.AccessToken, uc.cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate exchanged app token: %v", err)
	}
	appAud := appClaims.GetAudience()
	if len(appAud) != 1 || appAud[0] != "medioa2" {
		t.Errorf("expected exchanged token aud-restricted to [medioa2], got %v", appAud)
	}
	if perms := appClaims.GetPermsForApp("rainy"); len(perms) != 0 {
		t.Errorf("exchanged token leaked non-app perms (rainy): %v", perms)
	}

	// the two access tokens must be different strings (not crossed)
	if resp.AccessToken == exchange.AccessToken {
		t.Error("IdP and app access tokens must be distinct — they appear crossed")
	}
}

// TestSSOLoginCreatesIdPSessionEmptyAppServiceID verifies exactly two sessions are
// created: the app-bound session (AppServiceID="app-1") and the IdP session
// (AppServiceID="").
func TestSSOLoginCreatesIdPSessionEmptyAppServiceID(t *testing.T) {
	const sessionID = "sess-3"
	uc, sessionRepo, password := newSSOLoginFixture(t, sessionID, map[string][]string{
		"medioa2": {"storage:read"},
	})

	_, err := uc.Login(context.Background(), models.LoginRequest{
		Email:     "sso@example.com",
		Password:  password,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("expected SSO login to succeed, got %v", err)
	}

	if sessionRepo.createCalls != 2 {
		t.Fatalf("expected exactly 2 sessions created (app + IdP), got %d", sessionRepo.createCalls)
	}

	var hasApp, hasIdP bool
	for _, created := range sessionRepo.createdSessions {
		switch created.AppServiceID {
		case "app-1":
			hasApp = true
		case "":
			hasIdP = true
		}
	}
	if !hasApp {
		t.Error("expected one session bound to the requesting app (AppServiceID=app-1)")
	}
	if !hasIdP {
		t.Error("expected one IdP session (AppServiceID=\"\")")
	}
}

// TestNonSSOLoginUnchanged verifies the non-SSO branch is untouched: full-scope
// tokens, no authorization code, exactly one IdP session (AppServiceID="").
func TestNonSSOLoginUnchanged(t *testing.T) {
	uc, sessionRepo, password := newSSOLoginFixture(t, "", map[string][]string{
		"medioa2": {"storage:read"},
	})

	resp, err := uc.Login(context.Background(), models.LoginRequest{
		Email:    "sso@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("expected login to succeed, got %v", err)
	}

	if resp.AuthorizationCode != "" {
		t.Errorf("expected no authorization_code for non-SSO login, got %q", resp.AuthorizationCode)
	}
	if resp.RedirectURL != "" {
		t.Errorf("expected no redirect_url for non-SSO login, got %q", resp.RedirectURL)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Error("expected non-empty tokens for non-SSO login")
	}

	claims, err := jwt.ValidateJWTWithRSAPublicKey(resp.AccessToken, uc.cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if !slices.Contains(claims.GetAudience(), roleConstants.APP_CODE_ISME) {
		t.Errorf("expected full-scope token (audience contains isme), got %v", claims.GetAudience())
	}

	if sessionRepo.createCalls != 1 {
		t.Fatalf("expected exactly 1 session for non-SSO login, got %d", sessionRepo.createCalls)
	}
	if sessionRepo.createdSessions[0].AppServiceID != "" {
		t.Errorf("expected non-SSO session to have empty AppServiceID, got %q", sessionRepo.createdSessions[0].AppServiceID)
	}
}
