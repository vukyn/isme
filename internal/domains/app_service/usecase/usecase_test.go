package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/config"
	"github.com/vukyn/isme/internal/domains/app_service/constants"
	"github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/app_service/models"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleModels "github.com/vukyn/isme/internal/domains/role/models"
	roleUsecase "github.com/vukyn/isme/internal/domains/role/usecase"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"

	"github.com/vukyn/kuery/cryp/aes"
	pkgCtx "github.com/vukyn/kuery/ctx"
)

const testAESSecret = "test-aes-secret"

// === Fakes ===

type fakeAppServiceRepository struct {
	appServicesByID   map[string]entity.AppService
	appServicesByCode map[string]entity.AppService
	listResult        []entity.AppService
	listTotal         int64
	listRequests      []models.ListRequest
	updatedStatuses   map[string]int32
	updatedSecrets    map[string]string
}

var _ appServiceRepo.IRepository = (*fakeAppServiceRepository)(nil)

func newFakeAppServiceRepository() *fakeAppServiceRepository {
	return &fakeAppServiceRepository{
		appServicesByID:   map[string]entity.AppService{},
		appServicesByCode: map[string]entity.AppService{},
		updatedStatuses:   map[string]int32{},
		updatedSecrets:    map[string]string{},
	}
}

func (f *fakeAppServiceRepository) Create(ctx context.Context, req entity.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeAppServiceRepository) GetByID(ctx context.Context, id string) (entity.AppService, error) {
	return f.appServicesByID[id], nil
}

func (f *fakeAppServiceRepository) GetByIDs(ctx context.Context, ids []string) (map[string]entity.AppService, error) {
	result := map[string]entity.AppService{}
	for _, id := range ids {
		if app, ok := f.appServicesByID[id]; ok {
			result[id] = app
		}
	}
	return result, nil
}

func (f *fakeAppServiceRepository) GetByCode(ctx context.Context, code string) (entity.AppService, error) {
	return f.appServicesByCode[code], nil
}

func (f *fakeAppServiceRepository) Update(ctx context.Context, req entity.UpdateRequest) error {
	if req.AppSecret != nil {
		f.updatedSecrets[req.ID] = *req.AppSecret
	}
	return nil
}

func (f *fakeAppServiceRepository) List(ctx context.Context, req models.ListRequest) ([]entity.AppService, int64, error) {
	f.listRequests = append(f.listRequests, req)
	return f.listResult, f.listTotal, nil
}

func (f *fakeAppServiceRepository) UpdateStatus(ctx context.Context, id string, status int32) error {
	f.updatedStatuses[id] = status
	return nil
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

// === Helpers ===

type fakeRoleUsecase struct {
	provisionedAppIDs []string
}

var _ roleUsecase.IUseCase = (*fakeRoleUsecase)(nil)

func (f *fakeRoleUsecase) List(ctx context.Context, req roleModels.ListRequest) ([]roleModels.RoleListItem, error) {
	return nil, nil
}

func (f *fakeRoleUsecase) Create(ctx context.Context, req roleModels.CreateRequest) (roleModels.CreateResponse, error) {
	return roleModels.CreateResponse{}, nil
}

func (f *fakeRoleUsecase) GetDetail(ctx context.Context, id string) (roleModels.RoleDetailResponse, error) {
	return roleModels.RoleDetailResponse{}, nil
}

func (f *fakeRoleUsecase) Update(ctx context.Context, id string, req roleModels.UpdateRequest) error {
	return nil
}

func (f *fakeRoleUsecase) Delete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRoleUsecase) SetPermissions(ctx context.Context, id string, req roleModels.SetPermissionsRequest) error {
	return nil
}

func (f *fakeRoleUsecase) ListPermissions(ctx context.Context, req roleModels.ListPermissionsRequest) ([]roleModels.PermissionItem, error) {
	return nil, nil
}

func (f *fakeRoleUsecase) ProvisionDefaultRoles(ctx context.Context, appID string) error {
	f.provisionedAppIDs = append(f.provisionedAppIDs, appID)
	return nil
}

func (f *fakeRoleUsecase) ListMembers(ctx context.Context, id string, req roleModels.ListMembersRequest) (roleModels.ListMembersResponse, error) {
	return roleModels.ListMembersResponse{}, nil
}

func (f *fakeRoleUsecase) AddMembers(ctx context.Context, id string, req roleModels.AddMembersRequest) error {
	return nil
}

func (f *fakeRoleUsecase) RemoveMember(ctx context.Context, id string, userID string, appServiceID *string) error {
	return nil
}

func newTestUsecase(fakeAppService *fakeAppServiceRepository) IUseCase {
	cfg := &config.Config{}
	cfg.AES.Secret = testAESSecret
	return NewUsecase(fakeAppService, &fakeUserRepository{}, &fakeRoleUsecase{}, cfg)
}

func encryptTestSecret(t *testing.T, plainSecret string, ctxInfo string) string {
	t.Helper()
	encrypted, err := aes.Encrypt(plainSecret, testAESSecret, ctxInfo)
	if err != nil {
		t.Fatalf("failed to encrypt test secret: %v", err)
	}
	return encrypted
}

// === Tests ===

func TestListApps(t *testing.T) {
	t.Run("mapping correctness", func(t *testing.T) {
		fakeAppService := newFakeAppServiceRepository()
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		updatedAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
		fakeAppService.listResult = []entity.AppService{
			{
				ID:          "app-1",
				AppCode:     "code-1",
				AppName:     "App One",
				AppSecret:   "encrypted-secret",
				RedirectURL: "https://one.local/callback",
				CtxInfo:     constants.CtxInfoAuthen,
				Status:      constants.AppServiceStatusActive,
				CreatedAt:   createdAt,
				CreatedBy:   "user-a",
				UpdatedAt:   updatedAt,
				UpdatedBy:   "user-b",
			},
			{
				ID:      "app-2",
				AppCode: "code-2",
				AppName: "App Two",
				Status:  constants.AppServiceStatusInactive,
			},
		}
		fakeAppService.listTotal = 2
		testUsecase := newTestUsecase(fakeAppService)

		response, err := testUsecase.ListApps(context.Background(), models.ListRequest{Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if response.Total != 2 {
			t.Errorf("total = %d, want 2", response.Total)
		}
		if response.Page != 1 {
			t.Errorf("page = %d, want 1", response.Page)
		}
		if len(response.Items) != 2 {
			t.Fatalf("items length = %d, want 2", len(response.Items))
		}
		first := response.Items[0]
		if first.ID != "app-1" || first.AppCode != "code-1" || first.AppName != "App One" {
			t.Errorf("unexpected first item identity: %+v", first)
		}
		if first.RedirectURL != "https://one.local/callback" || first.CtxInfo != constants.CtxInfoAuthen {
			t.Errorf("unexpected first item details: %+v", first)
		}
		if first.Status != constants.AppServiceStatusActive {
			t.Errorf("first status = %d, want %d", first.Status, constants.AppServiceStatusActive)
		}
		if first.CreatedAt != createdAt.Format(time.RFC3339) {
			t.Errorf("created_at = %q, want %q", first.CreatedAt, createdAt.Format(time.RFC3339))
		}
		if first.UpdatedAt != updatedAt.Format(time.RFC3339) {
			t.Errorf("updated_at = %q, want %q", first.UpdatedAt, updatedAt.Format(time.RFC3339))
		}
		if first.CreatedBy != "user-a" || first.UpdatedBy != "user-b" {
			t.Errorf("unexpected audit fields: %+v", first)
		}
		second := response.Items[1]
		if second.CreatedAt != "" || second.UpdatedAt != "" {
			t.Errorf("zero timestamps must map to empty strings, got created_at=%q updated_at=%q", second.CreatedAt, second.UpdatedAt)
		}
	})

	t.Run("page zero normalized to one", func(t *testing.T) {
		fakeAppService := newFakeAppServiceRepository()
		testUsecase := newTestUsecase(fakeAppService)

		response, err := testUsecase.ListApps(context.Background(), models.ListRequest{Page: 0, PageSize: 0})
		if err != nil {
			t.Fatalf("ListApps() error = %v", err)
		}
		if response.Page != 1 {
			t.Errorf("page = %d, want 1", response.Page)
		}
		if len(fakeAppService.listRequests) != 1 {
			t.Fatalf("repo List calls = %d, want 1", len(fakeAppService.listRequests))
		}
		normalized := fakeAppService.listRequests[0]
		if normalized.Page != 1 || normalized.PageSize != 10 {
			t.Errorf("repo received page=%d pageSize=%d, want 1/10", normalized.Page, normalized.PageSize)
		}
	})

	t.Run("invalid filters rejected", func(t *testing.T) {
		tests := []struct {
			name    string
			request models.ListRequest
		}{
			{"invalid status", models.ListRequest{Status: 4}},
			{"negative status", models.ListRequest{Status: -1}},
			{"invalid ctx_info", models.ListRequest{CtxInfo: "unknown"}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fakeAppService := newFakeAppServiceRepository()
				testUsecase := newTestUsecase(fakeAppService)

				_, err := testUsecase.ListApps(context.Background(), tt.request)
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if len(fakeAppService.listRequests) != 0 {
					t.Error("repo List was called despite rejection")
				}
			})
		}
	})
}

func TestUpdateAppStatus(t *testing.T) {
	tests := []struct {
		name          string
		appServiceID  string
		currentStatus int32
		newStatus     int32
		wantErr       string
		wantUpdated   bool
	}{
		{"active to inactive", "app-1", constants.AppServiceStatusActive, constants.AppServiceStatusInactive, "", true},
		{"inactive to active", "app-1", constants.AppServiceStatusInactive, constants.AppServiceStatusActive, "", true},
		{"active to terminated", "app-1", constants.AppServiceStatusActive, constants.AppServiceStatusTerminated, "", true},
		{"inactive to terminated", "app-1", constants.AppServiceStatusInactive, constants.AppServiceStatusTerminated, "", true},
		{"terminated is terminal", "app-1", constants.AppServiceStatusTerminated, constants.AppServiceStatusActive, "app service is terminated", false},
		{"terminated to terminated rejected", "app-1", constants.AppServiceStatusTerminated, constants.AppServiceStatusTerminated, "app service is terminated", false},
		{"unknown id rejected", "app-unknown", constants.AppServiceStatusActive, constants.AppServiceStatusInactive, "app service not found", false},
		{"status zero rejected", "app-1", constants.AppServiceStatusActive, 0, "invalid status, must be 1 (active), 2 (inactive) or 3 (terminated)", false},
		{"status out of range rejected", "app-1", constants.AppServiceStatusActive, 4, "invalid status, must be 1 (active), 2 (inactive) or 3 (terminated)", false},
		{"same status no-op", "app-1", constants.AppServiceStatusActive, constants.AppServiceStatusActive, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeAppService := newFakeAppServiceRepository()
			fakeAppService.appServicesByID["app-1"] = entity.AppService{
				ID:     "app-1",
				Status: tt.currentStatus,
			}
			testUsecase := newTestUsecase(fakeAppService)

			err := testUsecase.UpdateStatus(context.Background(), tt.appServiceID, models.UpdateStatusRequest{Status: tt.newStatus})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				if len(fakeAppService.updatedStatuses) != 0 {
					t.Error("status was updated despite rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateStatus() error = %v", err)
			}
			if tt.wantUpdated {
				if got := fakeAppService.updatedStatuses[tt.appServiceID]; got != tt.newStatus {
					t.Errorf("updated status = %d, want %d", got, tt.newStatus)
				}
				return
			}
			if len(fakeAppService.updatedStatuses) != 0 {
				t.Error("no-op update must not call the repository")
			}
		})
	}
}

func TestVerifyApp(t *testing.T) {
	tests := []struct {
		name   string
		status int32
		wantOk bool
	}{
		{"active app passes", constants.AppServiceStatusActive, true},
		{"inactive app fails", constants.AppServiceStatusInactive, false},
		{"terminated app fails", constants.AppServiceStatusTerminated, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plainSecret := "plain-secret"
			encryptedSecret := encryptTestSecret(t, plainSecret, constants.CtxInfoAuthen)

			fakeAppService := newFakeAppServiceRepository()
			fakeAppService.appServicesByCode["code-1"] = entity.AppService{
				ID:        "app-1",
				AppCode:   "code-1",
				AppSecret: encryptedSecret,
				CtxInfo:   constants.CtxInfoAuthen,
				Status:    tt.status,
			}
			testUsecase := newTestUsecase(fakeAppService)

			response, err := testUsecase.VerifyApp(context.Background(), models.VerifyRequest{
				AppCode:   "code-1",
				CtxInfo:   constants.CtxInfoAuthen,
				AppSecret: plainSecret,
			})
			if err != nil {
				t.Fatalf("VerifyApp() error = %v", err)
			}
			if response.Ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", response.Ok, tt.wantOk)
			}
		})
	}
}

func TestRefreshApp(t *testing.T) {
	tests := []struct {
		name    string
		status  int32
		wantErr string
	}{
		{"active app refreshed", constants.AppServiceStatusActive, ""},
		{"terminated app rejected", constants.AppServiceStatusTerminated, "app service is terminated"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptedSecret := encryptTestSecret(t, "plain-secret", constants.CtxInfoAuthen)

			fakeAppService := newFakeAppServiceRepository()
			fakeAppService.appServicesByCode["code-1"] = entity.AppService{
				ID:        "app-1",
				AppCode:   "code-1",
				AppSecret: encryptedSecret,
				CtxInfo:   constants.CtxInfoAuthen,
				Status:    tt.status,
				CreatedBy: "user-a",
			}
			testUsecase := newTestUsecase(fakeAppService)

			ctx := context.WithValue(context.Background(), pkgCtx.UserIDKey, "user-a")
			response, err := testUsecase.RefreshApp(ctx, models.RefreshRequest{
				AppCode:   "code-1",
				AppSecret: encryptedSecret,
				CtxInfo:   constants.CtxInfoAuthen,
			})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				if len(fakeAppService.updatedSecrets) != 0 {
					t.Error("secret was rotated despite rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("RefreshApp() error = %v", err)
			}
			if response.AppSecret == "" {
				t.Error("expected a new app_secret, got empty string")
			}
			if _, ok := fakeAppService.updatedSecrets["app-1"]; !ok {
				t.Error("rotated secret was not persisted")
			}
		})
	}
}
