package repository

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	"github.com/vukyn/isme/internal/domains/role/models"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/vukyn/kuery/cryp"
)

// newTestDB opens an in-memory SQLite database and applies every migration
// from db/history/sqlite, including the RBAC seed and the app-owned migrations.
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

// insertApp seeds an additional app_service plus an admin role + perm catalog for it.
func insertApp(t *testing.T, db *bun.DB, appID, appCode string, resources []string) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO app_services (id, app_code, app_name, app_secret, redirect_url, ctx_info, status)
		VALUES (?, ?, ?, '', '', 'authen', 1)
	`, appID, appCode, appCode)
	if err != nil {
		t.Fatalf("insert app %s: %v", appID, err)
	}
	roleID := "rol_" + appCode + "_admin"
	if _, err := db.Exec(`
		INSERT INTO roles (id, app_id, code, name, is_system) VALUES (?, ?, 'admin', 'Admin', 0)
	`, roleID, appID); err != nil {
		t.Fatalf("insert app admin role: %v", err)
	}
	for _, resource := range resources {
		if _, err := db.Exec(`
			INSERT INTO permissions (app_id, resource, action) VALUES (?, ?, 'read')
		`, appID, resource); err != nil {
			t.Fatalf("insert app perm: %v", err)
		}
		if _, err := db.Exec(`
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT ?, id FROM permissions WHERE app_id = ? AND resource = ? AND action = 'read'
		`, roleID, appID, resource); err != nil {
			t.Fatalf("grant app perm: %v", err)
		}
	}
}

func TestGetPermissionCodesByUserID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	roleRepository := NewRepository(db)

	memberReadCodes := []string{"user:read", "user_session:read", "app_service:read", "role:read"}
	isme := roleConstants.APP_ID_ISME

	// user with an isme-app member role (scoped to app_isme)
	insertUser(t, db, "user-isme")
	assignRole(t, db, "user-isme", "rol_member", isme)

	// user with a custom role that gets soft-deleted
	insertUser(t, db, "user-deleted-role")
	if _, err := db.Exec(`
		INSERT INTO roles (id, app_id, code, name, is_system) VALUES ('rol_custom', ?, 'custom', 'Custom', 0)
	`, isme); err != nil {
		t.Fatalf("insert custom role: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT 'rol_custom', id FROM permissions WHERE app_id = ? AND resource = 'user' AND action = 'read'
	`, isme); err != nil {
		t.Fatalf("grant permission to custom role: %v", err)
	}
	assignRole(t, db, "user-deleted-role", "rol_custom", isme)
	if _, err := db.Exec(`UPDATE roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = 'rol_custom'`); err != nil {
		t.Fatalf("soft delete custom role: %v", err)
	}

	// user with the isme admin role
	insertUser(t, db, "user-admin")
	assignRole(t, db, "user-admin", "rol_admin", isme)

	tests := []struct {
		name      string
		userID    string
		appID     string
		wantCount int
		wantCodes []string
	}{
		{
			name:      "isme member role resolves with empty app filter",
			userID:    "user-isme",
			appID:     "",
			wantCount: 4,
			wantCodes: memberReadCodes,
		},
		{
			name:      "isme member role resolves with matching app id",
			userID:    "user-isme",
			appID:     isme,
			wantCount: 4,
			wantCodes: memberReadCodes,
		},
		{
			name:      "isme role does not resolve for another app",
			userID:    "user-isme",
			appID:     "app_other",
			wantCount: 0,
		},
		{
			name:      "soft-deleted role is excluded",
			userID:    "user-deleted-role",
			appID:     "",
			wantCount: 0,
		},
		{
			name:      "admin holds the full isme catalog",
			userID:    "user-admin",
			appID:     isme,
			wantCount: 21,
			wantCodes: []string{"user:read", "user:delete", "user:verify", "role:assign", "app_service:rotate_secret", "user_session:revoke", "settings:read", "settings:update"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := roleRepository.GetPermissionCodesByUserID(ctx, tt.userID, tt.appID)
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

func TestGetPermissionCodesGroupedByApp(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	roleRepository := NewRepository(db)

	isme := roleConstants.APP_ID_ISME
	insertApp(t, db, "app_medioa2", "medioa2", []string{"object", "bucket"})

	insertUser(t, db, "user-multi")
	assignRole(t, db, "user-multi", "rol_member", isme)            // isme: read perms
	assignRole(t, db, "user-multi", "rol_medioa2_admin", "app_medioa2") // medioa2: object/bucket read

	grouped, err := roleRepository.GetPermissionCodesGroupedByApp(ctx, "user-multi")
	if err != nil {
		t.Fatalf("GetPermissionCodesGroupedByApp() error = %v", err)
	}

	if _, ok := grouped["isme"]; !ok {
		t.Errorf("expected an isme group, got keys %v", keysOf(grouped))
	}
	if !slices.Contains(grouped["isme"], "user:read") {
		t.Errorf("expected isme group to contain user:read, got %v", grouped["isme"])
	}
	medioa2 := grouped["medioa2"]
	if !slices.Contains(medioa2, "object:read") || !slices.Contains(medioa2, "bucket:read") {
		t.Errorf("expected medioa2 group to contain object:read + bucket:read, got %v", medioa2)
	}
}

func TestGetAppCodesByUserID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	roleRepository := NewRepository(db)

	isme := roleConstants.APP_ID_ISME
	insertApp(t, db, "app_medioa2", "medioa2", []string{"object"})

	insertUser(t, db, "user-multi")
	assignRole(t, db, "user-multi", "rol_member", isme)
	assignRole(t, db, "user-multi", "rol_medioa2_admin", "app_medioa2")

	appCodes, err := roleRepository.GetAppCodesByUserID(ctx, "user-multi")
	if err != nil {
		t.Fatalf("GetAppCodesByUserID() error = %v", err)
	}
	for _, want := range []string{"isme", "medioa2"} {
		if !slices.Contains(appCodes, want) {
			t.Errorf("expected app codes to contain %q, got %v", want, appCodes)
		}
	}
}

// CreatePermissions stores the icon on a brand-new resource and reuses it for
// later rows of the same resource (never overwriting), so a resource keeps one
// consistent icon and ListPermissions reports it on every row.
func TestCreatePermissionsIconPerResource(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	roleRepository := NewRepository(db)

	insertApp(t, db, "app_medioa2", "medioa2", nil)

	// new resource takes the requested icon
	if _, err := roleRepository.CreatePermissions(ctx, "app_medioa2", []models.PermissionItem{
		{Resource: "report", Action: "read", Icon: "file"},
	}); err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}

	// a later row of the same resource requests a different icon — the stored
	// icon must win (existing-resource lock)
	if _, err := roleRepository.CreatePermissions(ctx, "app_medioa2", []models.PermissionItem{
		{Resource: "report", Action: "export", Icon: "database"},
	}); err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}

	permissions, err := roleRepository.ListPermissions(ctx, models.ListPermissionsRequest{AppID: "app_medioa2"})
	if err != nil {
		t.Fatalf("ListPermissions() error = %v", err)
	}
	if len(permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(permissions))
	}
	for _, permission := range permissions {
		if permission.Icon != "file" {
			t.Errorf("expected icon %q for %s:%s, got %q", "file", permission.Resource, permission.Action, permission.Icon)
		}
	}
}

// CreatePermissions stores the color on a brand-new resource and reuses it for
// later rows of the same resource (never overwriting), exactly like the icon, so
// a resource keeps one consistent color and ListPermissions reports it on every
// row.
func TestCreatePermissionsColorPerResource(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	roleRepository := NewRepository(db)

	insertApp(t, db, "app_medioa2", "medioa2", nil)

	// new resource takes the requested color
	if _, err := roleRepository.CreatePermissions(ctx, "app_medioa2", []models.PermissionItem{
		{Resource: "report", Action: "read", Color: "violet"},
	}); err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}

	// a later row of the same resource requests a different color — the stored
	// color must win (existing-resource lock)
	if _, err := roleRepository.CreatePermissions(ctx, "app_medioa2", []models.PermissionItem{
		{Resource: "report", Action: "export", Color: "rose"},
	}); err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}

	permissions, err := roleRepository.ListPermissions(ctx, models.ListPermissionsRequest{AppID: "app_medioa2"})
	if err != nil {
		t.Fatalf("ListPermissions() error = %v", err)
	}
	if len(permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(permissions))
	}
	for _, permission := range permissions {
		if permission.Color != "violet" {
			t.Errorf("expected color %q for %s:%s, got %q", "violet", permission.Resource, permission.Action, permission.Color)
		}
	}
}

func keysOf(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
