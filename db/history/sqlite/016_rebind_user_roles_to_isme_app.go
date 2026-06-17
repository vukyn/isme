package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

// Rebind existing global (NULL-scope) role assignments to the isme self-app
// so token-mint perm filtering keys off a concrete app_id.
var m016RebindUserRolesToIsmeApp = models.Migration{
	Name: "016_rebind_user_roles_to_isme_app",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`UPDATE user_roles SET app_service_id = 'app_isme' WHERE app_service_id IS NULL`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`UPDATE user_roles SET app_service_id = NULL WHERE app_service_id = 'app_isme'`)
		return err
	},
}
