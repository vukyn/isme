---
name: spa-catchall-masks-removed-api-route
description: A deleted GET /api/v1/... route returns 200 text/html (SPA catch-all), not 404, in services with app.Get("/*") — check content-type when smoking a removal
metadata:
  type: feedback
---

Smoking a removed endpoint in a service whose `internal/server/server.go` ends with
`app.Get("/*", renderHomePage)` (isme, verified 2026-11-18): an authenticated
`GET /api/v1/settings/database-backup` returned **200 `text/html`** after the route was
deleted; the PUT returned 405. The Fiber v3 move (PR #108) kept the same behavior: a POST
to an unknown `/api/*` path also returns 405, because the `GET /*` route matches the path.
A plan saying "expect 404" is wrong for these services.

**Why:** the SPA catch-all renders index.html for every unmatched GET, including `/api/*`.
A status-code-only smoke reads that as "route still alive".

**How to apply:** when proving an API route is gone, compare content-type against a
live sibling route (`application/json`) or use a non-GET method; report the 200-HTML
as expected rather than as a failure. Related: [[kiokku-health-route-is-api-v1]].
