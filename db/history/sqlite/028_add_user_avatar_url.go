package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

var m028AddUserAvatarURL = pkgMigrate.Migration{
	Name: "028_add_user_avatar_url",
	Up: func(db bun.IDB) error {
		// avatar_url holds the medioa object URL (or a pasted external link)
		// for the user's profile photo. Nullable TEXT — empty when the user
		// has no avatar (the UI falls back to the gradient + initials).
		_, err := db.ExecContext(context.Background(), `ALTER TABLE users ADD COLUMN avatar_url TEXT`)
		return err
	},
	Down: func(db bun.IDB) error {
		// SQLite DROP COLUMN is supported on the versions this repo targets.
		_, err := db.ExecContext(context.Background(), `ALTER TABLE users DROP COLUMN avatar_url`)
		return err
	},
}
