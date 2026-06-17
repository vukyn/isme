package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m004CreateAppServicesTable = pkgMigrate.Migration{
	Name: "004_create_app_services_table",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
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
		_, err = db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS app_services_app_code_idx ON app_services (app_code)`)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DROP TABLE IF EXISTS app_services`)
		return err
	},
}
