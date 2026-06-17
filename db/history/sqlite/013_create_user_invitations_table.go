package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m013CreateUserInvitationsTable = pkgMigrate.Migration{
	Name: "013_create_user_invitations_table",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			CREATE TABLE IF NOT EXISTS user_invitations (
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
		`)
		if err != nil {
			return err
		}

		// Create indexes
		_, err = db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS user_invitations_token_hash_idx ON user_invitations (token_hash)`)
		if err != nil {
			return err
		}

		_, err = db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS user_invitations_email_idx ON user_invitations (email)`)
		if err != nil {
			return err
		}

		// at most one live pending invitation per email
		_, err = db.ExecContext(context.Background(), `
			CREATE UNIQUE INDEX IF NOT EXISTS user_invitations_pending_email_uidx
			ON user_invitations (email) WHERE status = 1 AND deleted_at IS NULL
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DROP TABLE IF EXISTS user_invitations`)
		return err
	},
}
