package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m007CreatePermissionsTable = pkgMigrate.Migration{
	Name: "007_create_permissions_table",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			CREATE TABLE IF NOT EXISTS permissions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				resource TEXT NOT NULL,
				action TEXT NOT NULL,
				UNIQUE (resource, action)
			)
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DROP TABLE IF EXISTS permissions`)
		return err
	},
}
