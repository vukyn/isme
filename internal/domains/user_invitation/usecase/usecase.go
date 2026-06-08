package usecase

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/vukyn/isme/internal/config"
	appServiceRepo "github.com/vukyn/isme/internal/domains/app_service/repository"
	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	userModels "github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	"github.com/vukyn/isme/internal/domains/user_invitation/constants"
	"github.com/vukyn/isme/internal/domains/user_invitation/entity"
	"github.com/vukyn/isme/internal/domains/user_invitation/models"
	invitationRepo "github.com/vukyn/isme/internal/domains/user_invitation/repository"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"

	"github.com/vukyn/kuery/cryp"
	"github.com/vukyn/kuery/cryp/rand"
)

type usecase struct {
	cfg            *config.Config
	invitationRepo invitationRepo.IRepository
	userRepo       userRepo.IRepository
	roleRepo       roleRepo.IRepository
	appServiceRepo appServiceRepo.IRepository
}

func NewUsecase(
	cfg *config.Config,
	invitationRepo invitationRepo.IRepository,
	userRepo userRepo.IRepository,
	roleRepo roleRepo.IRepository,
	appServiceRepo appServiceRepo.IRepository,
) IUseCase {
	return &usecase{
		cfg:            cfg,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		appServiceRepo: appServiceRepo,
	}
}

func (u *usecase) Create(ctx context.Context, req models.CreateRequest) (models.CreateResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.CreateResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// validate every assignment: the role must exist and its owning app must
	// match the requested app_service_id (reject cross-app mismatches).
	assignments := make([]entity.UserInvitationRole, 0, len(req.Assignments))
	for _, assignment := range req.Assignments {
		role, err := u.roleRepo.GetByID(ctx, assignment.RoleID)
		if err != nil {
			return models.CreateResponse{}, err
		}
		if role.ID == "" {
			return models.CreateResponse{}, pkgErr.InvalidRequest("role not found")
		}
		if role.AppID != assignment.AppServiceID {
			return models.CreateResponse{}, pkgErr.InvalidRequest("role does not belong to the given app_service_id")
		}
		assignments = append(assignments, entity.UserInvitationRole{
			RoleID:       assignment.RoleID,
			AppServiceID: assignment.AppServiceID,
		})
	}

	// check if a user already holds this email
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return models.CreateResponse{}, err
	}
	if user.ID != "" {
		return models.CreateResponse{}, pkgErr.InvalidRequest("user with this email already exists")
	}

	// check if a pending invitation already exists
	pending, err := u.invitationRepo.GetPendingByEmail(ctx, req.Email)
	if err != nil {
		return models.CreateResponse{}, err
	}
	if pending.ID != "" {
		return models.CreateResponse{}, pkgErr.InvalidRequest("a pending invitation already exists for this email")
	}

	// generate the one-time token; only its hash is persisted
	rawToken := base64.RawURLEncoding.EncodeToString([]byte(rand.RandString(32)))
	invitationID, err := u.invitationRepo.Create(ctx, entity.UserInvitation{
		Email:     req.Email,
		TokenHash: cryp.HashSHA256(rawToken),
		ExpiresAt: time.Now().UTC().Add(constants.InvitationTTL),
		CreatedBy: pkgCtx.GetUserID(ctx),
	}, assignments)
	if err != nil {
		return models.CreateResponse{}, err
	}

	return models.CreateResponse{
		ID:         invitationID,
		InviteLink: u.cfg.Auth.EndpointWebAcceptInvite + "?token=" + rawToken,
	}, nil
}

func (u *usecase) List(ctx context.Context) (models.ListResponse, error) {
	items, err := u.invitationRepo.List(ctx)
	if err != nil {
		return models.ListResponse{}, err
	}
	return models.ListResponse{Items: items}, nil
}

func (u *usecase) Revoke(ctx context.Context, id string) error {
	invitation, err := u.invitationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if invitation.ID == "" {
		return pkgErr.NotFound("invitation not found")
	}
	if invitation.Status == int32(constants.InvitationStatusAccepted) {
		return pkgErr.InvalidRequest("invitation already used")
	}
	if invitation.Status == int32(constants.InvitationStatusRevoked) {
		return pkgErr.InvalidRequest("invitation already revoked")
	}

	// conditional update guards against a concurrent accept
	revoked, err := u.invitationRepo.MarkRevoked(ctx, id)
	if err != nil {
		return err
	}
	if !revoked {
		return pkgErr.InvalidRequest("invitation already used")
	}
	return nil
}

func (u *usecase) GetByToken(ctx context.Context, token string) (models.InviteDetailResponse, error) {
	// Public, pre-auth endpoint. We resolve by token hash only — an unknown
	// token is the single "invalid" case. For a known token we always return
	// the locked email, a display status (so the page can show valid / expired
	// / accepted / revoked), and a permission preview of the invited roles.
	if token == "" {
		return models.InviteDetailResponse{}, pkgErr.NotFound("invitation is invalid or expired")
	}

	invitation, err := u.invitationRepo.GetByTokenHash(ctx, cryp.HashSHA256(token))
	if err != nil {
		return models.InviteDetailResponse{}, err
	}
	if invitation.ID == "" {
		return models.InviteDetailResponse{}, pkgErr.NotFound("invitation is invalid or expired")
	}

	displayStatus := deriveDisplayStatus(invitation)

	assignments, err := u.buildAssignmentDetails(ctx, invitation.ID)
	if err != nil {
		return models.InviteDetailResponse{}, err
	}

	return models.InviteDetailResponse{
		Email:         invitation.Email,
		Status:        invitation.Status,
		DisplayStatus: displayStatus,
		Assignments:   assignments,
	}, nil
}

// deriveDisplayStatus maps the stored status (+ expiry) to a stable string the
// AcceptInvite page switches on. Pending-but-past-expiry collapses to expired.
func deriveDisplayStatus(invitation entity.UserInvitation) string {
	switch invitation.Status {
	case int32(constants.InvitationStatusAccepted):
		return constants.DisplayStatusAccepted
	case int32(constants.InvitationStatusRevoked):
		return constants.DisplayStatusRevoked
	default:
		if invitation.ExpiresAt.Before(time.Now().UTC()) {
			return constants.DisplayStatusExpired
		}
		return constants.DisplayStatusValid
	}
}

// buildAssignmentDetails loads an invitation's assignments joined to their app
// + role, and previews the permission codes each role grants. Exposes ONLY the
// invited roles' perms + app names — never the full catalog (pre-auth safe).
func (u *usecase) buildAssignmentDetails(ctx context.Context, invitationID string) ([]models.InviteAssignmentDetail, error) {
	assignments, err := u.invitationRepo.GetAssignmentsByInvitationID(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if len(assignments) == 0 {
		return []models.InviteAssignmentDetail{}, nil
	}

	roleIDs := make([]string, 0, len(assignments))
	appServiceIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		roleIDs = append(roleIDs, assignment.RoleID)
		appServiceIDs = append(appServiceIDs, assignment.AppServiceID)
	}

	permsByRole, err := u.roleRepo.GetPermissionCodesByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	appsByID, err := u.appServiceRepo.GetByIDs(ctx, appServiceIDs)
	if err != nil {
		return nil, err
	}

	details := make([]models.InviteAssignmentDetail, 0, len(assignments))
	for _, assignment := range assignments {
		role, err := u.roleRepo.GetByID(ctx, assignment.RoleID)
		if err != nil {
			return nil, err
		}
		perms := permsByRole[assignment.RoleID]
		if perms == nil {
			perms = []string{}
		}
		app := appsByID[assignment.AppServiceID]
		details = append(details, models.InviteAssignmentDetail{
			AppCode:     app.AppCode,
			AppName:     app.AppName,
			RoleName:    role.Name,
			RoleCode:    role.Code,
			Permissions: perms,
		})
	}
	return details, nil
}

func (u *usecase) Accept(ctx context.Context, req models.AcceptRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	invitation, err := u.resolveToken(ctx, req.Token)
	if err != nil {
		return err
	}

	// re-check the email is still free (a user may have been created since the invite)
	user, err := u.userRepo.GetByEmail(ctx, invitation.Email)
	if err != nil {
		return err
	}
	if user.ID != "" {
		return pkgErr.InvalidRequest("user with this email already exists")
	}

	// load the invitation's app-scoped role assignments
	assignments, err := u.invitationRepo.GetAssignmentsByInvitationID(ctx, invitation.ID)
	if err != nil {
		return err
	}
	if len(assignments) == 0 {
		return pkgErr.InvalidRequest("invitation has no role assignments")
	}

	// claim the invitation atomically — a lost race means it was already used
	accepted, err := u.invitationRepo.MarkAccepted(ctx, invitation.ID)
	if err != nil {
		return err
	}
	if !accepted {
		return pkgErr.InvalidRequest("invitation already used")
	}

	// create the user; roll the claim back on failure so the link stays usable.
	// Roles are assigned per-assignment below, not by user create.
	userID, err := u.userRepo.Create(ctx, userModels.CreateRequest{
		Name:  req.Name,
		Email: invitation.Email,
	})
	if err != nil {
		_ = u.invitationRepo.RevertToPending(ctx, invitation.ID)
		return err
	}

	// set user password
	if err := u.userRepo.SetPassword(ctx, userID, req.Password); err != nil {
		return err
	}

	// the inviting admin vouched for this account — no separate verify step
	if err := u.userRepo.Verify(ctx, userID); err != nil {
		return err
	}

	// grant one app-scoped role per assignment (ur.app_service_id = roles.app_id)
	for _, assignment := range assignments {
		appServiceID := assignment.AppServiceID
		if err := u.roleRepo.AddMembers(ctx, assignment.RoleID, []string{userID}, &appServiceID); err != nil {
			return err
		}
	}

	return nil
}

// resolveToken maps a raw token to its live pending invitation. Every failure
// mode returns the same generic error so callers can't probe token state.
func (u *usecase) resolveToken(ctx context.Context, token string) (entity.UserInvitation, error) {
	if token == "" {
		return entity.UserInvitation{}, pkgErr.NotFound("invitation is invalid or expired")
	}

	invitation, err := u.invitationRepo.GetByTokenHash(ctx, cryp.HashSHA256(token))
	if err != nil {
		return entity.UserInvitation{}, err
	}
	if invitation.ID == "" {
		return entity.UserInvitation{}, pkgErr.NotFound("invitation is invalid or expired")
	}
	if invitation.Status != int32(constants.InvitationStatusPending) {
		return entity.UserInvitation{}, pkgErr.NotFound("invitation is invalid or expired")
	}
	if invitation.ExpiresAt.Before(time.Now().UTC()) {
		return entity.UserInvitation{}, pkgErr.NotFound("invitation is invalid or expired")
	}
	return invitation, nil
}
