package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// isme is itself an app_service that owns the original RBAC catalog.
// The self-app row is a logical owner only — isme never SSO's into
// itself, so the app_secret is left as a placeholder (decision 1).
var m014SeedIsmeAppService = pkgMigrate.Migration{
	Name: "014_seed_isme_app_service",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			INSERT OR IGNORE INTO app_services
				(id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
			VALUES
				('app_isme', 'isme', 'ISME', '', '', 'authen', 1)
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DELETE FROM app_services WHERE id = 'app_isme'`)
		return err
	},
}
