---
name: isme-profile-avatar-shipped
description: isme self-service Profile page + avatar upload via medioa shipped; runtime prereq = paste mk_ MEDIOA_API_KEY
metadata: 
  node_type: memory
  type: project
---

isme self-service **Profile** page shipped — MERGED to main 2026-06-14 (PR #55, squash, `59ee0c2`): avatar upload, display-name edit, change-password. Mock = `isme/demo/aurora-profile-design.html`; page = `ui/src/pages/Profile.tsx` at `/profile` (reached via user-chip menu, not primary nav).

Avatar storage decision: isme stores **avatar URL** (not base64). Upload goes server-side through new `internal/domains/media` domain → medioa public-upload API via `kuery/medioa` SDK `Upload()` — mirrors rainy's `media` domain. kuery v1.19.0 already imported (medioa is a sub-package, no go.mod change). isme added MIME allowlist (png/jpeg/webp) + 2MB cap that rainy/medioa lack upstream. Endpoint `POST /api/v1/media/upload` (auth-gated). DB: `avatar_url` column on user entity, migration `028`. Self-update via `PATCH /api/v1/auth/me` (name+avatar); `GetMeResponse.avatar_url` added; `profile_updated` activity type.

**Runtime prereqs before avatar upload works e2e:**
1. Paste real `MEDIOA_API_KEY=mk_...` into `isme/.env` (mint from a medioa2 bucket member). Blank key → DI logs warning, boot succeeds, upload returns 502 "media service is not configured". `.env` gitignored ([[isme-db-gitignored]]).
2. Restart isme server — embedded UI ([[isme-ui-serving]]).

**Verified working e2e 2026-06-14.** Two debugging gotchas hit (both fixed):

1. **medioa bucket = physical R2 bucket per name.** medioa `resolveBucketName` maps the `mk_` key's `bucket_id` → the medioa bucket's NAME → used verbatim as the physical R2 bucket in `PutObject`. The R2 token + `R2_BUCKET_NAME` only provision `rainy`, so a key from any OTHER bucket (e.g. `admin`) → Cloudflare `403 AccessDenied` → medioa 400 → isme 502. Fix: either mint the key from a bucket whose physical R2 bucket exists+writable, or create that R2 bucket + grant the token. (Also hit transient R2 access-key expiry.) isme `media/exceptions.MapMediaError` now forwards medioa's real message on non-sentinel errors (was opaque "media upload failed").

2. **Topbar avatar showed default initials despite GetMe returning avatar_url.** Every page builds `<AppShell user={{name,email}} />` — prop-drilling DROPS `avatar_url` before it reaches Topbar→UserChip. Fix: `UserChip` reads `avatar_url` straight from `useUser()` context (`effectiveAvatar = prop || user?.avatar_url`), bypassing the 10 lossy call sites. Same pattern applies if rainy/medioa2 chips need the live avatar.

Pattern reference for medioa file upload from any service = rainy's `media` domain + `kuery/medioa` SDK. See [[medioa-public-upload-api-plan]].
