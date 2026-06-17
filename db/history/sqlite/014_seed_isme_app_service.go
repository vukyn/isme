package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

// isme is itself an app_service that owns the original RBAC catalog.
// The self-app row is a logical owner only — isme never SSO's into
// itself, so the app_secret is left as a placeholder (decision 1).
var m014SeedIsmeAppService = models.Migration{
	Name: "014_seed_isme_app_service",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO app_services
				(id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
			VALUES
				('app_isme', 'isme', 'ISME', '', '', 'authen', 1)
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`DELETE FROM app_services WHERE id = 'app_isme'`)
		return err
	},
}
