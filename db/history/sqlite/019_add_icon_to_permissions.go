package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m019AddIconToPermissions = pkgMigrate.Migration{
	Name: "019_add_icon_to_permissions",
	Up: func(db bun.IDB) error {
		// per-resource icon key (e.g. "file", "database"); empty = neutral
		// default in the UI. All rows of the same (app_id, resource) share it.
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE permissions
			ADD COLUMN icon TEXT NOT NULL DEFAULT ''
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE permissions
			DROP COLUMN icon
		`)
		return err
	},
}
