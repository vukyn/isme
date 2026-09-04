---
name: isme-invite-flow
description: isme invite-link user creation landed 2026-06-08 — public signup REMOVED (backend + routes); Signup.tsx/ForgotPassword.tsx kept on disk deliberately; 7-day TTL, status 1/2/3 with expired derived
metadata:
  type: project
---

Invite-link user creation implemented in isme 2026-06-08: new `user_invitation` domain (migration 013, entity/models/repo/usecase/handlers/DI), endpoints `POST|GET /api/v1/users/invites`, `POST /users/invites/:invitationID/revoke` (RBAC user:create/user:read), public `GET /auth/invites/:token` + `POST /auth/accept-invite`. UI: InviteUserDialog reworked to one-time-link pattern (ack-gated Done, no ✕ in success state), /users gained a Users|Invitations sub-tab, new public /accept-invite page.

**Why:** public signup was removed as the design decision — invite links are now the ONLY way to create accounts. Accept marks the user verified immediately (admin vouched). Contract details that override the designer mock's header comments: 7-day TTL (not 72h), status 1=pending/2=accepted/3=revoked with "expired" DERIVED from expires_at (never stored), endpoints under /users/invites + /auth (not /invitations).

**How to apply:** `Signup.tsx`, `ForgotPassword.tsx`, the `signup` api fn/types/schema, and `AUTH_SIGNUP` endpoint const are KEPT on disk deliberately (files must still compile) — only the routes/links were removed; don't "clean them up" without asking. Raw invite token exists only in the create response's `invite_link` (token_hash SHA-256 at rest); the UI prepends `window.location.origin`. Revoked/expired rows get a "New invite" re-issue action that prefills the dialog. The user_invitation routes register BEFORE user routes in server.go so /invites wins over /:userID. Related: [[isme-rbac-backend]].
