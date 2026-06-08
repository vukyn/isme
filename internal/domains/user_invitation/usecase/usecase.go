package usecase

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/vukyn/isme/internal/config"
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
}

func NewUsecase(
	cfg *config.Config,
	invitationRepo invitationRepo.IRepository,
	userRepo userRepo.IRepository,
	roleRepo roleRepo.IRepository,
) IUseCase {
	return &usecase{
		cfg:            cfg,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
	}
}

func (u *usecase) Create(ctx context.Context, req models.CreateRequest) (models.CreateResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.CreateResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	// check if role exists
	role, err := u.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return models.CreateResponse{}, err
	}
	if role.ID == "" {
		return models.CreateResponse{}, pkgErr.InvalidRequest("role not found")
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
		RoleID:    req.RoleID,
		TokenHash: cryp.HashSHA256(rawToken),
		ExpiresAt: time.Now().UTC().Add(constants.InvitationTTL),
		CreatedBy: pkgCtx.GetUserID(ctx),
	})
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
	invitation, err := u.resolveToken(ctx, token)
	if err != nil {
		return models.InviteDetailResponse{}, err
	}

	role, err := u.roleRepo.GetByID(ctx, invitation.RoleID)
	if err != nil {
		return models.InviteDetailResponse{}, err
	}

	return models.InviteDetailResponse{
		Email:    invitation.Email,
		RoleName: role.Name,
	}, nil
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

	// claim the invitation atomically — a lost race means it was already used
	accepted, err := u.invitationRepo.MarkAccepted(ctx, invitation.ID)
	if err != nil {
		return err
	}
	if !accepted {
		return pkgErr.InvalidRequest("invitation already used")
	}

	// create the user; roll the claim back on failure so the link stays usable
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

	// assign the invited role globally
	if err := u.roleRepo.AddMembers(ctx, invitation.RoleID, []string{userID}, nil); err != nil {
		return err
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
