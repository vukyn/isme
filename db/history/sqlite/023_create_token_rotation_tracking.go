package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

// Token-rotation tracking. Two per-session columns on user_sessions
// (refresh_count lifetime + last_refreshed_at timestamp) drive the
// per-session "refreshed N× · last refreshed {when}" UI. The separate
// token_rotation_events table records one row per refresh so the
// Welcome "Token rotations" card can compute an accurate sliding-24h
// count (a stored 24h counter would go stale).
var m023CreateTokenRotationTracking = models.Migration{
	Name: "023_create_token_rotation_tracking",
	Up: func(db *bun.DB) error {
		if _, err := db.Exec(`
			ALTER TABLE user_sessions
			ADD COLUMN refresh_count INTEGER NOT NULL DEFAULT 0
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			ALTER TABLE user_sessions
			ADD COLUMN last_refreshed_at TIMESTAMP
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS token_rotation_events (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				session_id TEXT NOT NULL,
				rotated_at TIMESTAMP NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
			)
		`); err != nil {
			return err
		}
		if _, err := db.Exec(`
			CREATE INDEX IF NOT EXISTS token_rotation_events_user_rotated_idx
			ON token_rotation_events (user_id, rotated_at)
		`); err != nil {
			return err
		}
		return nil
	},
	Down: func(db *bun.DB) error {
		if _, err := db.Exec(`DROP INDEX IF EXISTS token_rotation_events_user_rotated_idx`); err != nil {
			return err
		}
		if _, err := db.Exec(`DROP TABLE IF EXISTS token_rotation_events`); err != nil {
			return err
		}
		if _, err := db.Exec(`ALTER TABLE user_sessions DROP COLUMN last_refreshed_at`); err != nil {
			return err
		}
		if _, err := db.Exec(`ALTER TABLE user_sessions DROP COLUMN refresh_count`); err != nil {
			return err
		}
		return nil
	},
}
