package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

var m005AddIsAdminToUsers = models.Migration{
	Name: "005_add_is_admin_to_users",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE users
			ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE users
			DROP COLUMN is_admin
		`)
		return err
	},
}
