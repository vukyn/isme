package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m001CreateUsersTable = pkgMigrate.Migration{
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
}
