package history

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
)

// BaselineMigration is a squashed, dual-dialect (SQLite + Postgres) snapshot of
// the entire final schema (all 12 application tables + their indexes) plus the
// migration-embedded seed data (RBAC roles/permissions/grants, the isme
// self-app_service row, and the four schedule_config job rows), used as the
// fresh-install path for a brand-new database on either dialect.
//
// It is intentionally NOT registered in the Migrations slice in migrations.go —
// the incremental 001-029 set is left byte-identical so existing dev/prod SQLite
// databases keep migrating step-by-step. BaselineMigration is run explicitly via
// `go run db/migrate.go <sqlite|postgres> baseline`.
//
// The SQLite branch is transcribed verbatim from the authoritative live schema
// dump (`sqlite3 db/app.db .schema`); the Postgres branch applies the documented
// dialect translation rules (DATETIME -> TIMESTAMPTZ, INTEGER PK AUTOINCREMENT ->
// BIGINT GENERATED ALWAYS AS IDENTITY, IFNULL -> COALESCE, INSERT OR IGNORE ->
// ON CONFLICT DO NOTHING; INTEGER flags whose Go entity field is a bool
// — is_verified / is_system / enabled — become native BOOLEAN so bun's pgdialect
// (which emits TRUE/FALSE for Go bool) round-trips them; JSON-as-TEXT columns
// such as app_services.redirect_urls stay as-is — a TEXT JSON array defaulting
// to '[]' on both dialects).
// The user_sessions time columns are declared TEXT/TIMESTAMP in SQLite but the
// Go entity fields are time.Time, so the Postgres branch declares them
// TIMESTAMPTZ for correct bun round-tripping.
//
// After creating the schema and seeding, Up stamps every name in the 001-029
// history as already applied in the `migrations` bookkeeping table, so a
// subsequent `up` against the incremental set is a clean no-op rather than
// re-running ALTERs that would fail with "column already exists".
var BaselineMigration = pkgMigrate.Migration{
	Name: "000_baseline",
	Up:   baselineUp,
	Down: baselineDown,
}

// baselineSQLiteStatements returns the full ordered DDL statement list for the
// SQLite dialect, transcribed verbatim from the live schema dump.
func baselineSQLiteStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			password TEXT NOT NULL,
			status INTEGER NOT NULL,
			last_login_at DATETIME DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at DATETIME,
			deleted_by TEXT DEFAULT '',
			is_verified INTEGER NOT NULL DEFAULT 0,
			avatar_url TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			email TEXT NOT NULL,
			refresh_token TEXT,
			expires_at TEXT,
			last_login_at TEXT,
			status INTEGER DEFAULT 1,
			client_ip TEXT NOT NULL,
			user_agent TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
			token_id TEXT NOT NULL DEFAULT '',
			app_service_id TEXT NOT NULL DEFAULT '',
			refresh_count INTEGER NOT NULL DEFAULT 0,
			last_refreshed_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS user_sessions_refresh_token_idx ON user_sessions (refresh_token)`,
		`CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions (user_id)`,
		`CREATE TABLE IF NOT EXISTS app_services (
			id TEXT PRIMARY KEY NOT NULL,
			app_code TEXT UNIQUE NOT NULL,
			app_name TEXT NOT NULL,
			app_secret TEXT NOT NULL,
			redirect_url TEXT NOT NULL,
			ctx_info TEXT NOT NULL,
			status INTEGER DEFAULT 1 NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at DATETIME,
			deleted_by TEXT DEFAULT '',
			icon TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT '',
			redirect_urls TEXT NOT NULL DEFAULT '[]'
		)`,
		`CREATE INDEX IF NOT EXISTS app_services_app_code_idx ON app_services (app_code)`,
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id TEXT NOT NULL,
			permission_id INTEGER NOT NULL,
			PRIMARY KEY (role_id, permission_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			app_service_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS user_roles_user_id_idx ON user_roles (user_id)`,
		`CREATE INDEX IF NOT EXISTS user_roles_role_id_idx ON user_roles (role_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS user_roles_user_role_app_uidx ON user_roles (user_id, role_id, IFNULL(app_service_id, ''))`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			is_system INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at DATETIME,
			deleted_by TEXT DEFAULT '',
			icon TEXT NOT NULL DEFAULT '',
			color TEXT NOT NULL DEFAULT '',
			UNIQUE (app_id, code)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS permissions_app_resource_action_uidx ON permissions (app_id, resource, action)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS roles_app_code_uidx ON roles (app_id, code)`,
		`CREATE INDEX IF NOT EXISTS permissions_app_id_idx ON permissions (app_id)`,
		`CREATE INDEX IF NOT EXISTS roles_app_id_idx ON roles (app_id)`,
		`CREATE TABLE IF NOT EXISTS user_invitation_roles (
			id TEXT PRIMARY KEY NOT NULL,
			invitation_id TEXT NOT NULL,
			role_id TEXT NOT NULL,
			app_service_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at DATETIME,
			deleted_by TEXT DEFAULT '',
			FOREIGN KEY (invitation_id) REFERENCES user_invitations (id)
		)`,
		`CREATE INDEX IF NOT EXISTS user_invitation_roles_invitation_id_idx ON user_invitation_roles (invitation_id)`,
		`CREATE TABLE IF NOT EXISTS user_invitations (
			id TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL,
			role_id TEXT,
			token_hash TEXT NOT NULL,
			status INTEGER NOT NULL DEFAULT 1,
			expires_at DATETIME NOT NULL,
			accepted_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at DATETIME,
			deleted_by TEXT DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`,
		`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL`,
		`CREATE TABLE IF NOT EXISTS token_rotation_events (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			rotated_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
		)`,
		`CREATE INDEX IF NOT EXISTS token_rotation_events_user_rotated_idx ON token_rotation_events (user_id, rotated_at)`,
		`CREATE TABLE IF NOT EXISTS schedule_config (
			job_key TEXT PRIMARY KEY,
			enabled INTEGER NOT NULL DEFAULT 0,
			cron TEXT NOT NULL,
			params TEXT NOT NULL DEFAULT '{}',
			last_run_at DATETIME,
			last_result TEXT,
			updated_at DATETIME,
			updated_by TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS activity_events (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			type TEXT NOT NULL,
			meta TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_events_user_created ON activity_events (user_id, created_at)`,
	}
}

// baselinePostgresStatements returns the full ordered DDL statement list for the
// Postgres dialect, applying the documented translation rules.
//
// Unlike the SQLite branch (transcribed verbatim from a dependency-valid
// `.schema` dump), Postgres executes statements strictly in order and rejects a
// CREATE INDEX whose table does not yet exist, or a CREATE TABLE whose FK
// references a not-yet-created table. So this branch is split into two ordered
// phases: ALL CREATE TABLE statements first (with user_invitations created
// before user_invitation_roles, which has an FK to it), then ALL CREATE INDEX
// statements. The resulting schema is semantically identical to the SQLite
// branch — only the emission order differs.
func baselinePostgresStatements() []string {
	return []string{
		// --- Phase 1: tables (FK-safe order) ---
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
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
		// user_sessions: expires_at / last_login_at / created_at are declared
		// TEXT in SQLite but the Go entity fields are time.Time, and
		// last_refreshed_at is a *time.Time — all declared TIMESTAMPTZ here so
		// bun round-trips them correctly on Postgres.
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

// systemRoles is the RBAC seed catalog from migration 010 (final-state app_id).
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
// then the 2 settings permissions from migration 022. Order matches the
// step-by-step `up` so SQLite AUTOINCREMENT ids line up identically.
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

// baselineScheduleJobs is the final set of schedule_config rows after migrations
// 025 (session_revoke + rotation_cleanup), 027 (activity_cleanup) and 029
// (database_backup). All disabled by default.
var baselineScheduleJobs = []struct {
	jobKey string
	cron   string
	params string
}{
	{"session_revoke", "0 3 * * *", "{}"},
	{"rotation_cleanup", "0 4 * * *", `{"retention_hours":48}`},
	{"activity_cleanup", "0 5 * * *", `{"retention_days":90}`},
	{"database_backup", "0 3 * * *", `{"retain_count":10}`},
}

// baselineSeed reproduces the migration-embedded seed data (010/014/022/025/
// 027/029) in their final shape, dialect-aware (SQLite INSERT OR IGNORE vs
// Postgres ON CONFLICT DO NOTHING). Grants reference permissions by
// (resource, action) subquery so they are id-agnostic across dialects.
func baselineSeed(ctx context.Context, db bun.IDB) error {
	pg := isPostgres(db)

	// roles
	roleSQL := `INSERT OR IGNORE INTO roles (id, code, name, description, is_system) VALUES (?, ?, ?, ?, 1)`
	if pg {
		roleSQL = `INSERT INTO roles (id, code, name, description, is_system) VALUES (?, ?, ?, ?, TRUE) ON CONFLICT (id) DO NOTHING`
	}
	for _, role := range baselineSystemRoles {
		if _, err := db.ExecContext(ctx, roleSQL, role.id, role.code, role.name, role.description); err != nil {
			return fmt.Errorf("baseline seed role %s: %w", role.id, err)
		}
	}

	// permissions
	permSQL := `INSERT OR IGNORE INTO permissions (resource, action) VALUES (?, ?)`
	if pg {
		permSQL = `INSERT INTO permissions (resource, action) VALUES (?, ?) ON CONFLICT (app_id, resource, action) DO NOTHING`
	}
	for _, permission := range baselinePermissionCatalog {
		if _, err := db.ExecContext(ctx, permSQL, permission.resource, permission.action); err != nil {
			return fmt.Errorf("baseline seed permission %s:%s: %w", permission.resource, permission.action, err)
		}
	}

	// grants — admin holds the full catalog; member and viewer hold core reads.
	grantSQL := `INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
		SELECT ?, id FROM permissions WHERE resource = ? AND action = ?`
	if pg {
		grantSQL = `INSERT INTO role_permissions (role_id, permission_id)
			SELECT ?, id FROM permissions WHERE resource = ? AND action = ?
			ON CONFLICT (role_id, permission_id) DO NOTHING`
	}
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

	// isme self-app_service row (migration 014)
	appSQL := `INSERT OR IGNORE INTO app_services
		(id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
		VALUES ('app_isme', 'isme', 'ISME', '', '', 'authen', 1)`
	if pg {
		appSQL = `INSERT INTO app_services
			(id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
			VALUES ('app_isme', 'isme', 'ISME', '', '', 'authen', 1)
			ON CONFLICT (id) DO NOTHING`
	}
	if _, err := db.ExecContext(ctx, appSQL); err != nil {
		return fmt.Errorf("baseline seed app_isme: %w", err)
	}

	// schedule_config job rows (migrations 025/027/029)
	scheduleSQL := `INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params) VALUES (?, 0, ?, ?)`
	if pg {
		scheduleSQL = `INSERT INTO schedule_config (job_key, enabled, cron, params) VALUES (?, FALSE, ?, ?) ON CONFLICT (job_key) DO NOTHING`
	}
	for _, job := range baselineScheduleJobs {
		if _, err := db.ExecContext(ctx, scheduleSQL, job.jobKey, job.cron, job.params); err != nil {
			return fmt.Errorf("baseline seed schedule %s: %w", job.jobKey, err)
		}
	}

	return nil
}

// baselineUp creates the entire schema per dialect, seeds the migration-embedded
// data, then stamps the incremental 001-029 migrations as already applied so a
// subsequent `up` is a no-op.
func baselineUp(db bun.IDB) error {
	ctx := context.Background()

	statements := baselineSQLiteStatements()
	if isPostgres(db) {
		statements = baselinePostgresStatements()
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("baseline up failed on statement: %w", err)
		}
	}

	if err := baselineSeed(ctx, db); err != nil {
		return err
	}

	return stampIncrementalMigrations(ctx, db)
}

// stampIncrementalMigrations records every name in the 001-029 history slice as
// already applied in the `migrations` bookkeeping table, so the incremental set
// becomes a clean no-op after the baseline runs (and never re-runs an ALTER that
// would crash with "column already exists").
//
// IDs start at 2 because the migration runner inserts the baseline's own
// bookkeeping row at id = 1 on a fresh database (lastMigratedID + 1) after Up
// returns; stamping the incremental names at 2..N leaves id = 1 free for the
// runner and avoids a primary-key collision on Postgres, whose `migrations.id`
// is a plain BIGINT PRIMARY KEY with no server-side default.
func stampIncrementalMigrations(ctx context.Context, db bun.IDB) error {
	stampSQL := `INSERT INTO migrations (id, name, executed_at) VALUES (?, ?, CURRENT_TIMESTAMP) ON CONFLICT (name) DO NOTHING`
	if !isPostgres(db) {
		stampSQL = `INSERT OR IGNORE INTO migrations (id, name, executed_at) VALUES (?, ?, CURRENT_TIMESTAMP)`
	}

	for i, migration := range Migrations {
		id := int64(i) + 2
		if _, err := db.ExecContext(ctx, stampSQL, id, migration.Name); err != nil {
			return fmt.Errorf("baseline up failed to stamp %s as applied: %w", migration.Name, err)
		}
	}

	return nil
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
