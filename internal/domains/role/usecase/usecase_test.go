package usecase

import (
	"context"
	"slices"
	"strings"
	"testing"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	appServiceModels "github.com/vukyn/isme/internal/domains/app_service/models"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	"github.com/vukyn/isme/internal/domains/role/entity"
	"github.com/vukyn/isme/internal/domains/role/models"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
)

// === Fakes ===

type fakeRoleRepository struct {
	rolesByID           map[string]entity.Role
	permissionsByRole   map[string][]entity.Permission
	membersCount        map[string]int
	createID            string
	updatedIDs          []string
	deletedIDs          []string
	replacedPermissions map[string][]int64
}

var _ roleRepo.IRepository = (*fakeRoleRepository)(nil)

func newFakeRoleRepository() *fakeRoleRepository {
	return &fakeRoleRepository{
		rolesByID:           map[string]entity.Role{},
		permissionsByRole:   map[string][]entity.Permission{},
		membersCount:        map[string]int{},
		replacedPermissions: map[string][]int64{},
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
	return nil, nil
}

func (f *fakeRoleRepository) CreatePermissions(ctx context.Context, appID string, perms []models.PermissionItem) (map[string]int64, error) {
	ids := map[string]int64{}
	for i, perm := range perms {
		ids[perm.Resource+":"+perm.Action] = int64(i + 1)
	}
	return ids, nil
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

func TestProvisionDefaultRolesSeedsAdminRole(t *testing.T) {
	fakeRole := newFakeRoleRepository()
	fakeRole.createID = "rol_admin_new"

	err := newTestUsecase(fakeRole).ProvisionDefaultRoles(context.Background(), testAppID)
	if err != nil {
		t.Fatalf("ProvisionDefaultRoles() error = %v", err)
	}

	// the admin role must have received a non-empty CRUD permission set
	granted, ok := fakeRole.replacedPermissions["rol_admin_new"]
	if !ok {
		t.Fatal("expected the admin role to be granted permissions")
	}
	if len(granted) == 0 {
		t.Error("expected a non-empty CRUD permission seed for the admin role")
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
