package history

import (
	"context"
	"strings"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Make users.email unique among LIVE users only. The inline UNIQUE constraint
// also covered soft-deleted rows, so an email whose account was deleted could
// never be invited again. It is replaced by a partial unique index that skips
// rows with deleted_at set.
//
// SQLite cannot drop an inline constraint, so the table is rebuilt (ids and
// every column preserved). No other table references users, and users carries
// no other index, trigger or view to restore. Postgres drops the constraint by
// its looked-up name rather than assuming the default one.
//
// Down puts the table-wide constraint back; it fails, as expected, once a live
// and a soft-deleted row share an email.
var m032UsersEmailUniqueLiveOnly = pkgMigrate.Migration{
	Name: "032_users_email_unique_live_only",
	Up: func(db bun.IDB) error {
		if isPostgres(db) {
			names, err := usersEmailUniqueConstraintNames(db)
			if err != nil {
				return err
			}
			for _, name := range names {
				if _, err := db.ExecContext(context.Background(), `ALTER TABLE users DROP CONSTRAINT "`+strings.ReplaceAll(name, `"`, `""`)+`"`); err != nil {
					return err
				}
			}
			return createUsersEmailLiveIndex(db)
		}
		if err := rebuildSQLiteUsersTable(db, "email TEXT"); err != nil {
			return err
		}
		return createUsersEmailLiveIndex(db)
	},
	Down: func(db bun.IDB) error {
		if isPostgres(db) {
			if _, err := db.ExecContext(context.Background(), `DROP INDEX IF EXISTS users_email_live_uidx`); err != nil {
				return err
			}
			_, err := db.ExecContext(context.Background(), `ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email)`)
			return err
		}
		// the rebuild drops the old table, and the partial index with it
		return rebuildSQLiteUsersTable(db, "email TEXT UNIQUE")
	},
}

func createUsersEmailLiveIndex(db bun.IDB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE UNIQUE INDEX IF NOT EXISTS users_email_live_uidx
		ON users (email) WHERE deleted_at IS NULL
	`)
	return err
}

// usersEmailUniqueConstraintNames lists every single-column unique constraint
// on users.email.
func usersEmailUniqueConstraintNames(db bun.IDB) ([]string, error) {
	rows, err := db.QueryContext(context.Background(), `
		SELECT con.conname
		FROM pg_constraint con
		JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
		WHERE con.conrelid = 'users'::regclass
			AND con.contype = 'u'
			AND att.attname = 'email'
			AND array_length(con.conkey, 1) = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make([]string, 0, 1)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// rebuildSQLiteUsersTable recreates users with the given email column
// definition, copying every row and column across unchanged.
func rebuildSQLiteUsersTable(db bun.IDB, emailColumn string) error {
	statements := []string{
		`CREATE TABLE users_new (
			id TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			` + emailColumn + `,
			password TEXT NOT NULL,
			status INTEGER NOT NULL,
			last_login_at DATETIME DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by TEXT DEFAULT '',
			deleted_at DATETIME,
			deleted_by TEXT DEFAULT '',
			is_verified INTEGER NOT NULL DEFAULT 0,
			avatar_url TEXT
		)`,
		`INSERT INTO users_new
			(id, name, email, password, status, last_login_at, created_at, created_by,
			 updated_at, updated_by, deleted_at, deleted_by, is_verified, avatar_url)
		SELECT
			id, name, email, password, status, last_login_at, created_at, created_by,
			updated_at, updated_by, deleted_at, deleted_by, is_verified, avatar_url
		FROM users`,
		`DROP TABLE users`,
		`ALTER TABLE users_new RENAME TO users`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(context.Background(), statement); err != nil {
			return err
		}
	}
	return nil
}
