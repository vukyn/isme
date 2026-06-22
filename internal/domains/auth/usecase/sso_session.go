package usecase

import (
	"encoding/json"
	"strings"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
)

// ssoSession is the value frozen in the session cache for an in-flight SSO
// handshake. It pins BOTH the requesting app and the redirect chosen (and
// validated) at RequestLogin time, so the later login/consent steps redirect to
// exactly the destination the app asked for — never re-derived from client
// input — closing the open-redirect surface that a wider allowlist would open.
type ssoSession struct {
	AppServiceID string `json:"app_service_id"`
	RedirectURL  string `json:"redirect_url"`
}

// encodeSSOSession serializes the session to the JSON shape stored in the cache.
func encodeSSOSession(session ssoSession) string {
	encoded, err := json.Marshal(session)
	if err != nil {
		// AppServiceID is a plain ULID and RedirectURL is a validated URL, so
		// marshalling cannot realistically fail; degrade to the legacy bare-ID
		// shape so the handshake still resolves.
		return session.AppServiceID
	}
	return string(encoded)
}

// decodeSSOSession parses a cached session value. It is tolerant of the LEGACY
// shape: before this feature the cache stored a bare appServiceID string, so any
// value that does not unmarshal into the JSON envelope is treated as a legacy
// bare appServiceID (RedirectURL empty). This keeps SSO sessions minted before a
// deploy alive across the rollout. The bool reports whether a usable
// AppServiceID was recovered.
func decodeSSOSession(raw string) (ssoSession, bool) {
	if raw == "" {
		return ssoSession{}, false
	}
	var session ssoSession
	if err := json.Unmarshal([]byte(raw), &session); err != nil || session.AppServiceID == "" {
		// legacy bare appServiceID
		return ssoSession{AppServiceID: raw}, true
	}
	return session, true
}

// allowedRedirectURLs builds the exact-match allowlist for an app: its primary
// redirect_url (when non-empty) unioned with the additional redirect_urls. The
// primary need not be present in the stored allowlist — the union is formed
// here at match time.
func allowedRedirectURLs(app appServiceEntity.AppService) []string {
	allowed := make([]string, 0)
	if strings.TrimSpace(app.RedirectURL) != "" {
		allowed = append(allowed, strings.TrimSpace(app.RedirectURL))
	}
	allowed = append(allowed, parseRedirectURLs(app.RedirectURLs)...)
	return allowed
}

// parseRedirectURLs decodes the app_services.redirect_urls JSON-array TEXT into
// a slice. Tolerant: empty/garbage yields an empty (non-nil) slice. Local to the
// auth package so it does not import the app_service usecase.
func parseRedirectURLs(raw string) []string {
	urls := make([]string, 0)
	if raw == "" {
		return urls
	}
	if err := json.Unmarshal([]byte(raw), &urls); err != nil || urls == nil {
		return make([]string, 0)
	}
	return urls
}

// chooseRedirectURL resolves the redirect the SSO flow will use. An empty
// requested URI falls back to the app's primary redirect_url. A non-empty
// requested URI must EXACT-match (after trimming only — no canonicalization,
// trailing-slash folding, prefix or wildcard matching) one of the app's allowed
// URLs. Returns the chosen URL and whether it was allowed.
func chooseRedirectURL(app appServiceEntity.AppService, requested string) (string, bool) {
	trimmed := strings.TrimSpace(requested)
	if trimmed == "" {
		return app.RedirectURL, true
	}
	for _, candidate := range allowedRedirectURLs(app) {
		if candidate == trimmed {
			return trimmed, true
		}
	}
	return "", false
}
