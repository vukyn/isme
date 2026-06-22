package usecase

import (
	"context"
	"testing"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/auth/models"

	pkgCache "github.com/vukyn/kuery/cache"
	"github.com/vukyn/kuery/cryp/aes"
)

func TestEncodeDecodeSSOSession(t *testing.T) {
	t.Run("round trip preserves both fields", func(t *testing.T) {
		want := ssoSession{AppServiceID: "app-1", RedirectURL: "https://app.example.com/cb"}
		raw := encodeSSOSession(want)
		got, ok := decodeSSOSession(raw)
		if !ok {
			t.Fatal("decode reported not-ok for a freshly encoded session")
		}
		if got != want {
			t.Errorf("round trip = %+v, want %+v", got, want)
		}
	})

	t.Run("legacy bare app-id falls back", func(t *testing.T) {
		got, ok := decodeSSOSession("app-legacy")
		if !ok {
			t.Fatal("expected legacy bare id to decode ok")
		}
		if got.AppServiceID != "app-legacy" {
			t.Errorf("AppServiceID = %q, want %q", got.AppServiceID, "app-legacy")
		}
		if got.RedirectURL != "" {
			t.Errorf("expected empty RedirectURL for legacy session, got %q", got.RedirectURL)
		}
	})

	t.Run("empty raw reports not-ok", func(t *testing.T) {
		if _, ok := decodeSSOSession(""); ok {
			t.Error("expected not-ok for empty raw")
		}
	})
}

// requestLoginFixture builds a usecase + an app whose encrypted secret matches a
// known plaintext, so RequestLogin's secret check passes and the redirect_uri
// allowlist logic can be exercised.
func requestLoginFixture(t *testing.T, app appServiceEntity.AppService) (*usecase, *pkgCache.Cache[string, string], string) {
	t.Helper()

	const aesSecret = "test-aes-secret"
	const plainSecret = "plain-app-secret"

	encrypted, err := aes.Encrypt(plainSecret, aesSecret, app.CtxInfo)
	if err != nil {
		t.Fatalf("failed to encrypt app secret: %v", err)
	}
	app.AppSecret = encrypted

	cfg := newTestConfig(t)
	cfg.AES.Secret = aesSecret
	cfg.Auth.EndpointWebSSOLogin = "https://sso.isme.local/login"
	cfg.Auth.ExternalLoginSessionTTL = 300

	cache := pkgCache.NewCache[string, string]()
	appRepo := &byCodeAppServiceRepo{ssoAppServiceRepo: ssoAppServiceRepo{app: app}}
	uc := NewUsecase(cfg, cache, &fakeUserRepository{}, &ssoUserSessionRepo{}, appRepo, &fakeRoleRepository{}, &fakeActivityUsecase{}).(*usecase)

	return uc, cache, plainSecret
}

// byCodeAppServiceRepo extends the shared ssoAppServiceRepo so GetByCode (used by
// RequestLogin) also returns the fixture app.
type byCodeAppServiceRepo struct {
	ssoAppServiceRepo
}

func (r *byCodeAppServiceRepo) GetByCode(ctx context.Context, code string) (appServiceEntity.AppService, error) {
	return r.app, nil
}

func TestRequestLoginRedirectURI(t *testing.T) {
	const ctxInfo = "authen"
	baseApp := appServiceEntity.AppService{
		ID:           "app-1",
		AppCode:      "medioa2",
		AppName:      "medioa2",
		CtxInfo:      ctxInfo,
		RedirectURL:  "https://app.medioa.local/callback",
		RedirectURLs: `["https://app.medioa.local/callback/staging","https://medioa.hasaki.vn/oauth2/return"]`,
	}

	// extract the cached, frozen redirect from the only cache entry created.
	frozenRedirect := func(t *testing.T, cache *pkgCache.Cache[string, string], rawSessionID string) string {
		t.Helper()
		raw, ok := cache.Get(rawSessionID)
		if !ok {
			t.Fatal("session was not stored in cache")
		}
		session, ok := decodeSSOSession(raw)
		if !ok {
			t.Fatal("cached session failed to decode")
		}
		return session.RedirectURL
	}

	// pull the session_id out of the returned redirect_url query string.
	sessionIDFromResponse := func(t *testing.T, redirectURL string) string {
		t.Helper()
		const marker = "session_id="
		idx := -1
		for i := 0; i+len(marker) <= len(redirectURL); i++ {
			if redirectURL[i:i+len(marker)] == marker {
				idx = i + len(marker)
				break
			}
		}
		if idx == -1 {
			t.Fatalf("no session_id in redirect_url %q", redirectURL)
		}
		return redirectURL[idx:]
	}

	t.Run("no redirect_uri freezes primary redirect_url", func(t *testing.T) {
		uc, cache, secret := requestLoginFixture(t, baseApp)
		resp, err := uc.RequestLogin(context.Background(), models.RequestLoginRequest{
			AppCode:   baseApp.AppCode,
			AppSecret: secret,
			CtxInfo:   ctxInfo,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sessionID := sessionIDFromResponse(t, resp.RedirectURL)
		if got := frozenRedirect(t, cache, sessionID); got != baseApp.RedirectURL {
			t.Errorf("frozen redirect = %q, want primary %q", got, baseApp.RedirectURL)
		}
	})

	t.Run("allowed redirect_uri from allowlist freezes that URL", func(t *testing.T) {
		uc, cache, secret := requestLoginFixture(t, baseApp)
		want := "https://medioa.hasaki.vn/oauth2/return"
		resp, err := uc.RequestLogin(context.Background(), models.RequestLoginRequest{
			AppCode:     baseApp.AppCode,
			AppSecret:   secret,
			CtxInfo:     ctxInfo,
			RedirectURI: want,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sessionID := sessionIDFromResponse(t, resp.RedirectURL)
		if got := frozenRedirect(t, cache, sessionID); got != want {
			t.Errorf("frozen redirect = %q, want %q", got, want)
		}
	})

	t.Run("redirect_uri matching primary is allowed", func(t *testing.T) {
		uc, cache, secret := requestLoginFixture(t, baseApp)
		resp, err := uc.RequestLogin(context.Background(), models.RequestLoginRequest{
			AppCode:     baseApp.AppCode,
			AppSecret:   secret,
			CtxInfo:     ctxInfo,
			RedirectURI: baseApp.RedirectURL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sessionID := sessionIDFromResponse(t, resp.RedirectURL)
		if got := frozenRedirect(t, cache, sessionID); got != baseApp.RedirectURL {
			t.Errorf("frozen redirect = %q, want %q", got, baseApp.RedirectURL)
		}
	})

	t.Run("disallowed redirect_uri rejected", func(t *testing.T) {
		uc, _, secret := requestLoginFixture(t, baseApp)
		_, err := uc.RequestLogin(context.Background(), models.RequestLoginRequest{
			AppCode:     baseApp.AppCode,
			AppSecret:   secret,
			CtxInfo:     ctxInfo,
			RedirectURI: "https://evil.example.com/steal",
		})
		if err == nil {
			t.Fatal("expected rejection for disallowed redirect_uri, got nil")
		}
	})
}
