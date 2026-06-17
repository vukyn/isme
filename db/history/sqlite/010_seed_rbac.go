package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/cryp"
)

var m010SeedRBAC = pkgMigrate.Migration{
	Name: "010_seed_rbac",
	Up: func(db *bun.DB) error {
		// seed system roles with deterministic IDs
		systemRoles := []struct {
			id          string
			code        string
			name        string
			description string
		}{
			{"rol_admin", "admin", "Admin", "Full access to every resource"},
			{"rol_member", "member", "Member", "Read access to core resources"},
			{"rol_viewer", "viewer", "Viewer", "Read-only access to core resources"},
		}
		for _, role := range systemRoles {
			_, err := db.Exec(`
				INSERT OR IGNORE INTO roles (id, code, name, description, is_system)
				VALUES (?, ?, ?, ?, 1)
			`, role.id, role.code, role.name, role.description)
			if err != nil {
				return err
			}
		}

		// seed the permission catalog
		permissionCatalog := []struct {
			resource string
			action   string
		}{
			{"user", "read"},
			{"user", "create"},
			{"user", "update"},
			{"user", "delete"},
			{"user", "reset_password"},
			{"user_session", "read"},
			{"user_session", "delete"},
			{"user_session", "revoke"},
			{"app_service", "read"},
			{"app_service", "create"},
			{"app_service", "update"},
			{"app_service", "delete"},
			{"app_service", "rotate_secret"},
			{"role", "read"},
			{"role", "create"},
			{"role", "update"},
			{"role", "delete"},
			{"role", "assign"},
		}
		permissionIDs := map[string]int64{}
		for _, permission := range permissionCatalog {
			_, err := db.Exec(`INSERT OR IGNORE INTO permissions (resource, action) VALUES (?, ?)`, permission.resource, permission.action)
			if err != nil {
				return err
			}
			var permissionID int64
			row := db.QueryRow(`SELECT id FROM permissions WHERE resource = ? AND action = ?`, permission.resource, permission.action)
			if err := row.Scan(&permissionID); err != nil {
				return err
			}
			permissionIDs[permission.resource+":"+permission.action] = permissionID
		}

		// grants: admin holds the full catalog; member and viewer hold core reads
		grantPermission := func(roleID string, permissionID int64) error {
			_, err := db.Exec(`INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, roleID, permissionID)
			return err
		}
		for _, permission := range permissionCatalog {
			if err := grantPermission("rol_admin", permissionIDs[permission.resource+":"+permission.action]); err != nil {
				return err
			}
		}
		readOnlyCodes := []string{"user:read", "user_session:read", "app_service:read", "role:read"}
		for _, code := range readOnlyCodes {
			if err := grantPermission("rol_member", permissionIDs[code]); err != nil {
				return err
			}
			if err := grantPermission("rol_viewer", permissionIDs[code]); err != nil {
				return err
			}
		}

		// backfill global roles for existing users (new signups get no role by default)
		type userRow struct {
			id      string
			isAdmin bool
		}
		rows, err := db.Query(`SELECT id, is_admin FROM users WHERE deleted_at IS NULL`)
		if err != nil {
			return err
		}
		defer rows.Close()

		userRows := []userRow{}
		for rows.Next() {
			row := userRow{}
			if err := rows.Scan(&row.id, &row.isAdmin); err != nil {
				return err
			}
			userRows = append(userRows, row)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for _, row := range userRows {
			roleID := "rol_member"
			if row.isAdmin {
				roleID = "rol_admin"
			}
			_, err := db.Exec(`
				INSERT OR IGNORE INTO user_roles (id, user_id, role_id, app_service_id)
				VALUES (?, ?, ?, NULL)
			`, cryp.ULID(), row.id, roleID)
			if err != nil {
				return err
			}
		}
		return nil
	},
	Down: func(db *bun.DB) error {
		systemRoleIDs := []string{"rol_admin", "rol_member", "rol_viewer"}
		for _, roleID := range systemRoleIDs {
			if _, err := db.Exec(`DELETE FROM user_roles WHERE role_id = ?`, roleID); err != nil {
				return err
			}
			if _, err := db.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID); err != nil {
				return err
			}
			if _, err := db.Exec(`DELETE FROM roles WHERE id = ?`, roleID); err != nil {
				return err
			}
		}

		// the catalog is seeded exclusively by this migration
		_, err := db.Exec(`DELETE FROM permissions`)
		return err
	},
}
