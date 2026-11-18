// Package testdb gives isme tests a real, migrated Postgres database.
//
// Test-only: import it from _test.go files and nowhere else. It wraps kuery
// bun/pgtest, which starts one throwaway Postgres container per test binary
// (Docker must be running), migrates the full isme history into a template
// once, and hands each Open call its own clone of that template. Nothing here
// reads .env or DB_* variables. `go test -short` skips every database test.
//
// Every package that calls Open needs a TestMain that calls Main:
//
//	func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }
package testdb

import (
	"testing"

	"github.com/uptrace/bun"

	"github.com/vukyn/isme/db/history"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
	"github.com/vukyn/kuery/bun/pgtest"
)

// Main runs the package's tests and tears the container down afterwards.
func Main(m *testing.M) int {
	return pgtest.Run(m, pgtest.Options{
		Migrate: func(db *bun.DB) error {
			_, err := pkgMigrate.Run(db, history.Migrations)
			return err
		},
	})
}

// Open returns a fresh database holding the fully migrated schema and seed
// rows. It is dropped when the test finishes.
func Open(t testing.TB) *bun.DB {
	t.Helper()
	return pgtest.DB(t)
}
