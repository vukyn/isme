package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

var m012AddUserVerification = models.Migration{
	Name: "012_add_user_verification",
	Up: func(db *bun.DB) error {
		// new signups start unverified (login blocked until verified)
		_, err := db.Exec(`
			ALTER TABLE users
			ADD COLUMN is_verified INTEGER NOT NULL DEFAULT 0
		`)
		if err != nil {
			return err
		}

		// backfill: every existing account is considered verified
		_, err = db.Exec(`UPDATE users SET is_verified = 1`)
		if err != nil {
			return err
		}

		// extend the permission catalog with the verification action
		_, err = db.Exec(`INSERT OR IGNORE INTO permissions (resource, action) VALUES ('user', 'verify')`)
		if err != nil {
			return err
		}
		var permissionID int64
		row := db.QueryRow(`SELECT id FROM permissions WHERE resource = 'user' AND action = 'verify'`)
		if err := row.Scan(&permissionID); err != nil {
			return err
		}

		// admin keeps holding the full catalog
		_, err = db.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES ('rol_admin', ?)`, permissionID)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`
			DELETE FROM role_permissions
			WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'user' AND action = 'verify')
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`DELETE FROM permissions WHERE resource = 'user' AND action = 'verify'`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`
			ALTER TABLE users
			DROP COLUMN is_verified
		`)
		return err
	},
}
