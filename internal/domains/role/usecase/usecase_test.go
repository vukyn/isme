package usecase

import (
	"context"
	"slices"
	"strings"
	"testing"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	appServiceModels "github.com/vukyn/isme/internal/domains/app_service/models"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
	"github.com/vukyn/isme/internal/domains/role/entity"
	"github.com/vukyn/isme/internal/domains/role/models"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
)

// === Fakes ===

type fakeRoleRepository struct {
	rolesByID            map[string]entity.Role
	permissionsByID      map[int64]entity.Permission
	permissionsByRole    map[string][]entity.Permission
	membersCount         map[string]int
	createID             string
	updatedIDs           []string
	deletedIDs           []string
	replacedPermissions  map[string][]int64
	createdPermissions   map[string][]models.PermissionItem
	deletedPermissionIDs []int64
	updatedAppearances   []string
}

var _ roleRepo.IRepository = (*fakeRoleRepository)(nil)

func newFakeRoleRepository() *fakeRoleRepository {
	return &fakeRoleRepository{
		rolesByID:           map[string]entity.Role{},
		permissionsByID:     map[int64]entity.Permission{},
		permissionsByRole:   map[string][]entity.Permission{},
		membersCount:        map[string]int{},
		replacedPermissions: map[string][]int64{},
		createdPermissions:  map[string][]models.PermissionItem{},
	}
}

func (f *fakeRoleRepository) Create(ctx context.Context, req models.CreateRequest) (string, error) {
	return f.createID, nil
}

func (f *fakeRoleRepository) GetByID(ctx context.Context, id string) (entity.Role, error) {
	return f.rolesByID[id], nil
}

func (f *fakeRoleRepository) GetByAppAndCode(ctx context.Context, appID string, code string) (entity.Role, error) {
	for _, role := range f.rolesByID {
		if role.AppID == appID && role.Code == code {
			return role, nil
		}
	}
	return entity.Role{}, nil
}

func (f *fakeRoleRepository) List(ctx context.Context, req models.ListRequest) ([]models.RoleListItem, error) {
	return nil, nil
}

func (f *fakeRoleRepository) Update(ctx context.Context, id string, req models.UpdateRequest) error {
	f.updatedIDs = append(f.updatedIDs, id)
	return nil
}

func (f *fakeRoleRepository) SoftDelete(ctx context.Context, id string) error {
	f.deletedIDs = append(f.deletedIDs, id)
	return nil
}

func (f *fakeRoleRepository) ListPermissions(ctx context.Context, req models.ListPermissionsRequest) ([]entity.Permission, error) {
	items := []entity.Permission{}
	for _, perm := range f.createdPermissions[req.AppID] {
		items = append(items, entity.Permission{
			ID:       perm.ID,
			AppID:    req.AppID,
			Resource: perm.Resource,
			Action:   perm.Action,
			Icon:     perm.Icon,
			Color:    perm.Color,
		})
	}
	return items, nil
}

func (f *fakeRoleRepository) CreatePermissions(ctx context.Context, appID string, perms []models.PermissionItem) (map[string]int64, error) {
	// mirror the repo's per-resource icon/color rule: an existing resource keeps
	// its stored icon/color; a new resource takes the values on its first
	// incoming pair.
	iconByResource := map[string]string{}
	colorByResource := map[string]string{}
	for _, existing := range f.createdPermissions[appID] {
		if _, seen := iconByResource[existing.Resource]; !seen && existing.Icon != "" {
			iconByResource[existing.Resource] = existing.Icon
		}
		if _, seen := colorByResource[existing.Resource]; !seen && existing.Color != "" {
			colorByResource[existing.Resource] = existing.Color
		}
	}
	for _, perm := range perms {
		if _, resolved := iconByResource[perm.Resource]; !resolved {
			iconByResource[perm.Resource] = perm.Icon
		}
		if _, resolved := colorByResource[perm.Resource]; !resolved {
			colorByResource[perm.Resource] = perm.Color
		}
	}

	ids := map[string]int64{}
	for i, perm := range perms {
		id := int64(len(f.createdPermissions[appID]) + i + 1)
		f.createdPermissions[appID] = append(f.createdPermissions[appID], models.PermissionItem{
			ID:       id,
			AppID:    appID,
			Resource: perm.Resource,
			Action:   perm.Action,
			Icon:     iconByResource[perm.Resource],
			Color:    colorByResource[perm.Resource],
		})
		ids[perm.Resource+":"+perm.Action] = id
	}
	return ids, nil
}

func (f *fakeRoleRepository) GetPermissionByID(ctx context.Context, permissionID int64) (entity.Permission, error) {
	return f.permissionsByID[permissionID], nil
}

func (f *fakeRoleRepository) UpdatePermissionAppearance(ctx context.Context, appID string, resource string, icon string, color string) error {
	updated := false
	for i := range f.createdPermissions[appID] {
		if f.createdPermissions[appID][i].Resource == resource {
			f.createdPermissions[appID][i].Icon = icon
			f.createdPermissions[appID][i].Color = color
			updated = true
		}
	}
	if updated {
		f.updatedAppearances = append(f.updatedAppearances, appID+"\x00"+resource)
	}
	return nil
}

func (f *fakeRoleRepository) DeletePermission(ctx context.Context, permissionID int64) error {
	f.deletedPermissionIDs = append(f.deletedPermissionIDs, permissionID)
	return nil
}

func (f *fakeRoleRepository) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]entity.Permission, error) {
	return f.permissionsByRole[roleID], nil
}

func (f *fakeRoleRepository) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []int64) error {
	f.replacedPermissions[roleID] = permissionIDs
	return nil
}

func (f *fakeRoleRepository) ListMembers(ctx context.Context, roleID string, req models.ListMembersRequest) ([]models.MemberItem, int, error) {
	return nil, 0, nil
}

func (f *fakeRoleRepository) CountMembersByRoleID(ctx context.Context, roleID string) (int, error) {
	return f.membersCount[roleID], nil
}

func (f *fakeRoleRepository) AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error {
	return nil
}

func (f *fakeRoleRepository) RemoveMember(ctx context.Context, roleID string, userID string, appServiceID *string) error {
	return nil
}

func (f *fakeRoleRepository) GetPermissionCodesByUserID(ctx context.Context, userID string, appID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionCodesGroupedByApp(ctx context.Context, userID string) (map[string][]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetAppCodesByUserID(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionCodesByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	return map[string][]string{}, nil
}

func (f *fakeRoleRepository) GetRoleCodesGroupedByAppByUserIDs(ctx context.Context, userIDs []string) (map[string][]models.UserAppRole, error) {
	return map[string][]models.UserAppRole{}, nil
}

type fakeUserRepository struct{}

var _ userRepo.IRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Create(ctx context.Context, req userModels.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id string) (userEntity.User, error) {
	return userEntity.User{}, nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (userEntity.User, error) {
	return userEntity.User{}, nil
}

func (f *fakeUserRepository) SetPassword(ctx context.Context, id string, password string) error {
	return nil
}

func (f *fakeUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) Verify(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) List(ctx context.Context, req userModels.ListRequest) ([]userEntity.User, int64, error) {
	return nil, 0, nil
}

func (f *fakeUserRepository) UpdateStatus(ctx context.Context, id string, status int32) error {
	return nil
}

func (f *fakeUserRepository) SoftDelete(ctx context.Context, id string) error {
	return nil
}

type fakeAppServiceRepository struct {
	appServicesByID map[string]appServiceEntity.AppService
}

var _ appServiceRepo.IRepository = (*fakeAppServiceRepository)(nil)

func (f *fakeAppServiceRepository) Create(ctx context.Context, req appServiceEntity.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeAppServiceRepository) GetByID(ctx context.Context, id string) (appServiceEntity.AppService, error) {
	return f.appServicesByID[id], nil
}

func (f *fakeAppServiceRepository) GetByIDs(ctx context.Context, ids []string) (map[string]appServiceEntity.AppService, error) {
	result := map[string]appServiceEntity.AppService{}
	for _, id := range ids {
		if app, ok := f.appServicesByID[id]; ok {
			result[id] = app
		}
	}
	return result, nil
}

func (f *fakeAppServiceRepository) GetByCode(ctx context.Context, code string) (appServiceEntity.AppService, error) {
	return appServiceEntity.AppService{}, nil
}

func (f *fakeAppServiceRepository) Update(ctx context.Context, req appServiceEntity.UpdateRequest) error {
	return nil
}

func (f *fakeAppServiceRepository) List(ctx context.Context, req appServiceModels.ListRequest) ([]appServiceEntity.AppService, int64, error) {
	return nil, 0, nil
}

func (f *fakeAppServiceRepository) UpdateStatus(ctx context.Context, id string, status int32) error {
	return nil
}

const testAppID = "app_test"

func newTestUsecase(fakeRole *fakeRoleRepository) IUseCase {
	fakeAppService := &fakeAppServiceRepository{
		appServicesByID: map[string]appServiceEntity.AppService{
			testAppID: {ID: testAppID, AppCode: "test"},
		},
	}
	return NewUsecase(fakeRole, &fakeUserRepository{}, fakeAppService)
}

// === Tests ===

func TestSystemRoleImmutability(t *testing.T) {
	tests := []struct {
		name      string
		operation func(u IUseCase, ctx context.Context) error
		wantErr   string
	}{
		{
			name: "update rejected",
			operation: func(u IUseCase, ctx context.Context) error {
				return u.Update(ctx, "rol_admin", models.UpdateRequest{Name: "Renamed"})
			},
			wantErr: "system role cannot be modified",
		},
		{
			name: "delete rejected",
			operation: func(u IUseCase, ctx context.Context) error {
				return u.Delete(ctx, "rol_admin")
			},
			wantErr: "system role cannot be deleted",
		},
		{
			name: "set permissions rejected",
			operation: func(u IUseCase, ctx context.Context) error {
				return u.SetPermissions(ctx, "rol_admin", models.SetPermissionsRequest{PermissionIDs: []int64{1}})
			},
			wantErr: "system role cannot be modified",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRole := newFakeRoleRepository()
			fakeRole.rolesByID["rol_admin"] = entity.Role{ID: "rol_admin", Code: "admin", Name: "Admin", IsSystem: true}

			err := tt.operation(newTestUsecase(fakeRole), context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
			if len(fakeRole.updatedIDs) != 0 || len(fakeRole.deletedIDs) != 0 || len(fakeRole.replacedPermissions) != 0 {
				t.Error("repository was mutated for a system role")
			}
		})
	}
}

func TestDeleteRoleWithMembersRejected(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeRole.rolesByID["rol_custom"] = entity.Role{ID: "rol_custom", Code: "custom", Name: "Custom"}
	fakeRole.membersCount["rol_custom"] = 3

	err := newTestUsecase(fakeRole).Delete(context.Background(), "rol_custom")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "role has members") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "role has members")
	}
	if len(fakeRole.deletedIDs) != 0 {
		t.Error("SoftDelete was called for a role with members")
	}
}

func TestCreateWithCloneCopiesPermissions(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeRole.rolesByID["rol_source"] = entity.Role{ID: "rol_source", AppID: testAppID, Code: "source", Name: "Source"}
	fakeRole.permissionsByRole["rol_source"] = []entity.Permission{
		{ID: 1, Resource: "user", Action: "read"},
		{ID: 2, Resource: "role", Action: "read"},
		{ID: 3, Resource: "app_service", Action: "read"},
	}
	fakeRole.createID = "rol_new"

	resp, err := newTestUsecase(fakeRole).Create(context.Background(), models.CreateRequest{
		AppID:           testAppID,
		Code:            "support-team",
		Name:            "Support Team",
		CloneFromRoleID: "rol_source",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if resp.ID != "rol_new" {
		t.Errorf("Create() ID = %q, want %q", resp.ID, "rol_new")
	}
	if got, want := fakeRole.replacedPermissions["rol_new"], []int64{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("cloned permissions = %v, want %v", got, want)
	}
}

func TestCreateRejectsCrossAppClone(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	// clone source belongs to a different app than the new role's target app
	fakeRole.rolesByID["rol_other"] = entity.Role{ID: "rol_other", AppID: "app_other", Code: "other", Name: "Other"}
	fakeRole.createID = "rol_new"

	_, err := newTestUsecase(fakeRole).Create(context.Background(), models.CreateRequest{
		AppID:           testAppID,
		Code:            "support-team",
		Name:            "Support Team",
		CloneFromRoleID: "rol_other",
	})
	if err == nil {
		t.Fatal("expected error for cross-app clone, got nil")
	}
	if !strings.Contains(err.Error(), "same app") {
		t.Errorf("error = %q, want it to mention same app", err.Error())
	}
	if len(fakeRole.replacedPermissions) != 0 {
		t.Error("permissions were copied despite cross-app clone rejection")
	}
}

func TestProvisionDefaultRolesSeedsEmptyAdminRole(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeRole.createID = "rol_admin_new"

	err := newTestUsecase(fakeRole).ProvisionDefaultRoles(context.Background(), testAppID)
	if err != nil {
		t.Fatalf("ProvisionDefaultRoles() error = %v", err)
	}

	// the admin role is created with ZERO permissions — none are seeded on app create
	if len(fakeRole.replacedPermissions) != 0 {
		t.Error("expected no permissions to be seeded for the admin role")
	}
	if len(fakeRole.createdPermissions) != 0 {
		t.Error("expected no permissions to be created on app provisioning")
	}
}

func TestProvisionDefaultRolesIsIdempotent(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	// admin role already exists for the app
	fakeRole.rolesByID["rol_existing"] = entity.Role{ID: "rol_existing", AppID: testAppID, Code: "admin", Name: "Admin"}

	err := newTestUsecase(fakeRole).ProvisionDefaultRoles(context.Background(), testAppID)
	if err != nil {
		t.Fatalf("ProvisionDefaultRoles() error = %v", err)
	}
	if len(fakeRole.replacedPermissions) != 0 {
		t.Error("expected no provisioning when the admin role already exists")
	}
}

func TestCreatePermissionsForNormalApp(t *testing.T) {
	fakeRole := newFakeRoleRepository()

	items, err := newTestUsecase(fakeRole).CreatePermissions(context.Background(), models.CreatePermissionsRequest{
		AppID: testAppID,
		Permissions: []models.PermissionPair{
			{Resource: "report", Action: "read"},
			{Resource: "report", Action: "export"},
		},
	})
	if err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 created permission items, got %d", len(items))
	}
	for _, item := range items {
		if item.ID == 0 {
			t.Errorf("expected created permission %s:%s to have an id", item.Resource, item.Action)
		}
		if item.AppID != testAppID {
			t.Errorf("expected app_id %q, got %q", testAppID, item.AppID)
		}
	}
	if got := fakeRole.createdPermissions[testAppID]; len(got) != 2 {
		t.Errorf("expected 2 permissions persisted, got %d", len(got))
	}
}

// A brand-new resource takes the icon supplied on its first pair, and every
// returned row of that resource reports it.
func TestCreatePermissionsStoresIconForNewResource(t *testing.T) {
	fakeRole := newFakeRoleRepository()

	items, err := newTestUsecase(fakeRole).CreatePermissions(context.Background(), models.CreatePermissionsRequest{
		AppID: testAppID,
		Permissions: []models.PermissionPair{
			{Resource: "report", Action: "read", Icon: "file"},
			{Resource: "report", Action: "export", Icon: "file"},
		},
	})
	if err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}
	for _, item := range items {
		if item.Icon != "file" {
			t.Errorf("expected icon %q for %s:%s, got %q", "file", item.Resource, item.Action, item.Icon)
		}
	}
}

// When a resource already exists, new pairs reuse that resource's stored icon
// and ignore any icon supplied on the request (existing-resource lock).
func TestCreatePermissionsReusesExistingResourceIcon(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	usecase := newTestUsecase(fakeRole)

	if _, err := usecase.CreatePermissions(context.Background(), models.CreatePermissionsRequest{
		AppID:       testAppID,
		Permissions: []models.PermissionPair{{Resource: "report", Action: "read", Icon: "file"}},
	}); err != nil {
		t.Fatalf("seed CreatePermissions() error = %v", err)
	}

	// add another action to the same resource but request a DIFFERENT icon
	items, err := usecase.CreatePermissions(context.Background(), models.CreatePermissionsRequest{
		AppID:       testAppID,
		Permissions: []models.PermissionPair{{Resource: "report", Action: "export", Icon: "database"}},
	})
	if err != nil {
		t.Fatalf("CreatePermissions() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 returned item, got %d", len(items))
	}
	if items[0].Icon != "file" {
		t.Errorf("expected the existing resource icon %q to be reused, got %q", "file", items[0].Icon)
	}
}

// An unknown icon key is rejected by validation; the empty icon is allowed.
func TestCreatePermissionsIconValidation(t *testing.T) {
	tests := []struct {
		name    string
		icon    string
		wantErr bool
	}{
		{"known key", "database", false},
		{"empty allowed", "", false},
		{"unknown key rejected", "rocket", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRole := newFakeRoleRepository()
			_, err := newTestUsecase(fakeRole).CreatePermissions(context.Background(), models.CreatePermissionsRequest{
				AppID:       testAppID,
				Permissions: []models.PermissionPair{{Resource: "report", Action: "read", Icon: tt.icon}},
			})
			if tt.wantErr && err == nil {
				t.Errorf("CreatePermissions(icon=%q) expected error, got nil", tt.icon)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CreatePermissions(icon=%q) unexpected error: %v", tt.icon, err)
			}
		})
	}
}

func TestCreatePermissionsRejectsIsmeSystemApp(t *testing.T) {
	fakeRole := newFakeRoleRepository()

	_, err := newTestUsecase(fakeRole).CreatePermissions(context.Background(), models.CreatePermissionsRequest{
		AppID:       roleConstants.APP_ID_ISME,
		Permissions: []models.PermissionPair{{Resource: "report", Action: "read"}},
	})
	if err == nil {
		t.Fatal("expected error creating permissions for the isme system app, got nil")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Errorf("error = %q, want it to mention read-only", err.Error())
	}
	if len(fakeRole.createdPermissions) != 0 {
		t.Error("permissions were created for the isme system app despite the guard")
	}
}

func TestDeletePermissionForNormalApp(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeRole.permissionsByID[42] = entity.Permission{ID: 42, AppID: testAppID, Resource: "report", Action: "read"}

	err := newTestUsecase(fakeRole).DeletePermission(context.Background(), 42)
	if err != nil {
		t.Fatalf("DeletePermission() error = %v", err)
	}
	if got := fakeRole.deletedPermissionIDs; len(got) != 1 || got[0] != 42 {
		t.Errorf("deleted permission ids = %v, want [42] (delete + grant cleanup runs in the repo)", got)
	}
}

func TestDeletePermissionRejectsIsmeSystemApp(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeRole.permissionsByID[7] = entity.Permission{ID: 7, AppID: roleConstants.APP_ID_ISME, Resource: "user", Action: "read"}

	err := newTestUsecase(fakeRole).DeletePermission(context.Background(), 7)
	if err == nil {
		t.Fatal("expected error deleting an isme system-app permission, got nil")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Errorf("error = %q, want it to mention read-only", err.Error())
	}
	if len(fakeRole.deletedPermissionIDs) != 0 {
		t.Error("permission was deleted for the isme system app despite the guard")
	}
}

func TestDeletePermissionNotFound(t *testing.T) {
	fakeRole := newFakeRoleRepository()

	err := newTestUsecase(fakeRole).DeletePermission(context.Background(), 999)
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to mention not found", err.Error())
	}
	if len(fakeRole.deletedPermissionIDs) != 0 {
		t.Error("a missing permission was deleted")
	}
}

func TestCreatePermissionsValidation(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		wantErr  bool
	}{
		{"valid", "report", "read", false},
		{"underscore allowed", "audit_log", "read", false},
		{"empty resource", "", "read", true},
		{"empty action", "report", "", true},
		{"colon in resource rejected", "report:thing", "read", true},
		{"colon in action rejected", "report", "read:all", true},
		{"uppercase rejected", "Report", "read", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRole := newFakeRoleRepository()
			_, err := newTestUsecase(fakeRole).CreatePermissions(context.Background(), models.CreatePermissionsRequest{
				AppID:       testAppID,
				Permissions: []models.PermissionPair{{Resource: tt.resource, Action: tt.action}},
			})
			if tt.wantErr && err == nil {
				t.Errorf("CreatePermissions(%q:%q) expected error, got nil", tt.resource, tt.action)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CreatePermissions(%q:%q) unexpected error: %v", tt.resource, tt.action, err)
			}
		})
	}
}

// seedResource adds a resource (one read row) to the fake catalog so appearance
// edits have something to target.
func seedResource(t *testing.T, fakeRole *fakeRoleRepository, appID, resource string) {
	t.Helper()
	if _, err := newTestUsecase(fakeRole).CreatePermissions(context.Background(), models.CreatePermissionsRequest{
		AppID:       appID,
		Permissions: []models.PermissionPair{{Resource: resource, Action: "read"}},
	}); err != nil {
		t.Fatalf("seed resource %q: %v", resource, err)
	}
}

// The happy path updates the resource's appearance and reaches the repo.
func TestUpdatePermissionAppearanceHappyPath(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	seedResource(t, fakeRole, testAppID, "report")

	err := newTestUsecase(fakeRole).UpdatePermissionAppearance(context.Background(), models.UpdatePermissionAppearanceRequest{
		AppID:    testAppID,
		Resource: "report",
		Icon:     "file",
		Color:    "violet",
	})
	if err != nil {
		t.Fatalf("UpdatePermissionAppearance() error = %v", err)
	}
	if len(fakeRole.updatedAppearances) != 1 {
		t.Errorf("expected the appearance to be updated once, got %v", fakeRole.updatedAppearances)
	}
}

// An empty icon + empty color is allowed (both validators permit empty) and
// still reaches the repo.
func TestUpdatePermissionAppearanceAllowsEmpty(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	seedResource(t, fakeRole, testAppID, "report")

	err := newTestUsecase(fakeRole).UpdatePermissionAppearance(context.Background(), models.UpdatePermissionAppearanceRequest{
		AppID:    testAppID,
		Resource: "report",
		Icon:     "",
		Color:    "",
	})
	if err != nil {
		t.Fatalf("UpdatePermissionAppearance() with empty icon/color error = %v", err)
	}
	if len(fakeRole.updatedAppearances) != 1 {
		t.Errorf("expected the appearance to be updated once, got %v", fakeRole.updatedAppearances)
	}
}

// Editing the isme system app is rejected by app id.
func TestUpdatePermissionAppearanceRejectsIsmeByID(t *testing.T) {
	fakeRole := newFakeRoleRepository()

	err := newTestUsecase(fakeRole).UpdatePermissionAppearance(context.Background(), models.UpdatePermissionAppearanceRequest{
		AppID:    roleConstants.APP_ID_ISME,
		Resource: "user",
		Icon:     "user",
		Color:    "violet",
	})
	if err == nil {
		t.Fatal("expected error editing the isme system app, got nil")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Errorf("error = %q, want it to mention read-only", err.Error())
	}
	if len(fakeRole.updatedAppearances) != 0 {
		t.Error("appearance was updated for the isme system app despite the guard")
	}
}

// Editing an app that resolves to the isme app_code is rejected (defense in depth).
func TestUpdatePermissionAppearanceRejectsIsmeByCode(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeAppService := &fakeAppServiceRepository{
		appServicesByID: map[string]appServiceEntity.AppService{
			// a non-isme id that nonetheless resolves to the isme app_code
			"app_alias": {ID: "app_alias", AppCode: roleConstants.APP_CODE_ISME},
		},
	}
	uc := NewUsecase(fakeRole, &fakeUserRepository{}, fakeAppService)

	err := uc.UpdatePermissionAppearance(context.Background(), models.UpdatePermissionAppearanceRequest{
		AppID:    "app_alias",
		Resource: "user",
		Icon:     "user",
		Color:    "violet",
	})
	if err == nil {
		t.Fatal("expected error editing an app aliased to the isme app_code, got nil")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Errorf("error = %q, want it to mention read-only", err.Error())
	}
	if len(fakeRole.updatedAppearances) != 0 {
		t.Error("appearance was updated for the isme-aliased app despite the guard")
	}
}

// A resource that is not in the app catalog is rejected with not-found.
func TestUpdatePermissionAppearanceUnknownResource(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	seedResource(t, fakeRole, testAppID, "report")

	err := newTestUsecase(fakeRole).UpdatePermissionAppearance(context.Background(), models.UpdatePermissionAppearanceRequest{
		AppID:    testAppID,
		Resource: "ghost",
		Icon:     "file",
		Color:    "violet",
	})
	if err == nil {
		t.Fatal("expected not-found error for an unknown resource, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to mention not found", err.Error())
	}
	if len(fakeRole.updatedAppearances) != 0 {
		t.Error("appearance was updated for a resource missing from the catalog")
	}
}

// Unknown icon and unknown color keys are rejected by validation.
func TestUpdatePermissionAppearanceValidation(t *testing.T) {
	tests := []struct {
		name    string
		icon    string
		color   string
		wantErr bool
	}{
		{"known icon + color", "file", "violet", false},
		{"empty icon + color allowed", "", "", false},
		{"unknown icon rejected", "rocket", "violet", true},
		{"unknown color rejected", "file", "neon", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRole := newFakeRoleRepository()
			seedResource(t, fakeRole, testAppID, "report")
			err := newTestUsecase(fakeRole).UpdatePermissionAppearance(context.Background(), models.UpdatePermissionAppearanceRequest{
				AppID:    testAppID,
				Resource: "report",
				Icon:     tt.icon,
				Color:    tt.color,
			})
			if tt.wantErr && err == nil {
				t.Errorf("UpdatePermissionAppearance(icon=%q,color=%q) expected error, got nil", tt.icon, tt.color)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("UpdatePermissionAppearance(icon=%q,color=%q) unexpected error: %v", tt.icon, tt.color, err)
			}
		})
	}
}

func TestCreateCodeSlugValidation(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid simple slug", "support", false},
		{"valid slug with hyphen and digits", "tier2-support", false},
		{"empty code", "", true},
		{"uppercase rejected", "Support", true},
		{"space rejected", "support team", true},
		{"leading hyphen rejected", "-support", true},
		{"special characters rejected", "support@team", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRole := newFakeRoleRepository()
			fakeRole.createID = "rol_new"

			_, err := newTestUsecase(fakeRole).Create(context.Background(), models.CreateRequest{
				AppID: testAppID,
				Code:  tt.code,
				Name:  "Some Role",
			})
			if tt.wantErr && err == nil {
				t.Errorf("Create(code=%q) expected error, got nil", tt.code)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Create(code=%q) unexpected error: %v", tt.code, err)
			}
		})
	}
}
