package usecase

import (
	"context"
	"testing"
	"time"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	userSessionConstants "github.com/vukyn/isme/internal/domains/user_session/constants"
	userSessionEntity "github.com/vukyn/isme/internal/domains/user_session/entity"

	pkgCtx "github.com/vukyn/kuery/ctx"
)

// newSessionFixture wires a usecase exercising the self-service session manager.
// Only the session repo is consulted by these methods, so the other deps are nil.
func newSessionFixture(sessionRepo *ssoUserSessionRepo) *usecase {
	return &usecase{userSessionRepo: sessionRepo}
}

// newSessionFixtureWithApps wires the session manager with an app_service repo so
// the enrichment path (ListMySessions → appServiceRepo.GetByIDs) is exercised.
func newSessionFixtureWithApps(sessionRepo *ssoUserSessionRepo, appRepo *enrichAppServiceRepo) *usecase {
	return &usecase{userSessionRepo: sessionRepo, appServiceRepo: appRepo}
}

// enrichAppServiceRepo is a configurable app_service repo stub that records the
// ids passed to GetByIDs and returns the seeded apps, so a test can assert both
// the single batched call and the enrichment it produces.
type enrichAppServiceRepo struct {
	ssoAppServiceRepo
	apps         map[string]appServiceEntity.AppService
	getByIDsArgs [][]string
}

func (e *enrichAppServiceRepo) GetByIDs(ctx context.Context, ids []string) (map[string]appServiceEntity.AppService, error) {
	e.getByIDsArgs = append(e.getByIDsArgs, ids)
	return e.apps, nil
}

// ctxWithUser builds a context carrying the user_id and token_id the same way
// the AuthMiddleware populates it before a handler runs.
func ctxWithUser(userID, tokenID string) context.Context {
	ctx := context.WithValue(context.Background(), pkgCtx.UserIDKey, userID)
	return context.WithValue(ctx, pkgCtx.TokenIDKey, tokenID)
}

// TestListMySessionsEmptyUserID proves a missing user_id is rejected instead of
// silently listing everyone's sessions.
func TestListMySessionsEmptyUserID(t *testing.T) {
	uc := newSessionFixture(&ssoUserSessionRepo{})

	_, err := uc.ListMySessions(context.Background())
	if err == nil {
		t.Fatal("expected an error when user_id is missing from context")
	}
}

// TestListMySessionsFlagsCurrent proves Current is true exactly for the session
// whose token ID matches the caller's request token, and never leaks the token.
func TestListMySessionsFlagsCurrent(t *testing.T) {
	const userID = "user-1"
	const currentToken = "token-current"

	sessionRepo := &ssoUserSessionRepo{
		activeSessions: []userSessionEntity.UserSession{
			{ID: "s1", UserID: userID, TokenID: currentToken, Status: userSessionConstants.UserSessionStatusActive},
			{ID: "s2", UserID: userID, TokenID: "token-other", Status: userSessionConstants.UserSessionStatusActive},
		},
	}
	uc := newSessionFixture(sessionRepo)

	items, err := uc.ListMySessions(ctxWithUser(userID, currentToken))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	for _, item := range items {
		switch item.ID {
		case "s1":
			if !item.Current {
				t.Error("expected s1 (matching request token) to be flagged Current")
			}
		case "s2":
			if item.Current {
				t.Error("expected s2 (different token) to NOT be flagged Current")
			}
		}
	}
}

// TestRevokeMySessionForeignSession proves a session owned by another user is
// treated as not-found, so a caller cannot probe or revoke others' sessions.
func TestRevokeMySessionForeignSession(t *testing.T) {
	sessionRepo := &ssoUserSessionRepo{
		session: userSessionEntity.UserSession{
			ID:      "foreign",
			UserID:  "someone-else",
			TokenID: "token-foreign",
		},
	}
	uc := newSessionFixture(sessionRepo)

	err := uc.RevokeMySession(ctxWithUser("user-1", "token-current"), "foreign")
	if err == nil {
		t.Fatal("expected not-found error revoking a foreign session")
	}
	if len(sessionRepo.inactiveByIDCalls) != 0 {
		t.Errorf("expected no revoke to be issued for a foreign session, got %v", sessionRepo.inactiveByIDCalls)
	}
}

// TestRevokeMySessionRejectsCurrent proves the current session cannot be revoked
// via the session manager (logout owns that path).
func TestRevokeMySessionRejectsCurrent(t *testing.T) {
	const userID = "user-1"
	const currentToken = "token-current"
	sessionRepo := &ssoUserSessionRepo{
		session: userSessionEntity.UserSession{
			ID:      "s1",
			UserID:  userID,
			TokenID: currentToken,
		},
	}
	uc := newSessionFixture(sessionRepo)

	err := uc.RevokeMySession(ctxWithUser(userID, currentToken), "s1")
	if err == nil {
		t.Fatal("expected an error revoking the current session")
	}
	if len(sessionRepo.inactiveByIDCalls) != 0 {
		t.Errorf("expected no revoke for the current session, got %v", sessionRepo.inactiveByIDCalls)
	}
}

// TestRevokeMyOtherSessionsKeepsCurrent proves the current token is passed as the
// exception so it survives the bulk revoke.
func TestRevokeMyOtherSessionsKeepsCurrent(t *testing.T) {
	const currentToken = "token-current"
	sessionRepo := &ssoUserSessionRepo{}
	uc := newSessionFixture(sessionRepo)

	if err := uc.RevokeMyOtherSessions(ctxWithUser("user-1", currentToken)); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(sessionRepo.exceptTokenIDCalls) != 1 || sessionRepo.exceptTokenIDCalls[0] != currentToken {
		t.Errorf("expected bulk revoke to preserve the current token, got %v", sessionRepo.exceptTokenIDCalls)
	}
}

// TestCountMySessionsDelta proves the total + 24h delta are surfaced.
func TestCountMySessionsDelta(t *testing.T) {
	const userID = "user-1"
	sessionRepo := &ssoUserSessionRepo{newIn24h: 2}
	sessionRepo.session = userSessionEntity.UserSession{}
	// CountActiveByUserIDs returns nil map in the fake → Count resolves to 0.
	uc := newSessionFixture(sessionRepo)

	count, err := uc.CountMySessions(ctxWithUser(userID, "token-current"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if count.NewIn24h != 2 {
		t.Errorf("expected NewIn24h=2, got %d", count.NewIn24h)
	}
}

// TestCountMySessionsRotations proves the sliding-24h rotation count is surfaced
// from the event log and last_refreshed_at is the MAX across the user's active
// sessions (the most-recent rotation), formatted RFC3339.
func TestCountMySessionsRotations(t *testing.T) {
	const userID = "user-1"
	older := time.Now().UTC().Add(-3 * time.Hour)
	newest := time.Now().UTC().Add(-5 * time.Minute)

	sessionRepo := &ssoUserSessionRepo{
		rotations24h: 7,
		activeSessions: []userSessionEntity.UserSession{
			{ID: "s1", UserID: userID, LastRefreshedAt: &older, Status: userSessionConstants.UserSessionStatusActive},
			{ID: "s2", UserID: userID, LastRefreshedAt: &newest, Status: userSessionConstants.UserSessionStatusActive},
			// never-refreshed session contributes no timestamp
			{ID: "s3", UserID: userID, LastRefreshedAt: nil, Status: userSessionConstants.UserSessionStatusActive},
		},
	}
	uc := newSessionFixture(sessionRepo)

	count, err := uc.CountMySessions(ctxWithUser(userID, "token-current"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if count.Rotations24h != 7 {
		t.Errorf("expected Rotations24h=7, got %d", count.Rotations24h)
	}
	want := newest.Format(time.RFC3339)
	if count.LastRefreshedAt != want {
		t.Errorf("expected LastRefreshedAt=%q (max across sessions), got %q", want, count.LastRefreshedAt)
	}
}

// TestCountMySessionsNoRotations proves an all-never-refreshed session set yields
// an empty last_refreshed_at (the card shows "no refreshes yet").
func TestCountMySessionsNoRotations(t *testing.T) {
	const userID = "user-1"
	sessionRepo := &ssoUserSessionRepo{
		rotations24h: 0,
		activeSessions: []userSessionEntity.UserSession{
			{ID: "s1", UserID: userID, LastRefreshedAt: nil, Status: userSessionConstants.UserSessionStatusActive},
		},
	}
	uc := newSessionFixture(sessionRepo)

	count, err := uc.CountMySessions(ctxWithUser(userID, "token-current"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if count.Rotations24h != 0 {
		t.Errorf("expected Rotations24h=0, got %d", count.Rotations24h)
	}
	if count.LastRefreshedAt != "" {
		t.Errorf("expected empty LastRefreshedAt when never refreshed, got %q", count.LastRefreshedAt)
	}
}

// TestListMySessionsRotationFields proves the per-session refresh_count and
// last_refreshed_at are mapped through to the DTO, with the never-refreshed
// session reporting count 0 and empty timestamp.
func TestListMySessionsRotationFields(t *testing.T) {
	const userID = "user-1"
	refreshedAt := time.Now().UTC().Add(-18 * time.Minute)

	sessionRepo := &ssoUserSessionRepo{
		activeSessions: []userSessionEntity.UserSession{
			{ID: "s1", UserID: userID, RefreshCount: 12, LastRefreshedAt: &refreshedAt, Status: userSessionConstants.UserSessionStatusActive},
			{ID: "s2", UserID: userID, RefreshCount: 0, LastRefreshedAt: nil, Status: userSessionConstants.UserSessionStatusActive},
		},
	}
	uc := newSessionFixture(sessionRepo)

	items, err := uc.ListMySessions(ctxWithUser(userID, "token-current"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	byID := make(map[string]struct {
		count int64
		last  string
	})
	for _, item := range items {
		byID[item.ID] = struct {
			count int64
			last  string
		}{item.RefreshCount, item.LastRefreshedAt}
	}

	if got := byID["s1"]; got.count != 12 || got.last != refreshedAt.Format(time.RFC3339) {
		t.Errorf("expected s1 refresh_count=12 + RFC3339 last_refreshed_at, got %+v", got)
	}
	if got := byID["s2"]; got.count != 0 || got.last != "" {
		t.Errorf("expected never-refreshed s2 to have count 0 + empty last_refreshed_at, got %+v", got)
	}
}

// TestListMySessionsEnrichesApps proves SSO sessions get app name/icon/color from
// a SINGLE batched GetByIDs (distinct, non-empty ids only), first-party isme
// sessions (empty AppServiceID) are left blank, and a session whose app_service is
// missing from the map is left blank without erroring.
func TestListMySessionsEnrichesApps(t *testing.T) {
	const userID = "user-1"

	sessionRepo := &ssoUserSessionRepo{
		activeSessions: []userSessionEntity.UserSession{
			{ID: "s1", UserID: userID, AppServiceID: "app-medioa", Status: userSessionConstants.UserSessionStatusActive},
			// duplicate app id → must collapse to one id in the lookup
			{ID: "s2", UserID: userID, AppServiceID: "app-medioa", Status: userSessionConstants.UserSessionStatusActive},
			// first-party isme session → no lookup, blank app_* fields
			{ID: "s3", UserID: userID, AppServiceID: "", Status: userSessionConstants.UserSessionStatusActive},
			// deleted app_service (absent from map) → blank app_* fields, no error
			{ID: "s4", UserID: userID, AppServiceID: "app-gone", Status: userSessionConstants.UserSessionStatusActive},
		},
	}
	appRepo := &enrichAppServiceRepo{
		apps: map[string]appServiceEntity.AppService{
			"app-medioa": {ID: "app-medioa", AppName: "Medioa", Icon: "cloud", Color: "cyan"},
		},
	}
	uc := newSessionFixtureWithApps(sessionRepo, appRepo)

	items, err := uc.ListMySessions(ctxWithUser(userID, "token-current"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// single batched call carrying the two DISTINCT non-empty ids
	if len(appRepo.getByIDsArgs) != 1 {
		t.Fatalf("expected exactly one batched GetByIDs call, got %d", len(appRepo.getByIDsArgs))
	}
	if len(appRepo.getByIDsArgs[0]) != 2 {
		t.Errorf("expected 2 distinct ids in the lookup, got %v", appRepo.getByIDsArgs[0])
	}

	byID := make(map[string]struct {
		name, icon, color string
	})
	for _, item := range items {
		byID[item.ID] = struct{ name, icon, color string }{item.AppName, item.AppIcon, item.AppColor}
	}

	for _, id := range []string{"s1", "s2"} {
		got := byID[id]
		if got.name != "Medioa" || got.icon != "cloud" || got.color != "cyan" {
			t.Errorf("expected %s enriched with Medioa app, got %+v", id, got)
		}
	}
	if got := byID["s3"]; got.name != "" || got.icon != "" || got.color != "" {
		t.Errorf("expected first-party session s3 to have blank app_* fields, got %+v", got)
	}
	if got := byID["s4"]; got.name != "" || got.icon != "" || got.color != "" {
		t.Errorf("expected deleted-app session s4 to have blank app_* fields, got %+v", got)
	}
}

// TestListMySessionsSkipsLookupWhenAllFirstParty proves an all-first-party session
// set never issues the GetByIDs query (no wasted call, no N+1).
func TestListMySessionsSkipsLookupWhenAllFirstParty(t *testing.T) {
	const userID = "user-1"
	sessionRepo := &ssoUserSessionRepo{
		activeSessions: []userSessionEntity.UserSession{
			{ID: "s1", UserID: userID, AppServiceID: "", Status: userSessionConstants.UserSessionStatusActive},
			{ID: "s2", UserID: userID, AppServiceID: "", Status: userSessionConstants.UserSessionStatusActive},
		},
	}
	appRepo := &enrichAppServiceRepo{}
	uc := newSessionFixtureWithApps(sessionRepo, appRepo)

	if _, err := uc.ListMySessions(ctxWithUser(userID, "token-current")); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(appRepo.getByIDsArgs) != 0 {
		t.Errorf("expected no GetByIDs call for an all-first-party set, got %v", appRepo.getByIDsArgs)
	}
}
