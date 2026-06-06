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
}
