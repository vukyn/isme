---
name: demo-mock-source-of-truth
description: demo/ mocks are design source of truth (EXCEPT gardener = Storybook); default theme is PER-SERVICE — isme=aurora, medioa2=cursor
metadata: 
  node_type: memory
  type: project
  modified: 2026-07-27T10:12:07.587Z
---

The `demo/` HTML mocks are the **design source of truth** for every project in pet-platform — **except gardener**, which moved to Storybook on 2026-07-27 (its `demo/` mocks are frozen legacy; see [[gardener-storybook-track]]). **Default theme is PER-SERVICE** (user corrected 2026-06-07): **isme = aurora** (match `isme/demo/aurora-*` mocks, e.g. `aurora-users-design.html`), **medioa2 = cursor** (`medioa2/demo/cursor-design.html`, Tailwind CDN + lucide). New mocks for a service match that service's default theme. Theme names go in the filename prefix (`<theme>-<feature>-design.html`). When building/changing real UI, mirror the active theme's mock styles, components, and patterns. Never delete unported mocks — demo/ may hold more screens/themes than the React UI implements; extras are kept for future planning. (Designer agent `.claude/agents/designer.md` encodes these rules.)

**Why:** User keeps this demo as the canonical design reference; real UIs (e.g. medioa2/ui React+Chakra) should match it.

**How to apply:**
- Screen pattern: each view is `<section class="screen">` + `<template id="X-content">`, mounted via JS `mountShell` + `show(id)`; register new screens in the `PAGES` array + `PAGE_TITLE` map + topbar `data-target` switcher button.
- Modals are full screens (like `upload`, `bucket-create`) rendered as a `gradient-ring` card on a `grid-bg` backdrop; open via `data-go="<id>"`, close via `data-go` back to parent.
- Tokens: `ink-*` grays, `accent-{violet,cyan,green,red,amber}`, `.hairline`, `.mono`, `.pill`, `.focus-ring`, `.btn-cursor`; primary CTA = `bg-gradient-to-r from-accent-violet to-accent-cyan shadow-glow-violet`.
- Known quirk: Tailwind CDN injects utilities AFTER the static `<style>`, so auth screens with `flex`/`grid` on the section override `.screen{display:none}` and bleed through in headless screenshots — force-hide inactive screens when screenshotting only.

**Mock can be richer than real UI:** the mock may show fields the real React UI / BE don't support yet — treat those as *plan for future*, NOT a 1:1 contract. Don't strip mock extras just because real UI lacks them. Example: bucket-create mock has VISIBILITY + DESCRIPTION, but BE bucket create (AWS S3 SDK v2 → Cloudflare R2, `s3.CreateBucketInput`) only supports name + `LocationConstraint` (location hint); R2 ignores bucket ACL on create and S3 has no bucket description — so those two stay mock-only for now. When porting mock→React, only wire fields the BE actually backs.
