package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

var m009CreateUserRolesTable = models.Migration{
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
}
