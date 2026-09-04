# Memory Index

- [isme UI lint baseline](project_isme-ui-lint-baseline.md) — baseline now 0 errors (legacy 13 fixed); set-state-in-effect rule gotcha

## Moved in from the platform-level store (2026-09-05)

⚠️ These were written while the session's working directory was the platform
root, so they landed in `<root>/.claude/agent-memory/` where this repo's own
agent could not see them — same knowledge, same agent, different cwd. They live
here now, and new ones belong here.

- [isme invite flow](project_isme_invite_flow.md) — signup REMOVED 2026-06-08, invite links only; Signup/ForgotPassword files kept on disk; 7d TTL, expired derived
- [isme RBAC backend](project_isme_rbac_backend.md) — RBAC landed 2026-06-06; kuery v1.15.0 tagged; isme /auth/me dropped is_admin (adm in JWT claims) — consumer shims must parse the token
- [isme Users screen stubbed](project_isme_users_screen_stubbed.md) — /users UI wired up; only resetUserPassword still a stub (email infra)
