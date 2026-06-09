package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/config"
	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	appServiceModels "github.com/vukyn/isme/internal/domains/app_service/models"
	"github.com/vukyn/isme/internal/domains/auth/models"
	userConstants "github.com/vukyn/isme/internal/domains/user/constants"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionEntity "github.com/vukyn/isme/internal/domains/user_session/entity"
	userSessionModels "github.com/vukyn/isme/internal/domains/user_session/models"

	pkgCache "github.com/vukyn/kuery/cache"
	"github.com/vukyn/kuery/jwt"
)

// === configurable stubs for the SSO consent flow ===

// ssoAppServiceRepo is a minimal app_service repo stub returning a fixed app.
type ssoAppServiceRepo struct {
	app appServiceEntity.AppService
}

func (s *ssoAppServiceRepo) Create(ctx context.Context, req appServiceEntity.CreateRequest) (string, error) {
	return "", nil
}
func (s *ssoAppServiceRepo) GetByID(ctx context.Context, id string) (appServiceEntity.AppService, error) {
	return s.app, nil
}
func (s *ssoAppServiceRepo) GetByIDs(ctx context.Context, ids []string) (map[string]appServiceEntity.AppService, error) {
	return nil, nil
}
func (s *ssoAppServiceRepo) GetByCode(ctx context.Context, code string) (appServiceEntity.AppService, error) {
	return appServiceEntity.AppService{}, nil
}
func (s *ssoAppServiceRepo) Update(ctx context.Context, req appServiceEntity.UpdateRequest) error {
	return nil
}
func (s *ssoAppServiceRepo) List(ctx context.Context, req appServiceModels.ListRequest) ([]appServiceEntity.AppService, int64, error) {
	return nil, 0, nil
}
func (s *ssoAppServiceRepo) UpdateStatus(ctx context.Context, id string, status int32) error {
	return nil
}

// ssoUserSessionRepo returns a fixed session for every lookup and records the
// number of rotation (UpdateLastLogin) calls and the created sessions so a test
// can assert neither /check nor /consent rotates the caller's session, and that
// /consent creates a fresh app-bound session instead.
type ssoUserSessionRepo struct {
	session         userSessionEntity.UserSession
	updateCalls     int
	createCalls     int
	createdSessions []userSessionModels.CreateRequest
}

func (s *ssoUserSessionRepo) Create(ctx context.Context, req userSessionModels.CreateRequest) (userSessionEntity.UserSession, error) {
	s.createCalls++
	s.createdSessions = append(s.createdSessions, req)
	return userSessionEntity.UserSession{ID: "session-id"}, nil
}
func (s *ssoUserSessionRepo) UpdateLastLogin(ctx context.Context, req userSessionModels.UpdateLastLoginRequest) error {
	s.updateCalls++
	return nil
}
func (s *ssoUserSessionRepo) InactiveAllUserSession(ctx context.Context, userID string) error {
	return nil
}
func (s *ssoUserSessionRepo) InactiveSessionByTokenID(ctx context.Context, tokenID string) error {
	return nil
}
func (s *ssoUserSessionRepo) InactiveSessionByID(ctx context.Context, sessionID string) error {
	return nil
}
func (s *ssoUserSessionRepo) FindByRefreshToken(ctx context.Context, refreshToken string) (userSessionEntity.UserSession, error) {
	return s.session, nil
}
func (s *ssoUserSessionRepo) FindByTokenID(ctx context.Context, tokenID string) (userSessionEntity.UserSession, error) {
	return s.session, nil
}
func (s *ssoUserSessionRepo) GetByID(ctx context.Context, sessionID string) (userSessionEntity.UserSession, error) {
	return s.session, nil
}
func (s *ssoUserSessionRepo) GetListActiveByUserID(ctx context.Context, userID string) ([]userSessionEntity.UserSession, error) {
	return nil, nil
}
func (s *ssoUserSessionRepo) CountActiveByUserIDs(ctx context.Context, userIDs []string) (map[string]int, error) {
	return nil, nil
}

// ssoFixture wires a usecase with controllable cache, session, user and app.
type ssoFixture struct {
	usecase      *usecase
	cache        *pkgCache.Cache[string, string]
	sessionRepo  *ssoUserSessionRepo
	cfg          *config.Config
	accessToken  string
	refreshToken string
	sessionID    string
}

func newSSOFixture(t *testing.T, sessionStatus int32, sessionExpiresAt time.Time, userStatus int32) *ssoFixture {
	t.Helper()

	cfg := newTestConfig(t)
	cache := pkgCache.NewCache[string, string]()

	const userID = "user-sso"
	const email = "sso@example.com"

	user := userEntity.User{
		ID:     userID,
		Name:   "Thao Nguyen",
		Email:  email,
		Status: userStatus,
	}
	userRepo := &fakeUserRepository{user: user}

	sessionRepo := &ssoUserSessionRepo{
		session: userSessionEntity.UserSession{
			ID:           "session-id",
			UserID:       userID,
			Email:        email,
			Status:       sessionStatus,
			ExpiresAt:    sessionExpiresAt,
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

	uc := NewUsecase(cfg, cache, userRepo, sessionRepo, appRepo, &fakeRoleRepository{}).(*usecase)

	// live access token (token_id is random; the session stub matches any lookup)
	accessToken, _, err := jwt.GenerateJWTWithRSAPrivateKey(cfg.Auth.AccessTokenPrivateKey, cfg.Auth.AccessTokenExpireIn, userID, email)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	// live refresh token
	refreshToken, _, err := jwt.GenerateJWT(cfg.Auth.RefreshTokenSecretKey, cfg.Auth.RefreshTokenExpireIn, userID, email)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	sessionID := "sess-" + cfg.Auth.AppCode
	cache.Set(sessionID, "app-1", time.Minute)

	return &ssoFixture{
		usecase:      uc,
		cache:        cache,
		sessionRepo:  sessionRepo,
		cfg:          cfg,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		sessionID:    sessionID,
	}
}

func expiredAccessToken(t *testing.T, cfg *config.Config) string {
	t.Helper()
	token, _, err := jwt.GenerateJWTWithRSAPrivateKey(cfg.Auth.AccessTokenPrivateKey, -3600, "user-sso", "sso@example.com")
	if err != nil {
		t.Fatalf("failed to generate expired access token: %v", err)
	}
	return token
}

// (a) valid live access token → valid, no rotation.
func TestSSOCheckValidAccessToken(t *testing.T) {
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusActive, time.Now().Add(time.Hour), userConstants.UserStatusActive)

	res, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Valid {
		t.Fatal("expected valid=true for live access token")
	}
	if res.User.Email != "sso@example.com" || res.User.Name != "Thao Nguyen" {
		t.Errorf("unexpected user identity: %+v", res.User)
	}
	if res.App.Name != "medioa2" {
		t.Errorf("unexpected app name: %q", res.App.Name)
	}
	if len(res.Scopes) != 3 {
		t.Errorf("expected 3 static scopes, got %d", len(res.Scopes))
	}
	if res.Nonce == "" {
		t.Error("expected a nonce to be minted")
	}
	// read-only: /check must not rotate the session
	if f.sessionRepo.updateCalls != 0 {
		t.Errorf("expected 0 rotation calls in /check, got %d", f.sessionRepo.updateCalls)
	}
}

// (b) expired access token + valid refresh + ACTIVE session → valid.
// The session's ExpiresAt is intentionally in the PAST: that field tracks the
// ACCESS token's expiry (set from accessTokenClaims), so an expired access
// token always leaves it past. The refresh probe must still accept this — a
// future ExpiresAt here would hide the regression where the probe wrongly
// gated on it and dropped valid-refresh sessions to the password form.
func TestSSOCheckExpiredAccessValidRefresh(t *testing.T) {
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusActive, time.Now().Add(-time.Minute), userConstants.UserStatusActive)

	res, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    f.sessionID,
		AccessToken:  expiredAccessToken(t, f.cfg),
		RefreshToken: f.refreshToken,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Valid {
		t.Fatal("expected valid=true via the refresh-token probe")
	}
	if f.sessionRepo.updateCalls != 0 {
		t.Errorf("expected 0 rotation calls in /check, got %d", f.sessionRepo.updateCalls)
	}
}

// (c) expired access + revoked/inactive session → invalid (no error).
func TestSSOCheckRevokedSessionInvalid(t *testing.T) {
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusInactive, time.Now().Add(time.Hour), userConstants.UserStatusActive)

	res, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    f.sessionID,
		AccessToken:  expiredAccessToken(t, f.cfg),
		RefreshToken: f.refreshToken,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Valid {
		t.Fatal("expected valid=false for a revoked session")
	}
	if res.Nonce != "" {
		t.Error("expected no nonce when invalid")
	}
}

// (c.2) expired access + inactive user → invalid.
func TestSSOCheckInactiveUserInvalid(t *testing.T) {
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusActive, time.Now().Add(time.Hour), userConstants.UserStatusInactive)

	res, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    f.sessionID,
		AccessToken:  expiredAccessToken(t, f.cfg),
		RefreshToken: f.refreshToken,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Valid {
		t.Fatal("expected valid=false for an inactive user")
	}
}

// (d) missing/expired session_id → error.
func TestSSOCheckMissingSessionID(t *testing.T) {
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusActive, time.Now().Add(time.Hour), userConstants.UserStatusActive)

	_, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    "does-not-exist",
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
	})
	if err == nil {
		t.Fatal("expected an error for an unknown session_id")
	}
}

// (e) consent mints a single-use code, consumes the nonce, deletes session_id,
// does NOT rotate the caller's browser session, and creates a fresh app-scoped
// session bound to the requesting app whose token is aud-restricted to that app.
func TestSSOConsentMintsCodeAndConsumesNonce(t *testing.T) {
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusActive, time.Now().Add(time.Hour), userConstants.UserStatusActive)
	// give the user some perms in the requesting app so the scoped token carries them
	f.usecase.roleRepo = &fakeRoleRepository{
		groupedPermissionCodes: map[string][]string{
			"medioa2": {"storage:read", "storage:write"},
			"rainy":   {"playlist:read"},
		},
	}

	check, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
	})
	if err != nil || !check.Valid {
		t.Fatalf("setup check failed: err=%v valid=%v", err, check.Valid)
	}

	res, err := f.usecase.SSOConsent(context.Background(), models.SSOConsentRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
		Nonce:        check.Nonce,
	})
	if err != nil {
		t.Fatalf("expected consent to succeed, got %v", err)
	}
	if res.AuthorizationCode == "" {
		t.Error("expected an authorization code")
	}
	if res.RedirectURL != "https://app.medioa.local/callback" {
		t.Errorf("unexpected redirect_url: %q", res.RedirectURL)
	}

	// consent must NOT rotate/invalidate the caller's existing browser session
	if f.sessionRepo.updateCalls != 0 {
		t.Errorf("expected 0 rotation calls in /consent, got %d", f.sessionRepo.updateCalls)
	}

	// the caller's refresh token is still valid against an ACTIVE session after
	// consent: the browser session row's refresh token is unchanged, so a
	// subsequent silent re-validation still authenticates the same tokens.
	if _, valid := f.usecase.validateSessionForConsent(context.Background(), f.accessToken, f.refreshToken); !valid {
		t.Error("expected the caller's session to remain valid after consent (not rotated)")
	}

	// a NEW user_session was created, bound to the requesting app
	if f.sessionRepo.createCalls != 1 {
		t.Fatalf("expected exactly 1 new session created in /consent, got %d", f.sessionRepo.createCalls)
	}
	created := f.sessionRepo.createdSessions[0]
	if created.AppServiceID != "app-1" {
		t.Errorf("expected new session bound to requesting app (app-1), got %q", created.AppServiceID)
	}

	// the authorization code resolves to the freshly minted tokens (one-time)
	exchange, err := f.usecase.ExchangeCode(context.Background(), models.ExchangeCodeRequest{
		AuthorizationCode: res.AuthorizationCode,
	})
	if err != nil {
		t.Fatalf("expected exchange to succeed, got %v", err)
	}
	if exchange.AccessToken == "" || exchange.RefreshToken == "" {
		t.Error("expected exchanged tokens to be non-empty")
	}

	// the minted access token is aud-restricted to the requesting app, NOT a
	// full multi-app token
	claims, err := jwt.ValidateJWTWithRSAPublicKey(exchange.AccessToken, f.cfg.Auth.AccessTokenPublicKey)
	if err != nil {
		t.Fatalf("failed to validate minted access token: %v", err)
	}
	audience := claims.GetAudience()
	if len(audience) != 1 || audience[0] != "medioa2" {
		t.Errorf("expected token aud-restricted to [medioa2], got %v", audience)
	}
	if perms := claims.GetPermsForApp("rainy"); len(perms) != 0 {
		t.Errorf("expected resource_access scoped to requesting app only, leaked rainy perms: %v", perms)
	}
	if perms := claims.GetPermsForApp("medioa2"); len(perms) == 0 {
		t.Errorf("expected resource_access to carry the requesting app perms, got none")
	}

	// session_id deleted from cache → a second consent cannot mint again
	if _, ok := f.cache.Get(f.sessionID); ok {
		t.Error("expected session_id to be deleted from cache after mint")
	}
}

// (f) consent rejects a bad nonce and a reused nonce.
func TestSSOConsentRejectsBadAndReusedNonce(t *testing.T) {
	// bad nonce
	f := newSSOFixture(t, userSessionConstants.UserSessionStatusActive, time.Now().Add(time.Hour), userConstants.UserStatusActive)
	check, err := f.usecase.SSOCheck(context.Background(), models.SSOCheckRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
	})
	if err != nil || !check.Valid {
		t.Fatalf("setup check failed: err=%v valid=%v", err, check.Valid)
	}

	if _, err := f.usecase.SSOConsent(context.Background(), models.SSOConsentRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
		Nonce:        "not-the-real-nonce",
	}); err == nil {
		t.Fatal("expected consent to reject a bad nonce")
	}
	// session_id must survive a rejected attempt (no mint happened)
	if _, ok := f.cache.Get(f.sessionID); !ok {
		t.Error("session_id should not be deleted on a rejected nonce")
	}

	// reused nonce: first consent consumes it, second must fail
	if _, err := f.usecase.SSOConsent(context.Background(), models.SSOConsentRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
		Nonce:        check.Nonce,
	}); err != nil {
		t.Fatalf("expected first consent to succeed, got %v", err)
	}
	if _, err := f.usecase.SSOConsent(context.Background(), models.SSOConsentRequest{
		SessionID:    f.sessionID,
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
		Nonce:        check.Nonce,
	}); err == nil {
		t.Fatal("expected reused nonce (and consumed session) to be rejected")
	}
}
