---
name: isme-app-multi-redirect
description: isme app_service multiple allowed redirect URLs (OAuth-style allowlist) — design + SSO session shape
metadata: 
  node_type: memory
  type: project
---

isme app_service supports an allowlist of additional redirect URLs (OAuth-style), SHIPPED PR #80 (main 444f293) + migration 031 applied prod Neon.

- Schema: `redirect_urls TEXT NOT NULL DEFAULT '[]'` (JSON array of ADDITIONAL urls); `redirect_url` stays the PRIMARY/default. Migration `031_add_redirect_urls_to_app_services` (portable TEXT, no dialect branch) + baseline both branches. **Cap = 3** additional URLs; primary need NOT be in the list (union formed at match time).
- SSO flow (auth usecase): `RequestLogin` accepts optional `redirect_uri` → validate by STRICT exact-match (trim only, no canonicalize/wildcard — open-redirect guard) against {redirect_url} ∪ redirect_urls → freeze chosen into the session. **Session cache value changed** from bare appServiceID string → JSON `{app_service_id, redirect_url}` with legacy bare-id fallback (`sso_session.go`: encode/decode/allowedRedirectURLs/chooseRedirectURL). Consent/Login use ONLY the frozen redirect, never re-derive from client input. No redirect_uri → primary default.
- Backward-compat (original ship): medioa2/rainy sent no `redirect_uri` → default path; `external/auth` SDK `RequestLoginRequest` had NO redirect_uri field.
- **UPDATE 2026-06-24 — origin-aware redirect wired (the SDK was the missing link).** Symptom: login from a non-primary origin (e.g. `app.rainy.local:8083`, `app.medioa.local:8082`) always bounced to the PRIMARY (`*.fly.dev`) because (a) the published `external/auth` SDK struct lacked `RedirectURI` so no client could send it, and (b) apps sent an empty login body. Fix = 3 layers: SDK adds `RedirectURI string json:"redirect_uri,omitempty"` (tagged **isme v1.0.0-dev.10**, additive/backward-safe) → app FE sends `window.location.origin + "/auth/callback"` → threaded through app `ExternalLoginRequest`→usecase→`RequestLoginRequest.RedirectURI`. Shipped rainy PR#222, medioa2 PR#69 (medioa2 dev.9→dev.10 transitively bumped kuery v1.19→v1.36, scan clean). **Prereq: each origin's exact `…/auth/callback` string must be in that app_service `redirect_urls` allowlist (exact byte-match — scheme/port/no-trailing-slash all matter); prod is safe when prod callback = primary, only local origins need adding.** Empty redirect_uri still → primary (safe). Any NEW service consuming isme SSO should send origin+/auth/callback from the start.
- FE: add/remove rows (cap 3) in RegisterAppServiceDialog + EditAppService; "+N more" badge in AppServices list. Matches aurora demo mocks (edit/list/detail updated, max 3).
- **Runtime remaining: deploy isme to fly** (FE+BE binary) — migration already applied prod.

See [[isme-auth-topology]], [[rainy-sso-bare-origin-redirect-bug]], [[demo-mock-source-of-truth]], [[postgres-sqlite-gotchas]] (kuery v1.36.0 runner fix lets 031 apply after baseline).
