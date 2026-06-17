package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

var m003AddTokenIDToUserSessions = models.Migration{
	Name: "003_add_token_id_to_user_sessions",
	Up: func(db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE user_sessions
			ADD COLUMN token_id TEXT NOT NULL DEFAULT ''
		`)
		return err
	},
	Down: func(db *bun.DB) error {
		_, err := db.Exec(`
			ALTER TABLE user_sessions
			DROP COLUMN token_id
		`)
		return err
	},
}
