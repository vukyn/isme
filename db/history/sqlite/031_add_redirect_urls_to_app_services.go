package history

import (
	"context"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"

	"github.com/uptrace/bun"
)

// Add the additional-redirect-URL allowlist column to app_services. The single
// redirect_url stays the default/PRIMARY SSO callback; redirect_urls is a JSON
// array (stored as TEXT) of EXTRA permitted callbacks the SSO client may also
// redirect to after login (capped at 3 in the usecase/model layer).
//
// The column is a portable TEXT JSON array with a '[]' default, so — like
// m020's icon/color TEXT columns — both SQLite and Postgres take the identical
// DDL and no isPostgres branch is needed (JSON-as-TEXT stays as-is on both
// dialects per the baseline translation rules). Existing rows default to the
// empty list, leaving the redirect_url-only behaviour unchanged.
var m031AddRedirectURLsToAppServices = pkgMigrate.Migration{
	Name: "031_add_redirect_urls_to_app_services",
	Up: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE app_services
			ADD COLUMN redirect_urls TEXT NOT NULL DEFAULT '[]'
		`)
		return err
	},
	Down: func(db bun.IDB) error {
		_, err := db.ExecContext(context.Background(), `
			ALTER TABLE app_services
			DROP COLUMN redirect_urls
		`)
		return err
	},
}
