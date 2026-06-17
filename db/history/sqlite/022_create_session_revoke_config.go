package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Session auto-revoke scheduler config: a typed single-row table that
// drives the cronjob revoking expired sessions. Also seeds the
// settings:read / settings:update permissions used to gate the
// management endpoints, granting both to the admin system role.
var m022CreateSessionRevokeConfig = pkgMigrate.Migration{
	Name: "022_create_session_revoke_config",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS session_revoke_config (
				id INTEGER PRIMARY KEY,
				enabled INTEGER NOT NULL DEFAULT 0,
				cron TEXT NOT NULL DEFAULT '0 3 * * *',
				last_run_at DATETIME,
				last_revoked_count INTEGER,
				updated_at DATETIME,
				updated_by TEXT
			)
		`)
		if err != nil {
			return err
		}

		// single config row (id=1, disabled by default)
		_, err = db.Exec(`
			INSERT OR IGNORE INTO session_revoke_config (id, enabled, cron)
			VALUES (1, 0, '0 3 * * *')
		`)
		if err != nil {
			return err
		}

		// extend the permission catalog with the settings actions
		settingsPermissions := []struct {
			resource string
			action   string
		}{
			{"settings", "read"},
			{"settings", "update"},
		}
		for _, permission := range settingsPermissions {
			_, err := db.Exec(`INSERT OR IGNORE INTO permissions (resource, action) VALUES (?, ?)`, permission.resource, permission.action)
			if err != nil {
				return err
			}
			var permissionID int64
			row := db.QueryRow(`SELECT id FROM permissions WHERE resource = ? AND action = ?`, permission.resource, permission.action)
			if err := row.Scan(&permissionID); err != nil {
				return err
			}
			// admin keeps holding the full catalog
			_, err = db.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES ('rol_admin', ?)`, permissionID)
			if err != nil {
				return err
			}
		}
		return nil
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`
			DELETE FROM role_permissions
			WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'settings' AND action IN ('read', 'update'))
		`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`DELETE FROM permissions WHERE resource = 'settings' AND action IN ('read', 'update')`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`DROP TABLE IF EXISTS session_revoke_config`)
		return err
	},
}
