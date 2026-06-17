package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m012AddUserVerification = pkgMigrate.Migration{
	Name: "012_add_user_verification",
	Up: func(db bun.IDB) error {
		// new signups start unverified (login blocked until verified)
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE users
			ADD COLUMN is_verified INTEGER NOT NULL DEFAULT 0
		`)
		if err != nil {
			return err
		}

		// backfill: every existing account is considered verified
		_, err = db.ExecContext(context.Background(), `UPDATE users SET is_verified = 1`)
		if err != nil {
			return err
		}

		// extend the permission catalog with the verification action
		_, err = db.ExecContext(context.Background(), `INSERT OR IGNORE INTO permissions (resource, action) VALUES ('user', 'verify')`)
		if err != nil {
			return err
		}
		var permissionID int64
		row := db.QueryRowContext(context.Background(), `SELECT id FROM permissions WHERE resource = 'user' AND action = 'verify'`)
		if err := row.Scan(&permissionID); err != nil {
			return err
		}

		// admin keeps holding the full catalog
		_, err = db.ExecContext(context.Background(), `INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES ('rol_admin', ?)`, permissionID)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			DELETE FROM role_permissions
			WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'user' AND action = 'verify')
		`)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(context.Background(), `DELETE FROM permissions WHERE resource = 'user' AND action = 'verify'`)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(context.Background(), `
			ALTER TABLE users
			DROP COLUMN is_verified
		`)
		return err
	},
}
