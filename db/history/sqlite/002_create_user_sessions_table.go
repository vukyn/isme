package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m002CreateUserSessionsTable = pkgMigrate.Migration{
	Name: "002_create_user_sessions_table",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			CREATE TABLE IF NOT EXISTS user_sessions (
				id TEXT PRIMARY KEY NOT NULL,
				user_id TEXT NOT NULL,
				email TEXT NOT NULL,
				refresh_token TEXT,
				expires_at TEXT,
				last_login_at TEXT,
				status INTEGER DEFAULT 1,
				client_ip TEXT NOT NULL,
				user_agent TEXT,
				created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL
			)
		`)
		if err != nil {
			return err
		}

		// Create indexes
		_, err = db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS user_sessions_refresh_token_idx ON user_sessions (refresh_token)`)
		if err != nil {
			return err
		}

		_, err = db.ExecContext(context.Background(), `CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions (user_id)`)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `DROP TABLE IF EXISTS user_sessions`)
		return err
	},
}
