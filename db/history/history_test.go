package history

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/uptrace/bun"

	pkgMigrate "github.com/vukyn/kuery/bun/migrate"
	"github.com/vukyn/kuery/bun/pgtest"
)

// These tests exercise the migrations themselves, so every database starts
// empty (pgtest.EmptyDB) and no template schema is needed.
func TestMain(m *testing.M) {
	os.Exit(pgtest.Run(m, pgtest.Options{}))
}

// frozenNames is the exact name sequence the runner records. Production holds
// ids 1-35 = 000_baseline, 001…034, and the runner matches by name, so an
// applied name may never change, move or disappear. New migrations are only
// ever appended.
var frozenNames = []string{
	"000_baseline",
	"001_create_users_table",
	"002_create_user_sessions_table",
	"003_add_token_id_to_user_sessions",
	"004_create_app_services_table",
	"005_add_is_admin_to_users",
	"006_create_roles_table",
	"007_create_permissions_table",
	"008_create_role_permissions_table",
	"009_create_user_roles_table",
	"010_seed_rbac",
	"011_add_app_service_id_to_user_sessions",
	"012_add_user_verification",
	"013_create_user_invitations_table",
	"014_seed_isme_app_service",
	"015_add_app_id_to_rbac",
	"016_rebind_user_roles_to_isme_app",
	"017_migrate_is_admin_then_drop",
	"018_create_user_invitation_roles_table",
	"019_add_icon_to_permissions",
	"020_add_icon_color_to_app_services",
	"021_add_appearance_to_roles_permissions",
	"022_create_session_revoke_config",
	"023_create_token_rotation_tracking",
	"024_create_rotation_cleanup_config",
	"025_consolidate_schedule_config",
	"026_create_activity_events_table",
	"027_seed_activity_cleanup_schedule",
	"028_add_user_avatar_url",
	"029_seed_database_backup_schedule",
	"030_fix_bool_columns_pg",
	"031_add_redirect_urls_to_app_services",
	"032_users_email_unique_live_only",
	"033_seed_invitation_cleanup_schedule",
	"034_seed_unverified_user_cleanup_schedule",
	"035_delete_database_backup_schedule",
}

// productionApplied is how many leading entries production had recorded when
// the history was collapsed (000_baseline + 001…034).
const productionApplied = 35

var applicationTables = []string{
	"activity_events",
	"app_services",
	"permissions",
	"role_permissions",
	"roles",
	"schedule_config",
	"token_rotation_events",
	"user_invitation_roles",
	"user_invitations",
	"user_roles",
	"user_sessions",
	"users",
}

func migrationNames() []string {
	names := make([]string, 0, len(Migrations))
	for _, migration := range Migrations {
		names = append(names, migration.Name)
	}
	return names
}

func TestMigrationNamesFrozen(t *testing.T) {
	names := migrationNames()
	if len(names) < len(frozenNames) {
		t.Fatalf("expected at least %d migrations, got %d", len(frozenNames), len(names))
	}
	for i, want := range frozenNames {
		if names[i] != want {
			t.Errorf("migration %d: got %q, want %q", i, names[i], want)
		}
	}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			t.Errorf("duplicate migration name %q", name)
		}
		seen[name] = true
	}
}

// scheduleRow is one schedule_config row as stored.
type scheduleRow struct {
	JobKey  string `bun:"job_key"`
	Enabled bool   `bun:"enabled"`
	Cron    string `bun:"cron"`
	Params  string `bun:"params"`
}

func readSchedules(t *testing.T, db *bun.DB) map[string]scheduleRow {
	t.Helper()
	var rows []scheduleRow
	if err := db.NewSelect().Table("schedule_config").Column("job_key", "enabled", "cron", "params").Scan(context.Background(), &rows); err != nil {
		t.Fatalf("read schedule_config: %v", err)
	}
	schedules := map[string]scheduleRow{}
	for _, row := range rows {
		schedules[row.JobKey] = row
	}
	return schedules
}

// readBookkeeping returns the recorded migration names keyed by id.
func readBookkeeping(t *testing.T, db *bun.DB) map[int64]string {
	t.Helper()
	var rows []pkgMigrate.History
	if err := db.NewSelect().Model(&rows).Scan(context.Background()); err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	recorded := map[int64]string{}
	for _, row := range rows {
		recorded[row.ID] = row.Name
	}
	return recorded
}

func existingTables(t *testing.T, db *bun.DB) []string {
	t.Helper()
	var tables []string
	if err := db.NewRaw(`SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`).Scan(context.Background(), &tables); err != nil {
		t.Fatalf("list tables: %v", err)
	}
	return tables
}

func insertUser(db *bun.DB, id, email string, deleted bool) error {
	deletedAt := any(nil)
	if deleted {
		deletedAt = "2026-01-01 00:00:00+00"
	}
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO users (id, name, email, password, status, deleted_at)
		VALUES (?, ?, ?, 'hash', 1, ?)`, id, "name-"+id, email, deletedAt)
	return err
}

func TestFreshInstall(t *testing.T) {
	db := pgtest.EmptyDB(t)
	ctx := context.Background()

	stats, err := pkgMigrate.Run(db, Migrations)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if stats.TotalSuccess != len(Migrations) || stats.TotalSkipped != 0 {
		t.Fatalf("expected %d applied and 0 skipped, got %+v", len(Migrations), stats)
	}

	recorded := readBookkeeping(t, db)
	for i, name := range migrationNames() {
		if recorded[int64(i)+1] != name {
			t.Errorf("bookkeeping id %d: got %q, want %q", i+1, recorded[int64(i)+1], name)
		}
	}

	schedules := readSchedules(t, db)
	wantKeys := []string{"activity_cleanup", "invitation_cleanup", "rotation_cleanup", "session_revoke", "unverified_user_cleanup"}
	gotKeys := make([]string, 0, len(schedules))
	for key, row := range schedules {
		gotKeys = append(gotKeys, key)
		if row.Enabled {
			t.Errorf("job %s seeded enabled", key)
		}
	}
	slices.Sort(gotKeys)
	if !slices.Equal(gotKeys, wantKeys) {
		t.Fatalf("schedule keys: got %v, want %v", gotKeys, wantKeys)
	}

	// flags mapped to Go bools must be native booleans
	for _, column := range [][2]string{{"users", "is_verified"}, {"roles", "is_system"}, {"schedule_config", "enabled"}} {
		var dataType string
		if err := db.NewRaw(`SELECT data_type FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = ? AND column_name = ?`, column[0], column[1]).Scan(ctx, &dataType); err != nil {
			t.Fatalf("column type %s.%s: %v", column[0], column[1], err)
		}
		if dataType != "boolean" {
			t.Errorf("%s.%s: got %s, want boolean", column[0], column[1], dataType)
		}
	}

	// email is unique among live users only
	var indexDefinition string
	if err := db.NewRaw(`SELECT indexdef FROM pg_indexes WHERE tablename = 'users' AND indexname = 'users_email_live_uidx'`).Scan(ctx, &indexDefinition); err != nil {
		t.Fatalf("users_email_live_uidx missing: %v", err)
	}
	if err := insertUser(db, "usr_live", "live@example.com", false); err != nil {
		t.Fatalf("insert live user: %v", err)
	}
	if err := insertUser(db, "usr_duplicate", "live@example.com", false); err == nil {
		t.Error("expected a second live user with the same email to be rejected")
	}
	if err := insertUser(db, "usr_gone", "gone@example.com", true); err != nil {
		t.Fatalf("insert deleted user: %v", err)
	}
	if err := insertUser(db, "usr_reinvited", "gone@example.com", false); err != nil {
		t.Errorf("expected a live user to take a soft-deleted user's email, got %v", err)
	}
	if err := insertUser(db, "usr_gone_again", "live@example.com", true); err != nil {
		t.Errorf("expected a soft-deleted row to share a live email, got %v", err)
	}
}

// TestUpOnProductionShapedDatabase replays what `up` does on production: the
// baseline schema with its seed rows (database_backup included), and
// bookkeeping ids 1-35 = 000_baseline, 001…034.
func TestUpOnProductionShapedDatabase(t *testing.T) {
	db := pgtest.EmptyDB(t)
	ctx := context.Background()

	if err := BaselineMigration.Up(db); err != nil {
		t.Fatalf("baseline up: %v", err)
	}
	if _, ok := readSchedules(t, db)["database_backup"]; !ok {
		t.Fatal("expected the baseline to seed the database_backup row")
	}

	// creates the bookkeeping table only
	if _, err := pkgMigrate.Run(db, nil); err != nil {
		t.Fatalf("create bookkeeping table: %v", err)
	}
	for i, name := range frozenNames[:productionApplied] {
		if _, err := db.ExecContext(ctx, `INSERT INTO migrations (id, name, executed_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, int64(i)+1, name); err != nil {
			t.Fatalf("record %s: %v", name, err)
		}
	}

	stats, err := pkgMigrate.Run(db, Migrations)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if stats.TotalSuccess != len(Migrations)-productionApplied || stats.TotalSkipped != productionApplied {
		t.Fatalf("expected %d applied and %d skipped, got %+v", len(Migrations)-productionApplied, productionApplied, stats)
	}
	if _, ok := readSchedules(t, db)["database_backup"]; ok {
		t.Error("expected 035 to delete the database_backup row")
	}
	if got := readBookkeeping(t, db)[productionApplied+1]; got != "035_delete_database_backup_schedule" {
		t.Errorf("bookkeeping id %d: got %q, want 035_delete_database_backup_schedule", productionApplied+1, got)
	}
}

func TestResetRoundTrip(t *testing.T) {
	db := pgtest.EmptyDB(t)

	if _, err := pkgMigrate.Run(db, Migrations); err != nil {
		t.Fatalf("run: %v", err)
	}
	rolledBack, err := pkgMigrate.Reset(db, Migrations)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if rolledBack != len(Migrations) {
		t.Fatalf("expected %d rolled back, got %d", len(Migrations), rolledBack)
	}
	if tables := existingTables(t, db); !slices.Equal(tables, []string{"migrations"}) {
		t.Fatalf("expected only the migrations table after reset, got %v", tables)
	}
	if recorded := readBookkeeping(t, db); len(recorded) != 0 {
		t.Fatalf("expected empty bookkeeping after reset, got %v", recorded)
	}

	stats, err := pkgMigrate.Run(db, Migrations)
	if err != nil {
		t.Fatalf("run after reset: %v", err)
	}
	if stats.TotalSuccess != len(Migrations) {
		t.Fatalf("expected %d applied after reset, got %+v", len(Migrations), stats)
	}
	tables := existingTables(t, db)
	for _, table := range applicationTables {
		if !slices.Contains(tables, table) {
			t.Errorf("table %s missing after re-run", table)
		}
	}
}

func Test035DownRestoresRow(t *testing.T) {
	db := pgtest.EmptyDB(t)

	if _, err := pkgMigrate.Run(db, Migrations); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := pkgMigrate.RollbackLast(db, Migrations); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	row, ok := readSchedules(t, db)["database_backup"]
	if !ok {
		t.Fatal("expected 035 down to restore the database_backup row")
	}
	want := scheduleRow{JobKey: "database_backup", Enabled: false, Cron: "0 3 * * *", Params: `{"retain_count":10}`}
	if row != want {
		t.Errorf("restored row: got %+v, want %+v", row, want)
	}
}
