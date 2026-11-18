package history

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
)

// BaselineMigration is the squashed Postgres snapshot of the whole schema (all
// 12 application tables + their indexes) plus the seed data the old incremental
// history embedded (RBAC roles/permissions/grants, the isme self-app_service row
// and the schedule_config job rows). It is the first entry in Migrations, so a
// fresh database needs nothing but `up`.
//
// ⚠️ FROZEN — never edit this migration again. Production recorded it as applied
// (id 1) and will never re-run it, so a change here only reaches fresh installs
// and makes them drift from production. Every schema or seed change is a new
// migration appended to the tail of Migrations instead (e.g. 035 removes a seed
// row this baseline still inserts).
//
// INTEGER flags whose Go entity field is a bool (is_verified / is_system /
// enabled) are native BOOLEAN so bun's pgdialect round-trips them; JSON-as-TEXT
// columns such as app_services.redirect_urls stay TEXT; the user_sessions time
// columns are TIMESTAMPTZ to match the time.Time entity fields.
var BaselineMigration = pkgMigrate.Migration{
	Name: "000_baseline",
	Up:   baselineUp,
	Down: baselineDown,
}

// baselinePostgresStatements returns the full ordered DDL statement list.
// Postgres executes statements strictly in order and rejects a CREATE INDEX whose
// table does not exist yet, or an FK to a table not yet created, so ALL CREATE
// TABLE statements come first (user_invitations before user_invitation_roles),
// then ALL CREATE INDEX statements.
func baselinePostgresStatements() []string {
	return []string{
		// --- Phase 1: tables (FK-safe order) ---
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			email TEXT,
			password TEXT NOT NULL,
			status INTEGER NOT NULL,
			last_login_at TIMESTAMPTZ DEFAULT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at TIMESTAMPTZ,
			deleted_by TEXT DEFAULT '',
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			avatar_url TEXT
		)`,
		// user_sessions: expires_at / last_login_at / created_at map to time.Time
		// and last_refreshed_at to *time.Time — all TIMESTAMPTZ so bun
		// round-trips them.
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			email TEXT NOT NULL,
			refresh_token TEXT,
			expires_at TIMESTAMPTZ,
			last_login_at TIMESTAMPTZ,
			status INTEGER DEFAULT 1,
			client_ip TEXT NOT NULL,
			user_agent TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
			token_id TEXT NOT NULL DEFAULT '',
			app_service_id TEXT NOT NULL DEFAULT '',
			refresh_count INTEGER NOT NULL DEFAULT 0,
			last_refreshed_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS app_services (
			id TEXT PRIMARY KEY NOT NULL,
			app_code TEXT UNIQUE NOT NULL,
			app_name TEXT NOT NULL,
			app_secret TEXT NOT NULL,
			redirect_url TEXT NOT NULL,
			ctx_info TEXT NOT NULL,
			status INTEGER DEFAULT 1 NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at TIMESTAMPTZ,
			deleted_by TEXT DEFAULT '',
			icon TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT '',
			redirect_urls TEXT NOT NULL DEFAULT '[]'
		)`,
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id TEXT NOT NULL,
			permission_id BIGINT NOT NULL,
			PRIMARY KEY (role_id, permission_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			app_service_id TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			app_id TEXT NOT NULL DEFAULT 'app_isme',
			resource TEXT NOT NULL,
			action TEXT NOT NULL,
			icon TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT '',
			UNIQUE (app_id, resource, action)
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id TEXT PRIMARY KEY NOT NULL,
			app_id TEXT NOT NULL DEFAULT 'app_isme',
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			is_system BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at TIMESTAMPTZ,
			deleted_by TEXT DEFAULT '',
			icon TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT '',
			UNIQUE (app_id, code)
		)`,
		// user_invitations must precede user_invitation_roles (FK target).
		`CREATE TABLE IF NOT EXISTS user_invitations (
			id TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL,
			role_id TEXT,
			token_hash TEXT NOT NULL,
			status INTEGER NOT NULL DEFAULT 1,
			expires_at TIMESTAMPTZ NOT NULL,
			accepted_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at TIMESTAMPTZ,
			deleted_by TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS user_invitation_roles (
			id TEXT PRIMARY KEY NOT NULL,
			invitation_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			app_service_id TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at TIMESTAMPTZ,
			deleted_by TEXT DEFAULT '',
			FOREIGN KEY (invitation_id) REFERENCES user_invitations (id)
		)`,
		`CREATE TABLE IF NOT EXISTS token_rotation_events (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			rotated_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
		)`,
		`CREATE TABLE IF NOT EXISTS schedule_config (
			job_key TEXT PRIMARY KEY,
			enabled BOOLEAN NOT NULL DEFAULT FALSE,
			cron TEXT NOT NULL,
			params TEXT NOT NULL DEFAULT '{}',
			last_run_at TIMESTAMPTZ,
			last_result TEXT,
			updated_at TIMESTAMPTZ,
			updated_by TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS activity_events (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			meta TEXT NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		// --- Phase 2: indexes (every referenced table now exists) ---
		`CREATE UNIQUE INDEX IF NOT EXISTS users_email_live_uidx ON users (email) WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS user_sessions_refresh_token_idx ON user_sessions (refresh_token)`,
		`CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions (user_id)`,
		`CREATE INDEX IF NOT EXISTS app_services_app_code_idx ON app_services (app_code)`,
		`CREATE INDEX IF NOT EXISTS user_roles_user_id_idx ON user_roles (user_id)`,
		`CREATE INDEX IF NOT EXISTS user_roles_role_id_idx ON user_roles (role_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS user_roles_user_role_app_uidx ON user_roles (user_id, role_id, COALESCE(app_service_id, ''))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS permissions_app_resource_action_uidx ON permissions (app_id, resource, action)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS roles_app_code_uidx ON roles (app_id, code)`,
		`CREATE INDEX IF NOT EXISTS permissions_app_id_idx ON permissions (app_id)`,
		`CREATE INDEX IF NOT EXISTS roles_app_id_idx ON roles (app_id)`,
		`CREATE INDEX IF NOT EXISTS user_invitation_roles_invitation_id_idx ON user_invitation_roles (invitation_id)`,
		`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`,
		`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS token_rotation_events_user_rotated_idx ON token_rotation_events (user_id, rotated_at)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_events_user_created ON activity_events (user_id, created_at)`,
	}
}

// baselineSystemRoles is the RBAC seed catalog from migration 010 (final-state app_id).
var baselineSystemRoles = []struct {
	id          string
	code        string
	name        string
	description string
}{
	{"rol_admin", "admin", "Admin", "Full access to every resource"},
	{"rol_member", "member", "Member", "Read access to core resources"},
	{"rol_viewer", "viewer", "Viewer", "Read-only access to core resources"},
}

// baselinePermissionCatalog is the final permission catalog: the 18 base
// permissions from migration 010, the user:verify permission from migration 012,
// then the 2 settings permissions from migration 022, in their original order.
var baselinePermissionCatalog = []struct {
	resource string
	action   string
}{
	{"user", "read"},
	{"user", "create"},
	{"user", "update"},
	{"user", "delete"},
	{"user", "reset_password"},
	{"user_session", "read"},
	{"user_session", "delete"},
	{"user_session", "revoke"},
	{"app_service", "read"},
	{"app_service", "create"},
	{"app_service", "update"},
	{"app_service", "delete"},
	{"app_service", "rotate_secret"},
	{"role", "read"},
	{"role", "create"},
	{"role", "update"},
	{"role", "delete"},
	{"role", "assign"},
	{"user", "verify"},
	{"settings", "read"},
	{"settings", "update"},
}

// baselineReadOnlyCodes are the core read permissions granted to the member and
// viewer system roles (migration 010).
var baselineReadOnlyCodes = []struct {
	resource string
	action   string
}{
	{"user", "read"},
	{"user_session", "read"},
	{"app_service", "read"},
	{"role", "read"},
}

// baselineScheduleJobs is the set of schedule_config rows the baseline seeds,
// all disabled. database_backup is still seeded here because this migration is
// frozen; migration 035 deletes that row on every database.
var baselineScheduleJobs = []struct {
	jobKey string
	cron   string
	params string
}{
	{"session_revoke", "0 3 * * *", "{}"},
	{"rotation_cleanup", "0 4 * * *", `{"retention_hours":48}`},
	{"activity_cleanup", "0 5 * * *", `{"retention_days":90}`},
	{"database_backup", "0 3 * * *", `{"retain_count":10}`},
	{"invitation_cleanup", "0 6 * * *", "{}"},
	{"unverified_user_cleanup", "0 7 * * *", `{"min_age_days":30}`},
}

// baselineSeed inserts the seed data. Grants reference permissions by
// (resource, action) subquery so they never depend on generated ids.
func baselineSeed(ctx context.Context, db bun.IDB) error {
	roleSQL := `INSERT INTO roles (id, code, name, description, is_system) VALUES (?, ?, ?, ?, TRUE) ON CONFLICT (id) DO NOTHING`
	for _, role := range baselineSystemRoles {
		if _, err := db.ExecContext(ctx, roleSQL, role.id, role.code, role.name, role.description); err != nil {
			return fmt.Errorf("baseline seed role %s: %w", role.id, err)
		}
	}

	permissionSQL := `INSERT INTO permissions (resource, action) VALUES (?, ?) ON CONFLICT (app_id, resource, action) DO NOTHING`
	for _, permission := range baselinePermissionCatalog {
		if _, err := db.ExecContext(ctx, permissionSQL, permission.resource, permission.action); err != nil {
			return fmt.Errorf("baseline seed permission %s:%s: %w", permission.resource, permission.action, err)
		}
	}

	// grants — admin holds the full catalog; member and viewer hold core reads.
	grantSQL := `INSERT INTO role_permissions (role_id, permission_id)
		SELECT ?, id FROM permissions WHERE resource = ? AND action = ?
		ON CONFLICT (role_id, permission_id) DO NOTHING`
	for _, permission := range baselinePermissionCatalog {
		if _, err := db.ExecContext(ctx, grantSQL, "rol_admin", permission.resource, permission.action); err != nil {
			return fmt.Errorf("baseline grant admin %s:%s: %w", permission.resource, permission.action, err)
		}
	}
	for _, permission := range baselineReadOnlyCodes {
		if _, err := db.ExecContext(ctx, grantSQL, "rol_member", permission.resource, permission.action); err != nil {
			return fmt.Errorf("baseline grant member %s:%s: %w", permission.resource, permission.action, err)
		}
		if _, err := db.ExecContext(ctx, grantSQL, "rol_viewer", permission.resource, permission.action); err != nil {
			return fmt.Errorf("baseline grant viewer %s:%s: %w", permission.resource, permission.action, err)
		}
	}

	appSQL := `INSERT INTO app_services
		(id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
		VALUES ('app_isme', 'isme', 'ISME', '', '', 'authen', 1)
		ON CONFLICT (id) DO NOTHING`
	if _, err := db.ExecContext(ctx, appSQL); err != nil {
		return fmt.Errorf("baseline seed app_isme: %w", err)
	}

	scheduleSQL := `INSERT INTO schedule_config (job_key, enabled, cron, params) VALUES (?, FALSE, ?, ?) ON CONFLICT (job_key) DO NOTHING`
	for _, job := range baselineScheduleJobs {
		if _, err := db.ExecContext(ctx, scheduleSQL, job.jobKey, job.cron, job.params); err != nil {
			return fmt.Errorf("baseline seed schedule %s: %w", job.jobKey, err)
		}
	}

	return nil
}

// baselineUp creates the schema and seeds it. The legacy no-op migrations that
// follow it in Migrations are recorded by the runner itself.
func baselineUp(db bun.IDB) error {
	ctx := context.Background()

	for _, statement := range baselinePostgresStatements() {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("baseline up failed on statement: %w", err)
		}
	}

	return baselineSeed(ctx, db)
}

// baselineDown drops every application table created by Up. The `migrations`
// bookkeeping rows are owned by the runner and removed when the baseline's own
// row is rolled back.
func baselineDown(db bun.IDB) error {
	ctx := context.Background()

	tables := []string{
		"user_invitation_roles",
		"user_invitations",
		"user_roles",
		"role_permissions",
		"token_rotation_events",
		"activity_events",
		"schedule_config",
		"permissions",
		"roles",
		"app_services",
		"user_sessions",
		"users",
	}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS `+table); err != nil {
			return fmt.Errorf("baseline down failed to drop %s: %w", table, err)
		}
	}

	return nil
}
