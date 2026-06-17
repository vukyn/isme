package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/cryp"
)

// Migrate is_admin away: every active is_admin user gets the isme admin
// role (app-scoped to app_isme), then the column is dropped. Platform
// superadmin = admin on the isme app from here on.
var m017MigrateIsAdminThenDrop = pkgMigrate.Migration{
	Name: "017_migrate_is_admin_then_drop",
	Up: func(db bun.IDB) error {
		rows, err := db.QueryContext(context.Background(), `SELECT id FROM users WHERE is_admin = 1 AND deleted_at IS NULL`)
		if err != nil {
			return err
		}
		adminUserIDs := []string{}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			adminUserIDs = append(adminUserIDs, id)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		for _, userID := range adminUserIDs {
			_, err := db.ExecContext(context.Background(), `
				INSERT OR IGNORE INTO user_roles (id, user_id, role_id, app_service_id)
				VALUES (?, ?, 'rol_admin', 'app_isme')
			`, cryp.ULID(), userID)
			if err != nil {
				return err
			}
		}

		_, err = db.ExecContext(context.Background(), `ALTER TABLE users DROP COLUMN is_admin`)
		return err
	},
	Down: func(db bun.IDB) error {
		// re-add the column and re-derive admin flag from isme admin-role membership.
		// Lossy: users who became admin purely via the role (no original column)
		// are indistinguishable from migrated admins — both are flagged.
		if _, err := db.ExecContext(context.Background(), `ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
		_, err := db.ExecContext(context.Background(), `
			UPDATE users SET is_admin = 1
			WHERE id IN (
				SELECT ur.user_id FROM user_roles ur
				WHERE ur.role_id = 'rol_admin' AND ur.app_service_id = 'app_isme'
			)
		`)
		return err
	},
}
