---
name: isme-ui-lint-baseline
description: isme ui lint baseline is 0 errors (PR #16 fixed the 13 legacy errors); set-state-in-effect rule gotcha
metadata:
  type: project
---

isme `ui/` ESLint baseline is **0 errors** (verified 2026-06-08 on main 0776610). The 13 legacy errors noted earlier were fixed by PR #16 — that cleanup IS on main now.

**Why:** the repo treats lint-clean as a hard verification gate; any new error is a regression.

**How to apply:** verification target for all UI work is **0 total errors** (`cd ui && npm run lint`). The config enforces `react-hooks/set-state-in-effect` — calling setState synchronously in a useEffect body fails lint. Derive state during render or keep stale data visible during refetches; setState inside async callbacks (setTimeout/promise `.then`, async code after an `await`) is fine — e.g. UserProvider's fetchUser starts with `await Promise.resolve()` to push its setState calls into promise context. Chakra snippet files that must co-export hooks with components use file-level `/* eslint-disable react-refresh/only-export-components */`.
