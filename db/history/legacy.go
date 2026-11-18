package history

import (
	"github.com/uptrace/bun"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
)

// legacyMigrations keeps the names of the incremental history that existed
// before the baseline was squashed. Their effects are all inside
// BaselineMigration, so each one is a no-op that only gets recorded.
//
// ⚠️ The names are FROZEN. The runner matches migrations by name: production
// already recorded every one of these, and rolling back fails with
// "migration <name> not found" when an applied name leaves the slice. Never
// rename, reorder or remove an entry.
var legacyMigrations = []pkgMigrate.Migration{
	noop("001_create_users_table"),
	noop("002_create_user_sessions_table"),
	noop("003_add_token_id_to_user_sessions"),
	noop("004_create_app_services_table"),
	noop("005_add_is_admin_to_users"),
	noop("006_create_roles_table"),
	noop("007_create_permissions_table"),
	noop("008_create_role_permissions_table"),
	noop("009_create_user_roles_table"),
	noop("010_seed_rbac"),
	noop("011_add_app_service_id_to_user_sessions"),
	noop("012_add_user_verification"),
	noop("013_create_user_invitations_table"),
	noop("014_seed_isme_app_service"),
	noop("015_add_app_id_to_rbac"),
	noop("016_rebind_user_roles_to_isme_app"),
	noop("017_migrate_is_admin_then_drop"),
	noop("018_create_user_invitation_roles_table"),
	noop("019_add_icon_to_permissions"),
	noop("020_add_icon_color_to_app_services"),
	noop("021_add_appearance_to_roles_permissions"),
	noop("022_create_session_revoke_config"),
	noop("023_create_token_rotation_tracking"),
	noop("024_create_rotation_cleanup_config"),
	noop("025_consolidate_schedule_config"),
	noop("026_create_activity_events_table"),
	noop("027_seed_activity_cleanup_schedule"),
	noop("028_add_user_avatar_url"),
	noop("029_seed_database_backup_schedule"),
	noop("030_fix_bool_columns_pg"),
	noop("031_add_redirect_urls_to_app_services"),
	noop("032_users_email_unique_live_only"),
	noop("033_seed_invitation_cleanup_schedule"),
	noop("034_seed_unverified_user_cleanup_schedule"),
}

// noop builds a migration that changes nothing in either direction.
func noop(name string) pkgMigrate.Migration {
	return pkgMigrate.Migration{
		Name: name,
		Up:   func(bun.IDB) error { return nil },
		Down: func(bun.IDB) error { return nil },
	}
}
