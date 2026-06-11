package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/cryp"
)

// Migrations holds all database migrations
var Migrations = []models.Migration{
	{
		Name: "001_create_users_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS users (
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
					deleted_by TEXT DEFAULT ''
				)
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS users`)
			return err
		},
	},
	{
		Name: "002_create_user_sessions_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS user_sessions (
					id TEXT PRIMARY KEY NOT NULL,
					user_id TEXT NOT NULL,
					email TEXT NOT NULL,
					refresh_token TEXT,
					expires_at TEXT,
					last_login_at TEXT,
					status INTEGER DEFAULT 1,
					client_ip TEXT NOT NULL,
					user_agent TEXT,
					created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL
				)
			`)
			if err != nil {
				return err
			}

			// Create indexes
			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_sessions_refresh_token_idx ON user_sessions (refresh_token)`)
			if err != nil {
				return err
			}

			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions (user_id)`)
			if err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS user_sessions`)
			return err
		},
	},
	{
		Name: "003_add_token_id_to_user_sessions",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE user_sessions 
				ADD COLUMN token_id TEXT NOT NULL DEFAULT ''
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE user_sessions 
				DROP COLUMN token_id
			`)
			return err
		},
	},
	{
		Name: "004_create_app_services_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS app_services (
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
					deleted_by TEXT DEFAULT ''
				)
			`)
			if err != nil {
				return err
			}

			// Create index
			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS app_services_app_code_idx ON app_services (app_code)`)
			if err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS app_services`)
			return err
		},
	},
	{
		Name: "005_add_is_admin_to_users",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE users 
				ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE users
				DROP COLUMN is_admin
			`)
			return err
		},
	},
	{
		Name: "006_create_roles_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS roles (
					id TEXT PRIMARY KEY NOT NULL,
					code TEXT UNIQUE NOT NULL,
					name TEXT NOT NULL,
					description TEXT DEFAULT '',
					is_system INTEGER NOT NULL DEFAULT 0,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					created_by TEXT DEFAULT '',
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_by TEXT DEFAULT '',
					deleted_at DATETIME,
					deleted_by TEXT DEFAULT ''
				)
			`)
			if err != nil {
				return err
			}

			// Create index
			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS roles_code_idx ON roles (code)`)
			if err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS roles`)
			return err
		},
	},
	{
		Name: "007_create_permissions_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS permissions (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					resource TEXT NOT NULL,
					action TEXT NOT NULL,
					UNIQUE (resource, action)
				)
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS permissions`)
			return err
		},
	},
	{
		Name: "008_create_role_permissions_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS role_permissions (
					role_id TEXT NOT NULL,
					permission_id INTEGER NOT NULL,
					PRIMARY KEY (role_id, permission_id)
				)
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS role_permissions`)
			return err
		},
	},
	{
		Name: "009_create_user_roles_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS user_roles (
					id TEXT PRIMARY KEY NOT NULL,
					user_id TEXT NOT NULL,
					role_id TEXT NOT NULL,
					app_service_id TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					created_by TEXT DEFAULT ''
				)
			`)
			if err != nil {
				return err
			}

			// Create indexes
			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_roles_user_id_idx ON user_roles (user_id)`)
			if err != nil {
				return err
			}

			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_roles_role_id_idx ON user_roles (role_id)`)
			if err != nil {
				return err
			}

			// One assignment per user/role/scope; NULL scope (global) normalized to ''
			_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS user_roles_user_role_app_uidx ON user_roles (user_id, role_id, IFNULL(app_service_id, ''))`)
			if err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS user_roles`)
			return err
		},
	},
	{
		Name: "010_seed_rbac",
		Up: func(db *bun.DB) error {
			// seed system roles with deterministic IDs
			systemRoles := []struct {
				id          string
				code        string
				name        string
				description string
			}{
				{"rol_admin", "admin", "Admin", "Full access to every resource"},
				{"rol_member", "member", "Member", "Read access to core resources"},
				{"rol_viewer", "viewer", "Viewer", "Read-only access to core resources"},
			}
			for _, role := range systemRoles {
				_, err := db.Exec(`
					INSERT OR IGNORE INTO roles (id, code, name, description, is_system)
					VALUES (?, ?, ?, ?, 1)
				`, role.id, role.code, role.name, role.description)
				if err != nil {
					return err
				}
			}

			// seed the permission catalog
			permissionCatalog := []struct {
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
			}
			permissionIDs := map[string]int64{}
			for _, permission := range permissionCatalog {
				_, err := db.Exec(`INSERT OR IGNORE INTO permissions (resource, action) VALUES (?, ?)`, permission.resource, permission.action)
				if err != nil {
					return err
				}
				var permissionID int64
				row := db.QueryRow(`SELECT id FROM permissions WHERE resource = ? AND action = ?`, permission.resource, permission.action)
				if err := row.Scan(&permissionID); err != nil {
					return err
				}
				permissionIDs[permission.resource+":"+permission.action] = permissionID
			}

			// grants: admin holds the full catalog; member and viewer hold core reads
			grantPermission := func(roleID string, permissionID int64) error {
				_, err := db.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, roleID, permissionID)
				return err
			}
			for _, permission := range permissionCatalog {
				if err := grantPermission("rol_admin", permissionIDs[permission.resource+":"+permission.action]); err != nil {
					return err
				}
			}
			readOnlyCodes := []string{"user:read", "user_session:read", "app_service:read", "role:read"}
			for _, code := range readOnlyCodes {
				if err := grantPermission("rol_member", permissionIDs[code]); err != nil {
					return err
				}
				if err := grantPermission("rol_viewer", permissionIDs[code]); err != nil {
					return err
				}
			}

			// backfill global roles for existing users (new signups get no role by default)
			type userRow struct {
				id      string
				isAdmin bool
			}
			rows, err := db.Query(`SELECT id, is_admin FROM users WHERE deleted_at IS NULL`)
			if err != nil {
				return err
			}
			defer rows.Close()

			userRows := []userRow{}
			for rows.Next() {
				row := userRow{}
				if err := rows.Scan(&row.id, &row.isAdmin); err != nil {
					return err
				}
				userRows = append(userRows, row)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			for _, row := range userRows {
				roleID := "rol_member"
				if row.isAdmin {
					roleID = "rol_admin"
				}
				_, err := db.Exec(`
					INSERT OR IGNORE INTO user_roles (id, user_id, role_id, app_service_id)
					VALUES (?, ?, ?, NULL)
				`, cryp.ULID(), row.id, roleID)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			systemRoleIDs := []string{"rol_admin", "rol_member", "rol_viewer"}
			for _, roleID := range systemRoleIDs {
				if _, err := db.Exec(`DELETE FROM user_roles WHERE role_id = ?`, roleID); err != nil {
					return err
				}
				if _, err := db.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID); err != nil {
					return err
				}
				if _, err := db.Exec(`DELETE FROM roles WHERE id = ?`, roleID); err != nil {
					return err
				}
			}

			// the catalog is seeded exclusively by this migration
			_, err := db.Exec(`DELETE FROM permissions`)
			return err
		},
	},
	{
		Name: "011_add_app_service_id_to_user_sessions",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE user_sessions
				ADD COLUMN app_service_id TEXT NOT NULL DEFAULT ''
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE user_sessions
				DROP COLUMN app_service_id
			`)
			return err
		},
	},
	{
		Name: "012_add_user_verification",
		Up: func(db *bun.DB) error {
			// new signups start unverified (login blocked until verified)
			_, err := db.Exec(`
				ALTER TABLE users
				ADD COLUMN is_verified INTEGER NOT NULL DEFAULT 0
			`)
			if err != nil {
				return err
			}

			// backfill: every existing account is considered verified
			_, err = db.Exec(`UPDATE users SET is_verified = 1`)
			if err != nil {
				return err
			}

			// extend the permission catalog with the verification action
			_, err = db.Exec(`INSERT OR IGNORE INTO permissions (resource, action) VALUES ('user', 'verify')`)
			if err != nil {
				return err
			}
			var permissionID int64
			row := db.QueryRow(`SELECT id FROM permissions WHERE resource = 'user' AND action = 'verify'`)
			if err := row.Scan(&permissionID); err != nil {
				return err
			}

			// admin keeps holding the full catalog
			_, err = db.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES ('rol_admin', ?)`, permissionID)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`
				DELETE FROM role_permissions
				WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'user' AND action = 'verify')
			`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`DELETE FROM permissions WHERE resource = 'user' AND action = 'verify'`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`
				ALTER TABLE users
				DROP COLUMN is_verified
			`)
			return err
		},
	},
	{
		Name: "013_create_user_invitations_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS user_invitations (
					id TEXT PRIMARY KEY NOT NULL,
					email TEXT NOT NULL,
					role_id TEXT NOT NULL,
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
				)
			`)
			if err != nil {
				return err
			}

			// Create indexes
			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`)
			if err != nil {
				return err
			}

			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`)
			if err != nil {
				return err
			}

			// at most one live pending invitation per email
			_, err = db.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx
				ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS user_invitations`)
			return err
		},
	},
	{
		// isme is itself an app_service that owns the original RBAC catalog.
		// The self-app row is a logical owner only — isme never SSO's into
		// itself, so the app_secret is left as a placeholder (decision 1).
		Name: "014_seed_isme_app_service",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				INSERT OR IGNORE INTO app_services
					(id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
				VALUES
					('app_isme', 'isme', 'ISME', '', '', 'authen', 1)
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DELETE FROM app_services WHERE id = 'app_isme'`)
			return err
		},
	},
	{
		// App-owned RBAC: permissions and roles become owned by an app_service.
		// Existing rows default to the isme self-app. The old global UNIQUE
		// constraints are relaxed to be app-scoped via a table rebuild that
		// PRESERVES id values (role_permissions references permission ids, and
		// role ids like 'rol_admin' are referenced by value elsewhere).
		Name: "015_add_app_id_to_rbac",
		Up: func(db *bun.DB) error {
			// add the owning-app column (existing rows -> isme self-app)
			if _, err := db.Exec(`ALTER TABLE permissions ADD COLUMN app_id TEXT NOT NULL DEFAULT 'app_isme'`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE roles ADD COLUMN app_id TEXT NOT NULL DEFAULT 'app_isme'`); err != nil {
				return err
			}

			// rebuild permissions: UNIQUE(resource,action) -> UNIQUE(app_id,resource,action)
			if _, err := db.Exec(`
				CREATE TABLE permissions_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					app_id TEXT NOT NULL DEFAULT 'app_isme',
					resource TEXT NOT NULL,
					action TEXT NOT NULL,
					UNIQUE (app_id, resource, action)
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT INTO permissions_new (id, app_id, resource, action)
				SELECT id, app_id, resource, action FROM permissions
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE permissions`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE permissions_new RENAME TO permissions`); err != nil {
				return err
			}

			// rebuild roles: drop column-level UNIQUE on code -> UNIQUE(app_id,code)
			if _, err := db.Exec(`
				CREATE TABLE roles_new (
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
					UNIQUE (app_id, code)
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT INTO roles_new
					(id, app_id, code, name, description, is_system,
					 created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
				SELECT
					id, app_id, code, name, description, is_system,
					created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
				FROM roles
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE roles`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE roles_new RENAME TO roles`); err != nil {
				return err
			}

			// indexes
			if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS permissions_app_resource_action_uidx ON permissions (app_id, resource, action)`); err != nil {
				return err
			}
			if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS roles_app_code_uidx ON roles (app_id, code)`); err != nil {
				return err
			}
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS permissions_app_id_idx ON permissions (app_id)`); err != nil {
				return err
			}
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS roles_app_id_idx ON roles (app_id)`); err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			// drop new indexes
			for _, idx := range []string{
				"permissions_app_resource_action_uidx",
				"roles_app_code_uidx",
				"permissions_app_id_idx",
				"roles_app_id_idx",
			} {
				if _, err := db.Exec(`DROP INDEX IF EXISTS ` + idx); err != nil {
					return err
				}
			}

			// rebuild permissions back to the global UNIQUE(resource,action), dropping app_id
			if _, err := db.Exec(`
				CREATE TABLE permissions_old (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					resource TEXT NOT NULL,
					action TEXT NOT NULL,
					UNIQUE (resource, action)
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT INTO permissions_old (id, resource, action)
				SELECT id, resource, action FROM permissions
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE permissions`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE permissions_old RENAME TO permissions`); err != nil {
				return err
			}

			// rebuild roles back to the column-level UNIQUE on code, dropping app_id
			if _, err := db.Exec(`
				CREATE TABLE roles_old (
					id TEXT PRIMARY KEY NOT NULL,
					code TEXT UNIQUE NOT NULL,
					name TEXT NOT NULL,
					description TEXT DEFAULT '',
					is_system INTEGER NOT NULL DEFAULT 0,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					created_by TEXT DEFAULT '',
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_by TEXT DEFAULT '',
					deleted_at DATETIME,
					deleted_by TEXT DEFAULT ''
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT INTO roles_old
					(id, code, name, description, is_system,
					 created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
				SELECT
					id, code, name, description, is_system,
					created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
				FROM roles
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE roles`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE roles_old RENAME TO roles`); err != nil {
				return err
			}

			// restore the original roles_code_idx (created by 006)
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS roles_code_idx ON roles (code)`); err != nil {
				return err
			}
			return nil
		},
	},
	{
		// Rebind existing global (NULL-scope) role assignments to the isme self-app
		// so token-mint perm filtering keys off a concrete app_id.
		Name: "016_rebind_user_roles_to_isme_app",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`UPDATE user_roles SET app_service_id = 'app_isme' WHERE app_service_id IS NULL`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`UPDATE user_roles SET app_service_id = NULL WHERE app_service_id = 'app_isme'`)
			return err
		},
	},
	{
		// Migrate is_admin away: every active is_admin user gets the isme admin
		// role (app-scoped to app_isme), then the column is dropped. Platform
		// superadmin = admin on the isme app from here on.
		Name: "017_migrate_is_admin_then_drop",
		Up: func(db *bun.DB) error {
			rows, err := db.Query(`SELECT id FROM users WHERE is_admin = 1 AND deleted_at IS NULL`)
			if err != nil {
				return err
			}
			adminUserIDs := []string{}
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					rows.Close()
					return err
				}
				adminUserIDs = append(adminUserIDs, id)
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return err
			}
			rows.Close()

			for _, userID := range adminUserIDs {
				_, err := db.Exec(`
					INSERT OR IGNORE INTO user_roles (id, user_id, role_id, app_service_id)
					VALUES (?, ?, 'rol_admin', 'app_isme')
				`, cryp.ULID(), userID)
				if err != nil {
					return err
				}
			}

			_, err = db.Exec(`ALTER TABLE users DROP COLUMN is_admin`)
			return err
		},
		Down: func(db *bun.DB) error {
			// re-add the column and re-derive admin flag from isme admin-role membership.
			// Lossy: users who became admin purely via the role (no original column)
			// are indistinguishable from migrated admins — both are flagged.
			if _, err := db.Exec(`ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0`); err != nil {
				return err
			}
			_, err := db.Exec(`
				UPDATE users SET is_admin = 1
				WHERE id IN (
					SELECT ur.user_id FROM user_roles ur
					WHERE ur.role_id = 'rol_admin' AND ur.app_service_id = 'app_isme'
				)
			`)
			return err
		},
	},
	{
		// Multi-app invitations: an invitation may now carry several app-scoped
		// role assignments (e.g. medioa2->Editor + rainy->Viewer) instead of a
		// single role. Each assignment is one row in user_invitation_roles.
		//
		// The legacy user_invitations.role_id column is KEPT but relaxed to be
		// nullable (decision: keep-nullable over migrate+drop). Rationale: the
		// table carries a partial UNIQUE index on (email) WHERE status=1, so a
		// full rebuild is the only way to change a column constraint; keeping the
		// column avoids touching that index and stays trivially reversible while
		// preserving any historical single-role data. New invitations leave it
		// NULL and rely solely on the child rows. Existing rows are backfilled
		// into a child row (preserving the invitation id as the FK), so accept
		// works uniformly off the child table.
		Name: "018_create_user_invitation_roles_table",
		Up: func(db *bun.DB) error {
			if _, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS user_invitation_roles (
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
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				CREATE INDEX IF NOT EXISTS user_invitation_roles_invitation_id_idx
				ON user_invitation_roles (invitation_id)
			`); err != nil {
				return err
			}

			// relax user_invitations.role_id to be nullable via a table rebuild
			// (preserving ids + the partial pending-email unique index)
			if _, err := db.Exec(`
				CREATE TABLE user_invitations_new (
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
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT INTO user_invitations_new
					(id, email, role_id, token_hash, status, expires_at, accepted_at,
					 created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
				SELECT
					id, email, role_id, token_hash, status, expires_at, accepted_at,
					created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
				FROM user_invitations
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE user_invitations`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE user_invitations_new RENAME TO user_invitations`); err != nil {
				return err
			}

			// restore the indexes from 013 on the rebuilt table
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`); err != nil {
				return err
			}
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx
				ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL
			`); err != nil {
				return err
			}

			// backfill: migrate each existing single role_id into a child row.
			// The owning role's app_id is the assignment scope (matches accept).
			rows, err := db.Query(`
				SELECT uin.id, uin.role_id, rol.app_id
				FROM user_invitations uin
				JOIN roles rol ON rol.id = uin.role_id
				WHERE uin.role_id IS NOT NULL AND uin.role_id != '' AND uin.deleted_at IS NULL
			`)
			if err != nil {
				return err
			}
			type seed struct {
				invitationID string
				roleID       string
				appServiceID string
			}
			seeds := []seed{}
			for rows.Next() {
				var s seed
				if err := rows.Scan(&s.invitationID, &s.roleID, &s.appServiceID); err != nil {
					rows.Close()
					return err
				}
				seeds = append(seeds, s)
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return err
			}
			rows.Close()

			for _, s := range seeds {
				if _, err := db.Exec(`
					INSERT INTO user_invitation_roles (id, invitation_id, role_id, app_service_id)
					VALUES (?, ?, ?, ?)
				`, cryp.ULID(), s.invitationID, s.roleID, s.appServiceID); err != nil {
					return err
				}
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			// collapse child rows back into the legacy single role_id: pick the
			// earliest assignment per invitation (lossy for multi-app invites —
			// only the first role survives), then restore the NOT NULL column.
			if _, err := db.Exec(`
				CREATE TABLE user_invitations_old (
					id TEXT PRIMARY KEY NOT NULL,
					email TEXT NOT NULL,
					role_id TEXT NOT NULL,
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
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT INTO user_invitations_old
					(id, email, role_id, token_hash, status, expires_at, accepted_at,
					 created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
				SELECT
					uin.id,
					uin.email,
					COALESCE(uin.role_id, (
						SELECT uir.role_id FROM user_invitation_roles uir
						WHERE uir.invitation_id = uin.id AND uir.deleted_at IS NULL
						ORDER BY uir.created_at ASC LIMIT 1
					), ''),
					uin.token_hash, uin.status, uin.expires_at, uin.accepted_at,
					uin.created_at, uin.created_by, uin.updated_at, uin.updated_by, uin.deleted_at, uin.deleted_by
				FROM user_invitations uin
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE user_invitations`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE user_invitations_old RENAME TO user_invitations`); err != nil {
				return err
			}

			// restore the indexes from 013
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`); err != nil {
				return err
			}
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx
				ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL
			`); err != nil {
				return err
			}

			if _, err := db.Exec(`DROP INDEX IF EXISTS user_invitation_roles_invitation_id_idx`); err != nil {
				return err
			}
			_, err := db.Exec(`DROP TABLE IF EXISTS user_invitation_roles`)
			return err
		},
	},
	{
		Name: "019_add_icon_to_permissions",
		Up: func(db *bun.DB) error {
			// per-resource icon key (e.g. "file", "database"); empty = neutral
			// default in the UI. All rows of the same (app_id, resource) share it.
			_, err := db.Exec(`
				ALTER TABLE permissions
				ADD COLUMN icon TEXT NOT NULL DEFAULT ''
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`
				ALTER TABLE permissions
				DROP COLUMN icon
			`)
			return err
		},
	},
	{
		Name: "020_add_icon_color_to_app_services",
		Up: func(db *bun.DB) error {
			// appearance keys for the app tile: icon (e.g. "shield") + color
			// palette key (e.g. "violet"). Empty = neutral fallback in the UI
			// (hashed gradient / default icon), so existing rows look unchanged.
			if _, err := db.Exec(`
				ALTER TABLE app_services
				ADD COLUMN icon TEXT NOT NULL DEFAULT ''
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				ALTER TABLE app_services
				ADD COLUMN color TEXT NOT NULL DEFAULT ''
			`); err != nil {
				return err
			}

			// backfill the seeded isme self-app with its branded appearance
			_, err := db.Exec(`
				UPDATE app_services
				SET icon = 'isme', color = 'violet'
				WHERE id = 'app_isme' AND icon = ''
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			if _, err := db.Exec(`
				ALTER TABLE app_services
				DROP COLUMN icon
			`); err != nil {
				return err
			}
			_, err := db.Exec(`
				ALTER TABLE app_services
				DROP COLUMN color
			`)
			return err
		},
	},
	{
		Name: "021_add_appearance_to_roles_permissions",
		Up: func(db *bun.DB) error {
			// appearance keys: roles get an icon (e.g. "shield") + color palette
			// key (e.g. "violet"); permissions get a per-resource color palette
			// key (all rows of the same (app_id, resource) share it, mirroring
			// the per-resource icon added in 019). Empty = neutral fallback in
			// the UI, so existing rows look unchanged.
			if _, err := db.Exec(`
				ALTER TABLE roles
				ADD COLUMN icon TEXT NOT NULL DEFAULT ''
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				ALTER TABLE roles
				ADD COLUMN color TEXT NOT NULL DEFAULT ''
			`); err != nil {
				return err
			}
			_, err := db.Exec(`
				ALTER TABLE permissions
				ADD COLUMN color TEXT NOT NULL DEFAULT ''
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			if _, err := db.Exec(`
				ALTER TABLE permissions
				DROP COLUMN color
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				ALTER TABLE roles
				DROP COLUMN color
			`); err != nil {
				return err
			}
			_, err := db.Exec(`
				ALTER TABLE roles
				DROP COLUMN icon
			`)
			return err
		},
	},
	{
		// Session auto-revoke scheduler config: a typed single-row table that
		// drives the cronjob revoking expired sessions. Also seeds the
		// settings:read / settings:update permissions used to gate the
		// management endpoints, granting both to the admin system role.
		Name: "022_create_session_revoke_config",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS session_revoke_config (
					id INTEGER PRIMARY KEY,
					enabled INTEGER NOT NULL DEFAULT 0,
					cron TEXT NOT NULL DEFAULT '0 3 * * *',
					last_run_at DATETIME,
					last_revoked_count INTEGER,
					updated_at DATETIME,
					updated_by TEXT
				)
			`)
			if err != nil {
				return err
			}

			// single config row (id=1, disabled by default)
			_, err = db.Exec(`
				INSERT OR IGNORE INTO session_revoke_config (id, enabled, cron)
				VALUES (1, 0, '0 3 * * *')
			`)
			if err != nil {
				return err
			}

			// extend the permission catalog with the settings actions
			settingsPermissions := []struct {
				resource string
				action   string
			}{
				{"settings", "read"},
				{"settings", "update"},
			}
			for _, permission := range settingsPermissions {
				_, err := db.Exec(`INSERT OR IGNORE INTO permissions (resource, action) VALUES (?, ?)`, permission.resource, permission.action)
				if err != nil {
					return err
				}
				var permissionID int64
				row := db.QueryRow(`SELECT id FROM permissions WHERE resource = ? AND action = ?`, permission.resource, permission.action)
				if err := row.Scan(&permissionID); err != nil {
					return err
				}
				// admin keeps holding the full catalog
				_, err = db.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES ('rol_admin', ?)`, permissionID)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`
				DELETE FROM role_permissions
				WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'settings' AND action IN ('read', 'update'))
			`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`DELETE FROM permissions WHERE resource = 'settings' AND action IN ('read', 'update')`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`DROP TABLE IF EXISTS session_revoke_config`)
			return err
		},
	},
	{
		// Token-rotation tracking. Two per-session columns on user_sessions
		// (refresh_count lifetime + last_refreshed_at timestamp) drive the
		// per-session "refreshed N× · last refreshed {when}" UI. The separate
		// token_rotation_events table records one row per refresh so the
		// Welcome "Token rotations" card can compute an accurate sliding-24h
		// count (a stored 24h counter would go stale).
		Name: "023_create_token_rotation_tracking",
		Up: func(db *bun.DB) error {
			if _, err := db.Exec(`
				ALTER TABLE user_sessions
				ADD COLUMN refresh_count INTEGER NOT NULL DEFAULT 0
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				ALTER TABLE user_sessions
				ADD COLUMN last_refreshed_at TIMESTAMP
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS token_rotation_events (
					id TEXT PRIMARY KEY,
					user_id TEXT NOT NULL,
					session_id TEXT NOT NULL,
					rotated_at TIMESTAMP NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				CREATE INDEX IF NOT EXISTS token_rotation_events_user_rotated_idx
				ON token_rotation_events (user_id, rotated_at)
			`); err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			if _, err := db.Exec(`DROP INDEX IF EXISTS token_rotation_events_user_rotated_idx`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE IF EXISTS token_rotation_events`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE user_sessions DROP COLUMN last_refreshed_at`); err != nil {
				return err
			}
			if _, err := db.Exec(`ALTER TABLE user_sessions DROP COLUMN refresh_count`); err != nil {
				return err
			}
			return nil
		},
	},
	{
		// Rotation cleanup scheduler config: a typed single-row table that drives
		// the cronjob pruning token_rotation_events older than a configurable
		// retention window (the table grows unbounded — one row per refresh).
		// The settings:read / settings:update permissions are already seeded by
		// migration 022, so this migration only adds the config table.
		Name: "024_create_rotation_cleanup_config",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS rotation_cleanup_config (
					id INTEGER PRIMARY KEY,
					enabled INTEGER NOT NULL DEFAULT 0,
					cron TEXT NOT NULL DEFAULT '0 4 * * *',
					retention_hours INTEGER NOT NULL DEFAULT 48,
					last_run_at DATETIME,
					last_cleaned_count INTEGER,
					updated_at DATETIME,
					updated_by TEXT
				)
			`)
			if err != nil {
				return err
			}

			// single config row (id=1, disabled by default)
			_, err = db.Exec(`
				INSERT OR IGNORE INTO rotation_cleanup_config (id, enabled, cron, retention_hours)
				VALUES (1, 0, '0 4 * * *', 48)
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS rotation_cleanup_config`)
			return err
		},
	},
	{
		// Consolidate the two typed scheduled-job config tables
		// (session_revoke_config + rotation_cleanup_config) into one generic
		// schedule_config keyed by job_key. Shared columns stay typed;
		// job-specific knobs move into the params JSON blob and per-run results
		// into the last_result JSON blob (replacing last_revoked_count /
		// last_cleaned_count). The job-key strings must match the JobKey consts
		// in internal/scheduler (asserted by a cross-package test).
		Name: "025_consolidate_schedule_config",
		Up: func(db *bun.DB) error {
			if _, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS schedule_config (
					job_key TEXT PRIMARY KEY,
					enabled INTEGER NOT NULL DEFAULT 0,
					cron TEXT NOT NULL,
					params TEXT NOT NULL DEFAULT '{}',
					last_run_at DATETIME,
					last_result TEXT,
					updated_at DATETIME,
					updated_by TEXT
				)
			`); err != nil {
				return err
			}

			// data-migrate the existing session-revoke row (id=1) — empty params,
			// last_revoked_count folds into last_result json {"revoked": N}.
			if _, err := db.Exec(`
				INSERT INTO schedule_config (job_key, enabled, cron, params, last_run_at, last_result, updated_at, updated_by)
				SELECT 'session_revoke', enabled, cron, '{}', last_run_at,
					CASE WHEN last_revoked_count IS NULL THEN NULL ELSE json_object('revoked', last_revoked_count) END,
					updated_at, updated_by
				FROM session_revoke_config WHERE id = 1
			`); err != nil {
				return err
			}

			// data-migrate the existing rotation-cleanup row (id=1) — retention_hours
			// folds into params json, last_cleaned_count into last_result {"cleaned": N}.
			if _, err := db.Exec(`
				INSERT INTO schedule_config (job_key, enabled, cron, params, last_run_at, last_result, updated_at, updated_by)
				SELECT 'rotation_cleanup', enabled, cron, json_object('retention_hours', retention_hours), last_run_at,
					CASE WHEN last_cleaned_count IS NULL THEN NULL ELSE json_object('cleaned', last_cleaned_count) END,
					updated_at, updated_by
				FROM rotation_cleanup_config WHERE id = 1
			`); err != nil {
				return err
			}

			// guard: if a source row was somehow absent, seed the default job rows
			// so a fresh DB still ends with both jobs present (disabled by default).
			if _, err := db.Exec(`
				INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
				VALUES ('session_revoke', 0, '0 3 * * *', '{}')
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT OR IGNORE INTO schedule_config (job_key, enabled, cron, params)
				VALUES ('rotation_cleanup', 0, '0 4 * * *', json_object('retention_hours', 48))
			`); err != nil {
				return err
			}

			if _, err := db.Exec(`DROP TABLE IF EXISTS session_revoke_config`); err != nil {
				return err
			}
			if _, err := db.Exec(`DROP TABLE IF EXISTS rotation_cleanup_config`); err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			// recreate the two typed tables (mirror of migrations 022 / 024).
			if _, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS session_revoke_config (
					id INTEGER PRIMARY KEY,
					enabled INTEGER NOT NULL DEFAULT 0,
					cron TEXT NOT NULL DEFAULT '0 3 * * *',
					last_run_at DATETIME,
					last_revoked_count INTEGER,
					updated_at DATETIME,
					updated_by TEXT
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS rotation_cleanup_config (
					id INTEGER PRIMARY KEY,
					enabled INTEGER NOT NULL DEFAULT 0,
					cron TEXT NOT NULL DEFAULT '0 4 * * *',
					retention_hours INTEGER NOT NULL DEFAULT 48,
					last_run_at DATETIME,
					last_cleaned_count INTEGER,
					updated_at DATETIME,
					updated_by TEXT
				)
			`); err != nil {
				return err
			}

			// migrate rows back: extract job-specific knobs out of the JSON blobs.
			if _, err := db.Exec(`
				INSERT OR IGNORE INTO session_revoke_config (id, enabled, cron, last_run_at, last_revoked_count, updated_at, updated_by)
				SELECT 1, enabled, cron, last_run_at, json_extract(last_result, '$.revoked'), updated_at, updated_by
				FROM schedule_config WHERE job_key = 'session_revoke'
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT OR IGNORE INTO rotation_cleanup_config (id, enabled, cron, retention_hours, last_run_at, last_cleaned_count, updated_at, updated_by)
				SELECT 1, enabled, cron,
					COALESCE(json_extract(params, '$.retention_hours'), 48),
					last_run_at, json_extract(last_result, '$.cleaned'), updated_at, updated_by
				FROM schedule_config WHERE job_key = 'rotation_cleanup'
			`); err != nil {
				return err
			}

			// seed defaults if a source row was absent (mirror of 022 / 024).
			if _, err := db.Exec(`
				INSERT OR IGNORE INTO session_revoke_config (id, enabled, cron)
				VALUES (1, 0, '0 3 * * *')
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`
				INSERT OR IGNORE INTO rotation_cleanup_config (id, enabled, cron, retention_hours)
				VALUES (1, 0, '0 4 * * *', 48)
			`); err != nil {
				return err
			}

			_, err := db.Exec(`DROP TABLE IF EXISTS schedule_config`)
			return err
		},
	},
	{
		Name: "026_create_activity_events_table",
		Up: func(db *bun.DB) error {
			// activity_events is an append-only audit log powering the Welcome
			// "Recent activity" feed. NOTE: this table grows unbounded — one row
			// per recorded user action (sign_in/out, password change, invitation).
			// Retention/pruning is deliberately deferred (no scheduler job yet);
			// add a prune job before this becomes a volume concern.
			if _, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS activity_events (
					id TEXT PRIMARY KEY NOT NULL,
					user_id TEXT NOT NULL,
					type TEXT NOT NULL,
					meta TEXT NOT NULL DEFAULT '{}',
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
				)
			`); err != nil {
				return err
			}
			if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_events_user_created ON activity_events (user_id, created_at)`); err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			if _, err := db.Exec(`DROP INDEX IF EXISTS idx_activity_events_user_created`); err != nil {
				return err
			}
			_, err := db.Exec(`DROP TABLE IF EXISTS activity_events`)
			return err
		},
	},
}
