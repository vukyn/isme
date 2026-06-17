package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
	"github.com/vukyn/kuery/cryp"
)

// Multi-app invitations: an invitation may now carry several app-scoped
// role assignments (e.g. medioa2->Editor + rainy->Viewer) instead of a
// single role. Each assignment is one row in user_invitation_roles.
//
// The legacy user_invitations.role_id column is KEPT but relaxed to be
// nullable (decision: keep-nullable over migrate+drop). Rationale: the
// table carries a partial UNIQUE index on (email) WHERE status=1, so a
// full rebuild is the only way to change a column constraint; keeping the
// column avoids touching that index and stays trivially reversible while
// preserving any historical single-role data. New invitations leave it
// NULL and rely solely on the child rows. Existing rows are backfilled
// into a child row (preserving the invitation id as the FK), so accept
// works uniformly off the child table.
var m018CreateUserInvitationRolesTable = models.Migration{
	Name: "018_create_user_invitation_roles_table",
	Up: func(db *bun.DB) error {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS user_invitation_roles (
				id TEXT PRIMARY KEY NOT NULL,
				invitation_id TEXT NOT NULL,
				role_id TEXT NOT NULL,
				app_service_id TEXT NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				created_by TEXT DEFAULT '',
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_by TEXT DEFAULT '',
				deleted_at DATETIME,
				deleted_by TEXT DEFAULT '',
				FOREIGN KEY (invitation_id) REFERENCES user_invitations (id)
			)
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS user_invitation_roles_invitation_id_idx
			ON user_invitation_roles (invitation_id)
		`); err != nil {
			return err
		}

		// relax user_invitations.role_id to be nullable via a table rebuild
		// (preserving ids + the partial pending-email unique index)
		if _, err := db.Exec(`
			CREATE TABLE user_invitations_new (
				id TEXT PRIMARY KEY NOT NULL,
				email TEXT NOT NULL,
				role_id TEXT,
				token_hash TEXT NOT NULL,
				status INTEGER NOT NULL DEFAULT 1,
				expires_at DATETIME NOT NULL,
				accepted_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				created_by TEXT DEFAULT '',
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_by TEXT DEFAULT '',
				deleted_at DATETIME,
				deleted_by TEXT DEFAULT ''
			)
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			INSERT INTO user_invitations_new
				(id, email, role_id, token_hash, status, expires_at, accepted_at,
				 created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
			SELECT
				id, email, role_id, token_hash, status, expires_at, accepted_at,
				created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
			FROM user_invitations
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`DROP TABLE user_invitations`); err != nil {
			return err
		}
		if _, err := db.Exec(`ALTER TABLE user_invitations_new RENAME TO user_invitations`); err != nil {
			return err
		}

		// restore the indexes from 013 on the rebuilt table
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`); err != nil {
			return err
		}
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx
			ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL
		`); err != nil {
			return err
		}

		// backfill: migrate each existing single role_id into a child row.
		// The owning role's app_id is the assignment scope (matches accept).
		rows, err := db.Query(`
			SELECT uin.id, uin.role_id, rol.app_id
			FROM user_invitations uin
			JOIN roles rol ON rol.id = uin.role_id
			WHERE uin.role_id IS NOT NULL AND uin.role_id != '' AND uin.deleted_at IS NULL
		`)
		if err != nil {
			return err
		}
		type seed struct {
			invitationID string
			roleID       string
			appServiceID string
		}
		seeds := []seed{}
		for rows.Next() {
			var s seed
			if err := rows.Scan(&s.invitationID, &s.roleID, &s.appServiceID); err != nil {
				rows.Close()
				return err
			}
			seeds = append(seeds, s)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		for _, s := range seeds {
			if _, err := db.Exec(`
				INSERT INTO user_invitation_roles (id, invitation_id, role_id, app_service_id)
				VALUES (?, ?, ?, ?)
			`, cryp.ULID(), s.invitationID, s.roleID, s.appServiceID); err != nil {
				return err
			}
		}
		return nil
	},
	Down: func(db *bun.DB) error {
		// collapse child rows back into the legacy single role_id: pick the
		// earliest assignment per invitation (lossy for multi-app invites —
		// only the first role survives), then restore the NOT NULL column.
		if _, err := db.Exec(`
			CREATE TABLE user_invitations_old (
				id TEXT PRIMARY KEY NOT NULL,
				email TEXT NOT NULL,
				role_id TEXT NOT NULL,
				token_hash TEXT NOT NULL,
				status INTEGER NOT NULL DEFAULT 1,
				expires_at DATETIME NOT NULL,
				accepted_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				created_by TEXT DEFAULT '',
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_by TEXT DEFAULT '',
				deleted_at DATETIME,
				deleted_by TEXT DEFAULT ''
			)
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			INSERT INTO user_invitations_old
				(id, email, role_id, token_hash, status, expires_at, accepted_at,
				 created_at, created_by, updated_at, updated_by, deleted_at, deleted_by)
			SELECT
				uin.id,
				uin.email,
				COALESCE(uin.role_id, (
					SELECT uir.role_id FROM user_invitation_roles uir
					WHERE uir.invitation_id = uin.id AND uir.deleted_at IS NULL
					ORDER BY uir.created_at ASC LIMIT 1
				), ''),
				uin.token_hash, uin.status, uin.expires_at, uin.accepted_at,
				uin.created_at, uin.created_by, uin.updated_at, uin.updated_by, uin.deleted_at, uin.deleted_by
			FROM user_invitations uin
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`DROP TABLE user_invitations`); err != nil {
			return err
		}
		if _, err := db.Exec(`ALTER TABLE user_invitations_old RENAME TO user_invitations`); err != nil {
			return err
		}

		// restore the indexes from 013
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`); err != nil {
			return err
		}
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx
			ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL
		`); err != nil {
			return err
		}

		if _, err := db.Exec(`DROP INDEX IF EXISTS user_invitation_roles_invitation_id_idx`); err != nil {
			return err
		}
		_, err := db.Exec(`DROP TABLE IF EXISTS user_invitation_roles`)
		return err
	},
}
