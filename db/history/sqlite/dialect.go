package history

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// isPostgres reports whether the migration is running against a Postgres
// backend. Migrations use it to branch dialect-specific DDL: the SQLite branch
// stays byte-identical to the original SQLite-only DDL (zero risk to existing
// SQLite dev/prod databases), while the Postgres branch substitutes the
// equivalent native syntax (TIMESTAMPTZ, BOOLEAN TRUE/FALSE, ON CONFLICT,
// native ALTER, IDENTITY, ...).
func isPostgres(db bun.IDB) bool {
	return db.Dialect().Name() == dialect.PG
}
