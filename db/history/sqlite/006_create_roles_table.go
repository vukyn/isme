package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m006CreateRolesTable = pkgMigrate.Migration{
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
}
