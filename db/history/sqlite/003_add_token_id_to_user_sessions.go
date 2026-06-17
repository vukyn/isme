package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m003AddTokenIDToUserSessions = pkgMigrate.Migration{
	Name: "003_add_token_id_to_user_sessions",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE user_sessions
			ADD COLUMN token_id TEXT NOT NULL DEFAULT ''
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE user_sessions
			DROP COLUMN token_id
		`)
		return err
	},
}
