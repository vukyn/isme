package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vukyn/isme/internal/config"
	roleEntity "github.com/vukyn/isme/internal/domains/role/entity"
	roleModels "github.com/vukyn/isme/internal/domains/role/models"
	userEntity "github.com/vukyn/isme/internal/domains/user/entity"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	"github.com/vukyn/isme/internal/domains/user_invitation/constants"
	"github.com/vukyn/isme/internal/domains/user_invitation/entity"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"

	"github.com/vukyn/kuery/cryp"
)

// === user repository fake ===

type createUserCall struct {
	name  string
	email string
}

type setPasswordCall struct {
	id       string
	password string
}

type fakeUserRepository struct {
	userByEmail      userEntity.User
	createErr        error
	createCalls      []createUserCall
	setPasswordCalls []setPasswordCall
	verifyCalls      []string
}

func (f *fakeUserRepository) Create(ctx context.Context, req userModels.CreateRequest) (string, error) {
	if f.createErr != nil {
		return "", f.createErr
	}
	f.createCalls = append(f.createCalls, createUserCall{name: req.Name, email: req.Email})
	return "user-1", nil
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id string) (userEntity.User, error) {
	return userEntity.User{}, nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (userEntity.User, error) {
	return f.userByEmail, nil
}

func (f *fakeUserRepository) SetPassword(ctx context.Context, id string, password string) error {
	f.setPasswordCalls = append(f.setPasswordCalls, setPasswordCall{id: id, password: password})
	return nil
}

func (f *fakeUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) PromoteAdmin(ctx context.Context, id string) error {
	return nil
}

func (f *fakeUserRepository) IsAdmin(ctx context.Context, id string) (bool, error) {
	return false, nil
}

func (f *fakeUserRepository) Verify(ctx context.Context, id string) error {
	f.verifyCalls = append(f.verifyCalls, id)
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

// === role repository fake ===

type addMembersCall struct {
	roleID       string
	userIDs      []string
	appServiceID *string
}

type fakeRoleRepository struct {
	roles           map[string]roleEntity.Role
	addMembersCalls []addMembersCall
}

func (f *fakeRoleRepository) Create(ctx context.Context, req roleModels.CreateRequest) (string, error) {
	return "", nil
}

func (f *fakeRoleRepository) GetByID(ctx context.Context, id string) (roleEntity.Role, error) {
	return f.roles[id], nil
}

func (f *fakeRoleRepository) GetByCode(ctx context.Context, code string) (roleEntity.Role, error) {
	return roleEntity.Role{}, nil
}

func (f *fakeRoleRepository) List(ctx context.Context) ([]roleModels.RoleListItem, error) {
	return nil, nil
}

func (f *fakeRoleRepository) Update(ctx context.Context, id string, req roleModels.UpdateRequest) error {
	return nil
}

func (f *fakeRoleRepository) SoftDelete(ctx context.Context, id string) error {
	return nil
}

func (f *fakeRoleRepository) ListPermissions(ctx context.Context) ([]roleEntity.Permission, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]roleEntity.Permission, error) {
	return nil, nil
}

func (f *fakeRoleRepository) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []int64) error {
	return nil
}

func (f *fakeRoleRepository) ListMembers(ctx context.Context, roleID string, req roleModels.ListMembersRequest) ([]roleModels.MemberItem, int, error) {
	return nil, 0, nil
}

func (f *fakeRoleRepository) CountMembersByRoleID(ctx context.Context, roleID string) (int, error) {
	return 0, nil
}

func (f *fakeRoleRepository) AddMembers(ctx context.Context, roleID string, userIDs []string, appServiceID *string) error {
	f.addMembersCalls = append(f.addMembersCalls, addMembersCall{roleID: roleID, userIDs: userIDs, appServiceID: appServiceID})
	return nil
}

func (f *fakeRoleRepository) RemoveMember(ctx context.Context, roleID string, userID string, appServiceID *string) error {
	return nil
}

func (f *fakeRoleRepository) GetPermissionCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetRoleCodesByUserID(ctx context.Context, userID string, appServiceID string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleRepository) GetGlobalRoleCodesByUserIDs(ctx context.Context, userIDs []string) (map[string]string, error) {
	return nil, nil
}

// === invitation repository fake (in-memory) ===

type fakeInvitationRepository struct {
	invitations          map[string]*entity.UserInvitation
	nextID               int
	forceMarkAcceptedNop bool
}

func newFakeInvitationRepository() *fakeInvitationRepository {
	return &fakeInvitationRepository{invitations: map[string]*entity.UserInvitation{}}
}

func (f *fakeInvitationRepository) Create(ctx context.Context, invitation entity.UserInvitation) (string, error) {
	f.nextID++
	invitation.ID = fmt.Sprintf("inv-%d", f.nextID)
	invitation.Status = int32(constants.InvitationStatusPending)
	f.invitations[invitation.ID] = &invitation
	return invitation.ID, nil
}

func (f *fakeInvitationRepository) GetByID(ctx context.Context, id string) (entity.UserInvitation, error) {
	if invitation, ok := f.invitations[id]; ok {
		return *invitation, nil
	}
	return entity.UserInvitation{}, nil
}

func (f *fakeInvitationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (entity.UserInvitation, error) {
	for _, invitation := range f.invitations {
		if invitation.TokenHash == tokenHash {
			return *invitation, nil
		}
	}
	return entity.UserInvitation{}, nil
}

func (f *fakeInvitationRepository) GetPendingByEmail(ctx context.Context, email string) (entity.UserInvitation, error) {
	for _, invitation := range f.invitations {
		if invitation.Email == email && invitation.Status == int32(constants.InvitationStatusPending) {
			return *invitation, nil
		}
	}
	return entity.UserInvitation{}, nil
}

func (f *fakeInvitationRepository) List(ctx context.Context) ([]models.InvitationListItem, error) {
	items := []models.InvitationListItem{}
	for _, invitation := range f.invitations {
		items = append(items, models.InvitationListItem{
			ID:     invitation.ID,
			Email:  invitation.Email,
			RoleID: invitation.RoleID,
			Status: invitation.Status,
		})
	}
	return items, nil
}

func (f *fakeInvitationRepository) MarkAccepted(ctx context.Context, id string) (bool, error) {
	if f.forceMarkAcceptedNop {
		return false, nil
	}
	invitation, ok := f.invitations[id]
	if !ok || invitation.Status != int32(constants.InvitationStatusPending) {
		return false, nil
	}
	invitation.Status = int32(constants.InvitationStatusAccepted)
	invitation.AcceptedAt = time.Now().UTC()
	return true, nil
}

func (f *fakeInvitationRepository) MarkRevoked(ctx context.Context, id string) (bool, error) {
	invitation, ok := f.invitations[id]
	if !ok || invitation.Status != int32(constants.InvitationStatusPending) {
		return false, nil
	}
	invitation.Status = int32(constants.InvitationStatusRevoked)
	return true, nil
}

func (f *fakeInvitationRepository) RevertToPending(ctx context.Context, id string) error {
	invitation, ok := f.invitations[id]
	if !ok {
		return nil
	}
	invitation.Status = int32(constants.InvitationStatusPending)
	invitation.AcceptedAt = time.Time{}
	return nil
}

// === helpers ===

const acceptInvitePath = "/accept-invite"

func newTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Auth.EndpointWebAcceptInvite = acceptInvitePath
	return cfg
}

func newTestFixture() (*fakeInvitationRepository, *fakeUserRepository, *fakeRoleRepository, IUseCase) {
	invitationRepository := newFakeInvitationRepository()
	userRepository := &fakeUserRepository{}
	roleRepository := &fakeRoleRepository{
		roles: map[string]roleEntity.Role{
			"rol_member": {ID: "rol_member", Code: "member", Name: "Member"},
		},
	}
	invitationUsecase := NewUsecase(newTestConfig(), invitationRepository, userRepository, roleRepository)
	return invitationRepository, userRepository, roleRepository, invitationUsecase
}

func tokenFromLink(t *testing.T, link string) string {
	t.Helper()
	prefix := acceptInvitePath + "?token="
	if !strings.HasPrefix(link, prefix) {
		t.Fatalf("invite link %q does not start with %q", link, prefix)
	}
	return strings.TrimPrefix(link, prefix)
}

// === Create ===

func TestCreateInvitationHappyPath(t *testing.T) {
	invitationRepository, _, _, invitationUsecase := newTestFixture()

	before := time.Now().UTC()
	res, err := invitationUsecase.Create(context.Background(), models.CreateRequest{
		Email:  "linh.tran@hasaki.vn",
		RoleID: "rol_member",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got: %v", err)
	}
	if res.ID == "" {
		t.Fatal("expected invitation ID to be set")
	}

	rawToken := tokenFromLink(t, res.InviteLink)
	if rawToken == "" {
		t.Fatal("expected raw token in invite link")
	}

	stored := invitationRepository.invitations[res.ID]
	if stored == nil {
		t.Fatal("expected invitation to be stored")
	}
	if stored.TokenHash != cryp.HashSHA256(rawToken) {
		t.Error("stored token hash must equal SHA-256 of the raw token")
	}
	if stored.TokenHash == rawToken {
		t.Error("raw token must never be persisted")
	}
	wantExpiry := before.Add(constants.InvitationTTL)
	if stored.ExpiresAt.Before(wantExpiry.Add(-time.Minute)) || stored.ExpiresAt.After(wantExpiry.Add(time.Minute)) {
		t.Errorf("expires_at %v not within a minute of now+7d %v", stored.ExpiresAt, wantExpiry)
	}
	if stored.Status != int32(constants.InvitationStatusPending) {
		t.Errorf("expected pending status, got %d", stored.Status)
	}
}

func TestCreateInvitationRejections(t *testing.T) {
	t.Run("existing user email", func(t *testing.T) {
		_, userRepository, _, invitationUsecase := newTestFixture()
		userRepository.userByEmail = userEntity.User{ID: "user-1", Email: "taken@hasaki.vn"}

		_, err := invitationUsecase.Create(context.Background(), models.CreateRequest{Email: "taken@hasaki.vn", RoleID: "rol_member"})
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("expected existing-user rejection, got: %v", err)
		}
	})

	t.Run("pending invitation exists", func(t *testing.T) {
		_, _, _, invitationUsecase := newTestFixture()
		if _, err := invitationUsecase.Create(context.Background(), models.CreateRequest{Email: "dup@hasaki.vn", RoleID: "rol_member"}); err != nil {
			t.Fatalf("first create failed: %v", err)
		}

		_, err := invitationUsecase.Create(context.Background(), models.CreateRequest{Email: "dup@hasaki.vn", RoleID: "rol_member"})
		if err == nil || !strings.Contains(err.Error(), "pending invitation already exists") {
			t.Fatalf("expected pending-invite rejection, got: %v", err)
		}
	})

	t.Run("unknown role", func(t *testing.T) {
		_, _, _, invitationUsecase := newTestFixture()
		_, err := invitationUsecase.Create(context.Background(), models.CreateRequest{Email: "new@hasaki.vn", RoleID: "rol_ghost"})
		if err == nil || !strings.Contains(err.Error(), "role not found") {
			t.Fatalf("expected unknown-role rejection, got: %v", err)
		}
	})
}

// === Accept ===

func createInvitation(t *testing.T, invitationUsecase IUseCase, email string) (string, string) {
	t.Helper()
	res, err := invitationUsecase.Create(context.Background(), models.CreateRequest{Email: email, RoleID: "rol_member"})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	return res.ID, tokenFromLink(t, res.InviteLink)
}

func TestAcceptInvitationHappyPath(t *testing.T) {
	invitationRepository, userRepository, roleRepository, invitationUsecase := newTestFixture()
	invitationID, rawToken := createInvitation(t, invitationUsecase, "linh.tran@hasaki.vn")

	err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{
		Token:    rawToken,
		Name:     "Linh Tran",
		Password: "s3cret-pass",
	})
	if err != nil {
		t.Fatalf("expected accept to succeed, got: %v", err)
	}

	if len(userRepository.createCalls) != 1 {
		t.Fatalf("expected 1 user create call, got %d", len(userRepository.createCalls))
	}
	created := userRepository.createCalls[0]
	if created.email != "linh.tran@hasaki.vn" || created.name != "Linh Tran" {
		t.Errorf("unexpected user create call: %+v", created)
	}

	if len(userRepository.setPasswordCalls) != 1 {
		t.Fatalf("expected 1 SetPassword call, got %d", len(userRepository.setPasswordCalls))
	}
	if userRepository.setPasswordCalls[0] != (setPasswordCall{id: "user-1", password: "s3cret-pass"}) {
		t.Errorf("unexpected SetPassword call: %+v", userRepository.setPasswordCalls[0])
	}

	if len(userRepository.verifyCalls) != 1 || userRepository.verifyCalls[0] != "user-1" {
		t.Errorf("expected Verify(user-1), got %v", userRepository.verifyCalls)
	}

	if len(roleRepository.addMembersCalls) != 1 {
		t.Fatalf("expected 1 AddMembers call, got %d", len(roleRepository.addMembersCalls))
	}
	addMembers := roleRepository.addMembersCalls[0]
	if addMembers.roleID != "rol_member" || len(addMembers.userIDs) != 1 || addMembers.userIDs[0] != "user-1" || addMembers.appServiceID != nil {
		t.Errorf("unexpected AddMembers call: %+v", addMembers)
	}

	if invitationRepository.invitations[invitationID].Status != int32(constants.InvitationStatusAccepted) {
		t.Error("expected invitation to be marked accepted")
	}
}

func TestAcceptInvitationRejections(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		invitationRepository, userRepository, _, invitationUsecase := newTestFixture()
		invitationID, rawToken := createInvitation(t, invitationUsecase, "old@hasaki.vn")
		invitationRepository.invitations[invitationID].ExpiresAt = time.Now().UTC().Add(-time.Hour)

		err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Old", Password: "s3cret-pass"})
		if err == nil || !strings.Contains(err.Error(), "invalid or expired") {
			t.Fatalf("expected expired rejection, got: %v", err)
		}
		if len(userRepository.createCalls) != 0 {
			t.Error("expected no user creation for expired invite")
		}
	})

	t.Run("revoked", func(t *testing.T) {
		invitationRepository, _, _, invitationUsecase := newTestFixture()
		invitationID, rawToken := createInvitation(t, invitationUsecase, "rev@hasaki.vn")
		invitationRepository.invitations[invitationID].Status = int32(constants.InvitationStatusRevoked)

		err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Rev", Password: "s3cret-pass"})
		if err == nil || !strings.Contains(err.Error(), "invalid or expired") {
			t.Fatalf("expected revoked rejection, got: %v", err)
		}
	})

	t.Run("already accepted", func(t *testing.T) {
		_, _, _, invitationUsecase := newTestFixture()
		_, rawToken := createInvitation(t, invitationUsecase, "twice@hasaki.vn")
		if err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "First", Password: "s3cret-pass"}); err != nil {
			t.Fatalf("first accept failed: %v", err)
		}

		err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Second", Password: "s3cret-pass"})
		if err == nil || !strings.Contains(err.Error(), "invalid or expired") {
			t.Fatalf("expected already-accepted rejection, got: %v", err)
		}
	})

	t.Run("mark accepted claim lost", func(t *testing.T) {
		invitationRepository, userRepository, _, invitationUsecase := newTestFixture()
		_, rawToken := createInvitation(t, invitationUsecase, "race@hasaki.vn")
		invitationRepository.forceMarkAcceptedNop = true

		err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Race", Password: "s3cret-pass"})
		if err == nil || !strings.Contains(err.Error(), "already used") {
			t.Fatalf("expected already-used rejection, got: %v", err)
		}
		if len(userRepository.createCalls) != 0 {
			t.Error("user must NOT be created when the accept claim is lost")
		}
	})

	t.Run("email now exists", func(t *testing.T) {
		_, userRepository, _, invitationUsecase := newTestFixture()
		_, rawToken := createInvitation(t, invitationUsecase, "claimed@hasaki.vn")
		userRepository.userByEmail = userEntity.User{ID: "user-9", Email: "claimed@hasaki.vn"}

		err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Late", Password: "s3cret-pass"})
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("expected email-exists rejection, got: %v", err)
		}
	})

	t.Run("user create failure reverts the claim", func(t *testing.T) {
		invitationRepository, userRepository, _, invitationUsecase := newTestFixture()
		invitationID, rawToken := createInvitation(t, invitationUsecase, "fail@hasaki.vn")
		userRepository.createErr = errors.New("database unavailable")

		err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Fail", Password: "s3cret-pass"})
		if err == nil {
			t.Fatal("expected accept to fail when user creation fails")
		}
		if invitationRepository.invitations[invitationID].Status != int32(constants.InvitationStatusPending) {
			t.Error("expected invitation to be reverted to pending after user create failure")
		}
	})
}

// === Revoke ===

func TestRevokeInvitationTransitions(t *testing.T) {
	invitationRepository, _, _, invitationUsecase := newTestFixture()
	invitationID, _ := createInvitation(t, invitationUsecase, "rev@hasaki.vn")

	// pending → revoked
	if err := invitationUsecase.Revoke(context.Background(), invitationID); err != nil {
		t.Fatalf("expected revoke to succeed, got: %v", err)
	}
	if invitationRepository.invitations[invitationID].Status != int32(constants.InvitationStatusRevoked) {
		t.Error("expected invitation to be revoked")
	}

	// revoked → already revoked
	err := invitationUsecase.Revoke(context.Background(), invitationID)
	if err == nil || !strings.Contains(err.Error(), "already revoked") {
		t.Fatalf("expected already-revoked error, got: %v", err)
	}

	// accepted → already used
	acceptedID, rawToken := createInvitation(t, invitationUsecase, "used@hasaki.vn")
	if err := invitationUsecase.Accept(context.Background(), models.AcceptRequest{Token: rawToken, Name: "Used", Password: "s3cret-pass"}); err != nil {
		t.Fatalf("accept failed: %v", err)
	}
	err = invitationUsecase.Revoke(context.Background(), acceptedID)
	if err == nil || !strings.Contains(err.Error(), "already used") {
		t.Fatalf("expected already-used error, got: %v", err)
	}

	// unknown → not found
	err = invitationUsecase.Revoke(context.Background(), "inv-missing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
}

// === GetByToken ===

func TestGetInvitationByToken(t *testing.T) {
	t.Run("valid token resolves email and role name", func(t *testing.T) {
		_, _, _, invitationUsecase := newTestFixture()
		_, rawToken := createInvitation(t, invitationUsecase, "linh.tran@hasaki.vn")

		detail, err := invitationUsecase.GetByToken(context.Background(), rawToken)
		if err != nil {
			t.Fatalf("expected get by token to succeed, got: %v", err)
		}
		if detail.Email != "linh.tran@hasaki.vn" || detail.RoleName != "Member" {
			t.Errorf("unexpected detail: %+v", detail)
		}
	})

	t.Run("unknown token", func(t *testing.T) {
		_, _, _, invitationUsecase := newTestFixture()
		_, err := invitationUsecase.GetByToken(context.Background(), "no-such-token")
		if err == nil || !strings.Contains(err.Error(), "invalid or expired") {
			t.Fatalf("expected invalid-token error, got: %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		invitationRepository, _, _, invitationUsecase := newTestFixture()
		invitationID, rawToken := createInvitation(t, invitationUsecase, "old@hasaki.vn")
		invitationRepository.invitations[invitationID].ExpiresAt = time.Now().UTC().Add(-time.Hour)

		_, err := invitationUsecase.GetByToken(context.Background(), rawToken)
		if err == nil || !strings.Contains(err.Error(), "invalid or expired") {
			t.Fatalf("expected expired error, got: %v", err)
		}
	})

	t.Run("revoked token", func(t *testing.T) {
		_, _, _, invitationUsecase := newTestFixture()
		invitationID, rawToken := createInvitation(t, invitationUsecase, "rev@hasaki.vn")
		if err := invitationUsecase.Revoke(context.Background(), invitationID); err != nil {
			t.Fatalf("revoke failed: %v", err)
		}

		_, err := invitationUsecase.GetByToken(context.Background(), rawToken)
		if err == nil || !strings.Contains(err.Error(), "invalid or expired") {
			t.Fatalf("expected revoked error, got: %v", err)
		}
	})
}
