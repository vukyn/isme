---
name: isme-vite8-deferred-upgrade
description: "isme form-data CRLF fix shipped PR#58; 3 esbuild dev-only highs remain, need Vite 8 (Rolldown) — deferred; Vite 8 facts"
metadata: 
  node_type: memory
  type: project
---

isme `form-data` CRLF-injection fix (GHSA-hmw2-7cc7-3qxx, CVSS 8.7) SHIPPED — PR#58 squash-merged to main (`656b5db`), 2026-06-16. Bumped form-data 4.0.5→4.0.6 + dev-only js-yaml/@babel/core; only `ui/package-lock.json` changed. `npm audit --omit=dev` (prod tree) = 0 vulns.

**Remaining (deferred):** 3 esbuild dev-only HIGH advisories (GHSA-gv7w-rqvm-qjhr SSRF) in isme/ui. Dev-server only, NOT shipped to prod → low real risk. Only fix = Vite 8 (drops standalone esbuild). isme currently Vite `^7.2.4` resolved 7.3.5.

**Why deferred:** prod clean already; rainy was mid-coding so didn't batch. User said record-only, no migration now (2026-06-16).

**Vite 8 facts (web-verified, released ~May 2026):**
- Node req unchanged: 20.19+ / 22.12+ (isme fine, no bump).
- Rolldown (Rust bundler) replaces esbuild+Rollup → unified dev/prod, ~30x faster builds, removes the esbuild CVE by removal.
- Breaking: `build.rollupOptions`→`build.rolldownOptions`; default target → `baseline-widely-available` (chrome111/edge111/firefox114/safari16.4); removed `import.meta.hot.accept` fallback; CJS interop changes; Yarn PnP incompatible (isme=npm, N/A).
- Recommended path: 2-step — on Vite 7 swap `vite`→`rolldown-vite`, test, then bump to v8.

Same form-data High also in rainy: SHIPPED — PR#84 squash-merged to main (`b1640d6`), 2026-06-16, only `ui/package-lock.json` (form-data 4.0.5→4.0.6, prod audit clean); db WAL artifacts left untouched (committed lockfile-only off fresh branch despite mid-coding dirty tree). rainy also has same 3 esbuild dev-only highs (Vite 8 deferred). See [[open-followups-2026-06]].
