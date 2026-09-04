---
name: isme-ui-serving
description: "isme serves its own embedded SPA: the SSO popup loads the built bundle (not Vite), and root files like favicon.svg need an explicit route before the SPA catch-all."
metadata:
  type: project
---

Several memories merged into one topic file to keep the index readable. Each section below is the original entry, unedited.

## isme-sso-popup-embedded-ui

The medioa2 SSO popup opens isme at `http://sso.isme.local:8081/sso/login` (medioa2 `.env` `AUTH_ENDPOINT=http://sso.isme.local:8081`, `AUTH_ENDPOINT_WEB_SSO_LOGIN="/sso/login"`). That serves isme's built SPA from `internal/ui`, NOT the Vite dev server on :5173.

**HISTORY (2026-06-20):** isme/medioa2 ORIGINALLY served `internal/ui` at RUNTIME via `html.New("internal/ui")` + `http.Dir` (working-dir-relative, NOT go:embed) — an earlier version of this memory wrongly claimed compile-time embed. Later the SAME DAY all 3 services (rainy/medioa2/isme) were migrated TO go:embed — see [[rainy-embed-ui]]. So isme NOW embeds; the runtime-serve description above is historical.

**Why the stale-UI trap persists either way:** Fiber's html template engine loads/parses templates at server STARTUP, so a fresh `internal/ui` on disk isn't picked up until restart. During SSO UI work, fidelity fixes kept appearing "not applied" because the running isme server had stale templates loaded.

**How to apply:** when iterating on isme SSO UI and testing via the real medioa2 popup (:8081), after editing `ui/src/**` you must `make build-web` AND restart the isme server (kill the :8081 listener, re-`make run`) for changes to show. Vite (:5173) HMR is live but the popup never hits it. Don't burn turns debugging a "mismatch" that is really a stale embed — rebuild+restart first, then compare. To ground-truth a demo/ mock, render it with headless Chrome: `"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu --force-device-scale-factor=2 --window-size=620,1400 --screenshot=/tmp/x.png "file://<abs>/demo/<mock>.html"`.

Related: [[demo-mock-source-of-truth]]

## isme-ui-static-serving

isme `internal/server/server.go` serves the embedded SPA (`internal/ui/`, via `html.New` template engine + `filesystem.New` mounted at `/assets` only). The web catch-all is `app.Get("/*", renderHomePage)` which renders `index.html` for ANY unmatched path.

Consequence: root-level static files referenced by index.html (e.g. `/favicon.svg`, living at `internal/ui/favicon.svg`) are NOT under `/assets`, so the catch-all returns HTML for them → browser gets `text/html` instead of the asset → favicon falls back to the generic globe.

**Why:** `/assets` static mount doesn't cover root paths; the SPA fallback masks the 404 with index.html.

**How to apply:** any root static file index.html references needs an explicit Fiber route BEFORE `app.Get("/*", ...)`. Fixed for favicon (PR #53, 2026-06-11):
```go
app.Get("/favicon.svg", func(c *fiber.Ctx) error { return c.SendFile("internal/ui/favicon.svg") })
```
`SendFile` sets content-type from extension. Paths are cwd-relative (server run from repo root, same as `html.New("internal/ui", ...)`). The isme brand logo = the favicon i/person monogram — see [[demo-mock-source-of-truth]]. UI changes need build-web + restart to show ([[isme-ui-serving]]).
