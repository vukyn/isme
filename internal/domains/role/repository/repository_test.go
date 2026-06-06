package repository

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/vukyn/kuery/cryp"
)

// newTestDB opens an in-memory SQLite database and applies every migration
// from db/history/sqlite, including the RBAC seed.
func newTestDB(t *testing.T) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open(sqliteshim.ShimName, ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	// a single connection keeps every query on the same in-memory database
	sqldb.SetMaxOpenConns(1)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	for _, migration := range sqliteHistory.Migrations {
		if err := migration.Up(db); err != nil {
			t.Fatalf("migration %s failed: %v", migration.Name, err)
		}
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func insertUser(t *testing.T, db *bun.DB, id string) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO users (id, name, email, password, status)
		VALUES (?, ?, ?, 'secret', 1)
	`, id, id, id+"@example.com")
	if err != nil {
		t.Fatalf("insert user %s: %v", id, err)
	}
}

func assignRole(t *testing.T, db *bun.DB, userID, roleID string, appServiceID any) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO user_roles (id, user_id, role_id, app_service_id)
		VALUES (?, ?, ?, ?)
	`, cryp.ULID(), userID, roleID, appServiceID)
	if err != nil {
		t.Fatalf("assign role %s to user %s: %v", roleID, userID, err)
	}
}

func TestGetPermissionCodesByUserID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	roleRepository := NewRepository(db)

	memberReadCodes := []string{"user:read", "user_session:read", "app_service:read", "role:read"}

	// user with a global member role
	insertUser(t, db, "user-global")
	assignRole(t, db, "user-global", "rol_member", nil)

	// user with an app-scoped member role only
	insertUser(t, db, "user-scoped")
	assignRole(t, db, "user-scoped", "rol_member", "app-1")

	// user holding a custom role that gets soft-deleted
	insertUser(t, db, "user-deleted-role")
	if _, err := db.Exec(`
		INSERT INTO roles (id, code, name, is_system) VALUES ('rol_custom', 'custom', 'Custom', 0)
	`); err != nil {
		t.Fatalf("insert custom role: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT 'rol_custom', id FROM permissions WHERE resource = 'user' AND action = 'read'
	`); err != nil {
		t.Fatalf("grant permission to custom role: %v", err)
	}
	assignRole(t, db, "user-deleted-role", "rol_custom", nil)
	if _, err := db.Exec(`UPDATE roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = 'rol_custom'`); err != nil {
		t.Fatalf("soft delete custom role: %v", err)
	}

	// user with the global admin role
	insertUser(t, db, "user-admin")
	assignRole(t, db, "user-admin", "rol_admin", nil)

	tests := []struct {
		name         string
		userID       string
		appServiceID string
		wantCount    int
		wantCodes    []string
	}{
		{
			name:         "global role resolves with empty app service ID",
			userID:       "user-global",
			appServiceID: "",
			wantCount:    4,
			wantCodes:    memberReadCodes,
		},
		{
			name:         "global role also resolves within an app scope",
			userID:       "user-global",
			appServiceID: "app-1",
			wantCount:    4,
			wantCodes:    memberReadCodes,
		},
		{
			name:         "app-scoped role does not resolve globally",
			userID:       "user-scoped",
			appServiceID: "",
			wantCount:    0,
		},
		{
			name:         "app-scoped role resolves with matching app service ID",
			userID:       "user-scoped",
			appServiceID: "app-1",
			wantCount:    4,
			wantCodes:    memberReadCodes,
		},
		{
			name:         "app-scoped role does not resolve for another app",
			userID:       "user-scoped",
			appServiceID: "app-2",
			wantCount:    0,
		},
		{
			name:         "soft-deleted role is excluded",
			userID:       "user-deleted-role",
			appServiceID: "",
			wantCount:    0,
		},
		{
			name:         "admin holds the full catalog",
			userID:       "user-admin",
			appServiceID: "",
			wantCount:    18,
			wantCodes:    []string{"user:read", "user:delete", "role:assign", "app_service:rotate_secret", "user_session:revoke"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := roleRepository.GetPermissionCodesByUserID(ctx, tt.userID, tt.appServiceID)
			if err != nil {
				t.Fatalf("GetPermissionCodesByUserID() error = %v", err)
			}
			if len(codes) != tt.wantCount {
				t.Errorf("got %d codes %v, want %d", len(codes), codes, tt.wantCount)
			}
			for _, code := range tt.wantCodes {
				if !slices.Contains(codes, code) {
					t.Errorf("codes %v missing %q", codes, code)
				}
			}
		})
	}
}
