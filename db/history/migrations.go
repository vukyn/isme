package history

import (
	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
)

// Migrations holds every migration in execution order: the frozen baseline,
// the legacy no-op names, then the tail. One migration per file
// (NNN_<name>.go); append new ones at the end and never edit an applied one.
var Migrations = append(append(
	[]pkgMigrate.Migration{BaselineMigration},
	legacyMigrations...),
	m035DeleteDatabaseBackupSchedule,
)
