package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Rebind existing global (NULL-scope) role assignments to the isme self-app
// so token-mint perm filtering keys off a concrete app_id.
var m016RebindUserRolesToIsmeApp = pkgMigrate.Migration{
	Name: "016_rebind_user_roles_to_isme_app",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `UPDATE user_roles SET app_service_id = 'app_isme' WHERE app_service_id IS NULL`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `UPDATE user_roles SET app_service_id = NULL WHERE app_service_id = 'app_isme'`)
		return err
	},
}
