package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// App-owned RBAC: permissions and roles become owned by an app_service.
// Existing rows default to the isme self-app. The old global UNIQUE
// constraints are relaxed to be app-scoped via a table rebuild that
// PRESERVES id values (role_permissions references permission ids, and
// role ids like 'rol_admin' are referenced by value elsewhere).
var m015AddAppIDToRBAC = pkgMigrate.Migration{
	Name: "015_add_app_id_to_rbac",
	Up: func(db bun.IDB) error {
		// add the owning-app column (existing rows -> isme self-app)
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE permissions ADD COLUMN app_id TEXT NOT NULL DEFAULT 'app_isme'`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE roles ADD COLUMN app_id TEXT NOT NULL DEFAULT 'app_isme'`); err != nil {
			return err
		}

		// rebuild permissions: UNIQUE(resource,action) -> UNIQUE(app_id,resource,action)
		if _, err := db.ExecContext(context.Background(), `
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
		if _, err := db.ExecContext(context.Background(), `
			INSERT INTO permissions_new (id, app_id, resource, action)
			SELECT id, app_id, resource, action FROM permissions
		`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `DROP TABLE permissions`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE permissions_new RENAME TO permissions`); err != nil {
			return err
		}

		// rebuild roles: drop column-level UNIQUE on code -> UNIQUE(app_id,code)
		if _, err := db.ExecContext(context.Background(), `
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
		if _, err := db.ExecContext(context.Background(), `
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
		if _, err := db.ExecContext(context.Background(), `DROP TABLE roles`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE roles_new RENAME TO roles`); err != nil {
			return err
		}

		// indexes
		if _, err := db.ExecContext(context.Background(), `CREATE UNIQUE INDEX IF NOT EXISTS permissions_app_resource_action_uidx ON permissions (app_id, resource, action)`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `CREATE UNIQUE INDEX IF NOT EXISTS roles_app_code_uidx ON roles (app_id, code)`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS permissions_app_id_idx ON permissions (app_id)`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS roles_app_id_idx ON roles (app_id)`); err != nil {
			return err
		}
		return nil
	},
	Down: func(db bun.IDB) error {
		// drop new indexes
		for _, idx := range []string{
			"permissions_app_resource_action_uidx",
			"roles_app_code_uidx",
			"permissions_app_id_idx",
			"roles_app_id_idx",
		} {
			if _, err := db.ExecContext(context.Background(), `DROP INDEX IF EXISTS `+idx); err != nil {
				return err
			}
		}

		// rebuild permissions back to the global UNIQUE(resource,action), dropping app_id
		if _, err := db.ExecContext(context.Background(), `
			CREATE TABLE permissions_old (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				resource TEXT NOT NULL,
				action TEXT NOT NULL,
				UNIQUE (resource, action)
			)
		`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `
			INSERT INTO permissions_old (id, resource, action)
			SELECT id, resource, action FROM permissions
		`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `DROP TABLE permissions`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE permissions_old RENAME TO permissions`); err != nil {
			return err
		}

		// rebuild roles back to the column-level UNIQUE on code, dropping app_id
		if _, err := db.ExecContext(context.Background(), `
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
		if _, err := db.ExecContext(context.Background(), `
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
		if _, err := db.ExecContext(context.Background(), `DROP TABLE roles`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE roles_old RENAME TO roles`); err != nil {
			return err
		}

		// restore the original roles_code_idx (created by 006)
		if _, err := db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS roles_code_idx ON roles (code)`); err != nil {
			return err
		}
		return nil
	},
}
