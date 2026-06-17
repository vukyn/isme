package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m019AddIconToPermissions = pkgMigrate.Migration{
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
}
