---
name: open-followups-2026-06
description: "Pending follow-ups after June 2026 security sweep — memz deploy key cutover, rainy lint debt, isme DI flaw"
metadata: 
  node_type: memory
  type: project
---

After the 2026-06-06 security sweep (all 5 repos scanned + fixed, kuery v1.13.0 released, memz migrated Gin→Fiber):

1. **memz deploy is a hard cutover** — merged main replaces MD5(clientId) API keys with random 32-byte keys (SHA-256 stored, constant-time compare). Before `fly deploy`: distribute new plain keys to clients and update the fly.io secret (`make fly-secret KEY=CLIENTID_STRING VALUE=...`). Response envelope also changed to kuery-standard real HTTP status codes (was always-200). Old keys/parsers break on deploy.
2. **rainy ui lint debt — RESOLVED (verified 2026-06-20):** `npm run lint` = **0 errors**, only 2 cosmetic warnings (unused eslint-disable directives in vendored Chakra snippets color-mode.tsx + toaster.tsx, `--fix`-able). The old 11 errors (MediaPlayer exhaustive-deps, useAuth/useTracks no-explicit-any) are gone. No action needed.
3. **isme `internal/di/di.go` — ALREADY FAIL-FAST (verified 2026-06-20):** the swallow described here is gone — `NewEnhancedBuilder` err → `log.Fatal` (di.go:11-13), and each `builder.Add(def)` err → `log.Fatal` (di.go:27-29). No action needed.
4. memz now Fiber v2 (was Gin) with platform-standard kuery http envelope; local `pkg/` deleted; `CLIENTID_STRING` env format is `clientId:apiKeyHex` (64 hex chars) comma-separated.
5. **rainy HLS backfill — TOOLING SHIPPED P4, RUN pending (updated 2026-06-30):** backfill UI+endpoints shipped (see [[rainy-hls-delivery]] P4: `/manage/processing` + Tracks bulk re-segment + `POST /processing/backfill-segments`). REMAINING operator action: actually RUN the bulk backfill of the existing library — requires the LOCAL worker running with `SEGMENT_ENABLED=true` + ffmpeg + prod DB_DSN (fly worker skips, no ffmpeg), then click "Backfill library". Until run, existing un-segmented tracks are not live-eligible under HLS (manifest skips non-ready tracks). Also pending: end-to-end iPhone verify + P5 cutover (flip VITE_LIVE_HLS default true, delete prototype + useLiveListenPlayer).
