package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Fix the three flag columns whose Go entity field is a bool (users.is_verified,
// roles.is_system, schedule_config.enabled) but which were created as INTEGER on
// Postgres by the early baseline. bun's pgdialect serializes a Go bool as the
// boolean literal TRUE/FALSE, so any INSERT/UPDATE touching these columns failed
// on Postgres with: column "<col>" is of type integer but expression is of type
// boolean (SQLSTATE 42804). Convert them to native BOOLEAN.
//
// SQLite stores bool as INTEGER 0/1 and is loose about the type, so the existing
// SQLite databases already work — this migration is a Postgres-only no-op there
// (and SQLite has no ALTER COLUMN TYPE anyway). Fresh Postgres installs created
// via the baseline already declare these columns BOOLEAN, so this ALTER is an
// idempotent no-op for them.
var m030FixBoolColumnsPg = pkgMigrate.Migration{
	Name: "030_fix_bool_columns_pg",
	Up: func(db bun.IDB) error {
		if !isPostgres(db) {
			return nil
		}
		stmts := []string{
			`ALTER TABLE users ALTER COLUMN is_verified DROP DEFAULT`,
			`ALTER TABLE users ALTER COLUMN is_verified TYPE BOOLEAN USING (is_verified <> 0)`,
			`ALTER TABLE users ALTER COLUMN is_verified SET DEFAULT FALSE`,

			`ALTER TABLE roles ALTER COLUMN is_system DROP DEFAULT`,
			`ALTER TABLE roles ALTER COLUMN is_system TYPE BOOLEAN USING (is_system <> 0)`,
			`ALTER TABLE roles ALTER COLUMN is_system SET DEFAULT FALSE`,

			`ALTER TABLE schedule_config ALTER COLUMN enabled DROP DEFAULT`,
			`ALTER TABLE schedule_config ALTER COLUMN enabled TYPE BOOLEAN USING (enabled <> 0)`,
			`ALTER TABLE schedule_config ALTER COLUMN enabled SET DEFAULT FALSE`,
		}
		for _, stmt := range stmts {
			if _, err := db.ExecContext(context.Background(), stmt); err != nil {
				return err
			}
		}
		return nil
	},
	Down: func(db bun.IDB) error {
		if !isPostgres(db) {
			return nil
		}
		stmts := []string{
			`ALTER TABLE users ALTER COLUMN is_verified DROP DEFAULT`,
			`ALTER TABLE users ALTER COLUMN is_verified TYPE INTEGER USING (CASE WHEN is_verified THEN 1 ELSE 0 END)`,
			`ALTER TABLE users ALTER COLUMN is_verified SET DEFAULT 0`,

			`ALTER TABLE roles ALTER COLUMN is_system DROP DEFAULT`,
			`ALTER TABLE roles ALTER COLUMN is_system TYPE INTEGER USING (CASE WHEN is_system THEN 1 ELSE 0 END)`,
			`ALTER TABLE roles ALTER COLUMN is_system SET DEFAULT 0`,

			`ALTER TABLE schedule_config ALTER COLUMN enabled DROP DEFAULT`,
			`ALTER TABLE schedule_config ALTER COLUMN enabled TYPE INTEGER USING (CASE WHEN enabled THEN 1 ELSE 0 END)`,
			`ALTER TABLE schedule_config ALTER COLUMN enabled SET DEFAULT 0`,
		}
		for _, stmt := range stmts {
			if _, err := db.ExecContext(context.Background(), stmt); err != nil {
				return err
			}
		}
		return nil
	},
}
