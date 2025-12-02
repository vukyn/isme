package history

import (
	"isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

// Migrations holds all database migrations
var Migrations = []models.Migration{
	{
		Name: "001_create_users_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS users (
					id TEXT PRIMARY KEY NOT NULL,
					name TEXT NOT NULL,
					email TEXT UNIQUE,
					password TEXT NOT NULL,
					status INTEGER NOT NULL,
					last_login_at DATETIME DEFAULT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					created_by INTEGER DEFAULT 0,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_by INTEGER DEFAULT 0,
					deleted_at DATETIME,
					deleted_by INTEGER DEFAULT 0
				)
			`)
			return err
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS users`)
			return err
		},
	},
	{
		Name: "002_create_user_sessions_table",
		Up: func(db *bun.DB) error {
			_, err := db.Exec(`
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
			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_sessions_refresh_token_idx ON user_sessions (refresh_token)`)
			if err != nil {
				return err
			}

			_, err = db.Exec(`CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions (user_id)`)
			if err != nil {
				return err
			}
			return nil
		},
		Down: func(db *bun.DB) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS user_sessions`)
			return err
		},
	},
	{
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
	},
}
