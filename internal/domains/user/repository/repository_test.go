package repository

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	sqliteHistory "github.com/vukyn/isme/db/history/sqlite"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	"github.com/vukyn/isme/internal/domains/user/entity"
	"github.com/vukyn/isme/internal/domains/user/models"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/vukyn/kuery/cryp"
)

// newTestDB opens an in-memory SQLite database and applies every migration
// from db/history/sqlite, including the RBAC seed and the app-owned migrations
// (which seed the isme app `app_isme` plus system roles rol_admin/rol_member/rol_viewer).
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
	// created_by/updated_by/deleted_by columns are TEXT DEFAULT '' in the schema
	// but the User entity maps them to int64 (nullzero); insert NULL so the row
	// scans cleanly into the entity.
	_, err := db.Exec(`
		INSERT INTO users (id, name, email, password, status, created_by, updated_by, deleted_by)
		VALUES (?, ?, ?, 'secret', 1, NULL, NULL, NULL)
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

// insertApp seeds an additional app_service plus an `admin` role for it
// (id rol_<appCode>_admin, code 'admin'). Extra roles get added per test.
func insertApp(t *testing.T, db *bun.DB, appID, appCode string) {
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
}

// insertRole seeds an extra (non-admin) role on an existing app.
func insertRole(t *testing.T, db *bun.DB, roleID, appID, code string) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO roles (id, app_id, code, name, is_system) VALUES (?, ?, ?, ?, 0)
	`, roleID, appID, code, code); err != nil {
		t.Fatalf("insert role %s: %v", roleID, err)
	}
}

func userIDs(users []entity.User) []string {
	ids := make([]string, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}
	return ids
}

// TestList_AppRoleFilter exercises the app-scoped role filter that List applies
// against the user_roles -> roles -> app_services chain.
func TestList_AppRoleFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	userRepository := NewRepository(db)

	isme := roleConstants.APP_ID_ISME

	// isme app (seeded by migrations): one admin, one member.
	insertUser(t, db, "user-isme-admin")
	assignRole(t, db, "user-isme-admin", roleConstants.ROLE_ID_ADMIN, isme)
	insertUser(t, db, "user-isme-member")
	assignRole(t, db, "user-isme-member", roleConstants.ROLE_ID_MEMBER, isme)

	// second app: medioa2 with an admin role + an uploader role.
	insertApp(t, db, "app_medioa2", "medioa2")
	insertRole(t, db, "rol_medioa2_uploader", "app_medioa2", "uploader")
	insertUser(t, db, "user-medioa2-admin")
	assignRole(t, db, "user-medioa2-admin", "rol_medioa2_admin", "app_medioa2")
	insertUser(t, db, "user-medioa2-uploader")
	assignRole(t, db, "user-medioa2-uploader", "rol_medioa2_uploader", "app_medioa2")

	// a plain user with no roles at all.
	insertUser(t, db, "user-none")

	const allUsers = 5 // migrations seed no user rows; only the 5 inserted above exist.

	listReq := func(appCode, roleCode string) models.ListRequest {
		return models.ListRequest{Page: 1, PageSize: 100, AppCode: appCode, RoleID: roleCode}
	}

	tests := []struct {
		name        string
		appCode     string
		roleCode    string
		wantTotal   int64
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:      "no filter returns every user",
			appCode:   "",
			roleCode:  "",
			wantTotal: allUsers,
			wantPresent: []string{
				"user-isme-admin", "user-isme-member",
				"user-medioa2-admin", "user-medioa2-uploader", "user-none",
			},
		},
		{
			name:        "app only scopes to that app's members",
			appCode:     "medioa2",
			roleCode:    "",
			wantTotal:   2,
			wantPresent: []string{"user-medioa2-admin", "user-medioa2-uploader"},
			wantAbsent:  []string{"user-isme-admin", "user-isme-member", "user-none"},
		},
		{
			name:        "app plus role narrows to the matching role",
			appCode:     "medioa2",
			roleCode:    "admin",
			wantTotal:   1,
			wantPresent: []string{"user-medioa2-admin"},
			wantAbsent:  []string{"user-medioa2-uploader", "user-isme-admin"},
		},
		{
			name:        "role codes are scoped per app (cross-app isolation)",
			appCode:     "isme",
			roleCode:    "admin",
			wantTotal:   1,
			wantPresent: []string{"user-isme-admin"},
			wantAbsent:  []string{"user-medioa2-admin"},
		},
		{
			name:      "role without app is ignored (predicate dropped)",
			appCode:   "",
			roleCode:  "admin",
			wantTotal: allUsers,
			wantPresent: []string{
				"user-isme-admin", "user-isme-member",
				"user-medioa2-admin", "user-medioa2-uploader", "user-none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, total, err := userRepository.List(ctx, listReq(tt.appCode, tt.roleCode))
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d (ids: %v)", total, tt.wantTotal, userIDs(users))
			}
			if int64(len(users)) != tt.wantTotal {
				t.Errorf("len(users) = %d, want %d (ids: %v)", len(users), tt.wantTotal, userIDs(users))
			}
			ids := userIDs(users)
			for _, want := range tt.wantPresent {
				if !slices.Contains(ids, want) {
					t.Errorf("expected %q present, got %v", want, ids)
				}
			}
			for _, absent := range tt.wantAbsent {
				if slices.Contains(ids, absent) {
					t.Errorf("expected %q absent, got %v", absent, ids)
				}
			}
		})
	}

	// Soft-deleted roles must drop out of the filter: removing the uploader role
	// leaves user-medioa2-uploader (whose only role is uploader) unreachable via
	// the medioa2 app scope.
	t.Run("soft-deleted role is excluded", func(t *testing.T) {
		if _, err := db.Exec(`UPDATE roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = 'rol_medioa2_uploader'`); err != nil {
			t.Fatalf("soft delete uploader role: %v", err)
		}
		users, total, err := userRepository.List(ctx, listReq("medioa2", ""))
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if total != 1 {
			t.Errorf("total = %d, want 1 (ids: %v)", total, userIDs(users))
		}
		ids := userIDs(users)
		if !slices.Contains(ids, "user-medioa2-admin") {
			t.Errorf("expected user-medioa2-admin present, got %v", ids)
		}
		if slices.Contains(ids, "user-medioa2-uploader") {
			t.Errorf("expected user-medioa2-uploader excluded after soft delete, got %v", ids)
		}
	})
}
