package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
)

// Migrations holds all database migrations in execution order.
// One migration per file (NNN_<name>.go); append new ones at the end.
var Migrations = []pkgMigrate.Migration{
	m001CreateUsersTable,
	m002CreateUserSessionsTable,
	m003AddTokenIDToUserSessions,
	m004CreateAppServicesTable,
	m005AddIsAdminToUsers,
	m006CreateRolesTable,
	m007CreatePermissionsTable,
	m008CreateRolePermissionsTable,
	m009CreateUserRolesTable,
	m010SeedRBAC,
	m011AddAppServiceIDToUserSessions,
	m012AddUserVerification,
	m013CreateUserInvitationsTable,
	m014SeedIsmeAppService,
	m015AddAppIDToRBAC,
	m016RebindUserRolesToIsmeApp,
	m017MigrateIsAdminThenDrop,
	m018CreateUserInvitationRolesTable,
	m019AddIconToPermissions,
	m020AddIconColorToAppServices,
	m021AddAppearanceToRolesPermissions,
	m022CreateSessionRevokeConfig,
	m023CreateTokenRotationTracking,
	m024CreateRotationCleanupConfig,
	m025ConsolidateScheduleConfig,
	m026CreateActivityEventsTable,
	m027SeedActivityCleanupSchedule,
	m028AddUserAvatarURL,
	m029SeedDatabaseBackupSchedule,
	m030FixBoolColumnsPg,
}
