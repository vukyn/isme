package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m021AddAppearanceToRolesPermissions = pkgMigrate.Migration{
	Name: "021_add_appearance_to_roles_permissions",
	Up: func(db bun.IDB) error {
		// appearance keys: roles get an icon (e.g. "shield") + color palette
		// key (e.g. "violet"); permissions get a per-resource color palette
		// key (all rows of the same (app_id, resource) share it, mirroring
		// the per-resource icon added in 019). Empty = neutral fallback in
		// the UI, so existing rows look unchanged.
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE roles
			ADD COLUMN icon TEXT NOT NULL DEFAULT ''
		`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE roles
			ADD COLUMN color TEXT NOT NULL DEFAULT ''
		`); err != nil {
			return err
		}
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE permissions
			ADD COLUMN color TEXT NOT NULL DEFAULT ''
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE permissions
			DROP COLUMN color
		`); err != nil {
			return err
		}
		if _, err := db.ExecContext(context.Background(), `
			ALTER TABLE roles
			DROP COLUMN color
		`); err != nil {
			return err
		}
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE roles
			DROP COLUMN icon
		`)
		return err
	},
}
