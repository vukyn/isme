---
name: project-isme-invite-flow
description: isme replaces public signup with admin invite-link flow (new user_invitation domain); aurora mock by designer in parallel; UI invite dialog was a stub
metadata:
  type: project
---

isme is replacing public signup with an admin-driven invite-link flow (approved design, planned 2026-06-08): new `user_invitation` domain, POST/GET /users/invites + public /auth/invites/:token + /auth/accept-invite, signup backend deleted. Signup.tsx / ForgotPassword.tsx stay on disk deliberately (routes removed only).

**Why:** SSO service should not allow self-registration; accounts are provisioned by admins. No email infra exists, so the invite link is shown once at creation (copy button) instead of being emailed.

**How to apply:** UI already had a stubbed InviteUserDialog + `inviteUser`/`resendUserInvite` stubs in apis/user.ts with stale semantics (72h copy, pending user status 3, is_admin checkbox) — invite work builds on/reworks those, not greenfield. A designer is producing an aurora-theme invite mock in isme/demo/ in parallel; coder ports it. Note isme uses kuery pkgErr everywhere — no exceptions/ dirs exist despite the platform template.
