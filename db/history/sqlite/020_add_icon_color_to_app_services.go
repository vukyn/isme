package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m020AddIconColorToAppServices = pkgMigrate.Migration{
	Name: "020_add_icon_color_to_app_services",
	Up: func(db bun.IDB) error {
		// appearance keys for the app tile: icon (e.g. "shield") + color
		// palette key (e.g. "violet"). Empty = neutral fallback in the UI
		// (hashed gradient / default icon), so existing rows look unchanged.
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE app_services
			ADD COLUMN icon TEXT NOT NULL DEFAULT ''
		`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE app_services
			ADD COLUMN color TEXT NOT NULL DEFAULT ''
		`); err != nil {
			return err
		}

		// backfill the seeded isme self-app with its branded appearance
		_, err := db.ExecContext(context.Background(), `
			UPDATE app_services
			SET icon = 'isme', color = 'violet'
			WHERE id = 'app_isme' AND icon = ''
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE app_services
			DROP COLUMN icon
		`); err != nil {
			return err
		}
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE app_services
			DROP COLUMN color
		`)
		return err
	},
}
