package history

import (
	"github.com/vukyn/isme/internal/domains/migration/models"

	"github.com/uptrace/bun"
)

var m028AddUserAvatarURL = models.Migration{
	Name: "028_add_user_avatar_url",
	Up: func(db *bun.DB) error {
		// avatar_url holds the medioa object URL (or a pasted external link)
		// for the user's profile photo. Nullable TEXT — empty when the user
		// has no avatar (the UI falls back to the gradient + initials).
		_, err := db.Exec(`ALTER TABLE users ADD COLUMN avatar_url TEXT`)
		return err
	},
	Down: func(db *bun.DB) error {
		// SQLite DROP COLUMN is supported on the versions this repo targets.
		_, err := db.Exec(`ALTER TABLE users DROP COLUMN avatar_url`)
		return err
	},
}
