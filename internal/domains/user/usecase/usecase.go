package usecase

import (
	"context"
	"time"

	roleRepo "github.com/vukyn/isme/internal/domains/role/repository"
	"github.com/vukyn/isme/internal/domains/user/models"
	userRepo "github.com/vukyn/isme/internal/domains/user/repository"
	userSessionRepo "github.com/vukyn/isme/internal/domains/user_session/repository"

	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
)

type usecase struct {
	userRepo        userRepo.IRepository
	userSessionRepo userSessionRepo.IRepository
	roleRepo        roleRepo.IRepository
}

func NewUsecase(
	userRepo userRepo.IRepository,
	userSessionRepo userSessionRepo.IRepository,
	roleRepo roleRepo.IRepository,
) IUseCase {
	return &usecase{
		userRepo:        userRepo,
		userSessionRepo: userSessionRepo,
		roleRepo:        roleRepo,
	}
}

func (u *usecase) List(ctx context.Context, req models.ListRequest) (models.ListResponse, error) {
	// validation
	if err := req.Validate(); err != nil {
		return models.ListResponse{}, pkgErr.InvalidRequest(err.Error())
	}

	users, total, err := u.userRepo.List(ctx, req)
	if err != nil {
		return models.ListResponse{}, err
	}

	// enrich the page with session counts and each user's app-scoped roles
	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}
	sessionCounts, err := u.userSessionRepo.CountActiveByUserIDs(ctx, userIDs)
	if err != nil {
		return models.ListResponse{}, err
	}
	rolesByUser, err := u.roleRepo.GetRoleCodesGroupedByAppByUserIDs(ctx, userIDs)
	if err != nil {
		return models.ListResponse{}, err
	}

	items := make([]models.UserListItem, 0, len(users))
	for _, user := range users {
		lastLoginAt := ""
		if !user.LastLoginAt.IsZero() {
			lastLoginAt = user.LastLoginAt.Format(time.RFC3339)
		}
		roles := make([]models.AppRole, 0, len(rolesByUser[user.ID]))
		for _, role := range rolesByUser[user.ID] {
			roles = append(roles, models.AppRole{
				AppCode:  role.AppCode,
				AppName:  role.AppName,
				RoleCode: role.RoleCode,
				RoleName: role.RoleName,
			})
		}
		items = append(items, models.UserListItem{
			ID:            user.ID,
			Name:          user.Name,
			Email:         user.Email,
			Status:        user.Status,
			IsVerified:    user.IsVerified,
			Roles:         roles,
			SessionsCount: sessionCounts[user.ID],
			LastLoginAt:   lastLoginAt,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		})
	}

	return models.ListResponse{
		Items: items,
		Total: total,
		Page:  req.Page,
	}, nil
}

func (u *usecase) UpdateStatus(ctx context.Context, id string, req models.UpdateStatusRequest) error {
	// validation
	if err := req.Validate(); err != nil {
		return pkgErr.InvalidRequest(err.Error())
	}

	// check user exists
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.ID == "" {
		return pkgErr.NotFound("user not found")
	}

	return u.userRepo.UpdateStatus(ctx, id, req.Status)
}

func (u *usecase) VerifyUser(ctx context.Context, id string) error {
	// check user exists
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.ID == "" {
		return pkgErr.NotFound("user not found")
	}

	// verification is one-way — re-verifying is rejected explicitly
	if user.IsVerified {
		return pkgErr.InvalidRequest("user already verified")
	}

	return u.userRepo.Verify(ctx, id)
}

func (u *usecase) SoftDelete(ctx context.Context, id string) error {
	// reject self-delete
	if pkgCtx.GetUserID(ctx) == id {
		return pkgErr.InvalidRequest("cannot delete your own account")
	}

	// check user exists
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.ID == "" {
		return pkgErr.NotFound("user not found")
	}

	// delete user and revoke all sessions
	if err := u.userRepo.SoftDelete(ctx, id); err != nil {
		return err
	}
	return u.userSessionRepo.InactiveAllUserSession(ctx, id)
}

func (u *usecase) ListSessions(ctx context.Context, userID string) ([]models.SessionItem, error) {
	// check user exists
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.ID == "" {
		return nil, pkgErr.NotFound("user not found")
	}

	sessions, err := u.userSessionRepo.GetListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]models.SessionItem, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, models.SessionItem{
			ID:          session.ID,
			ClientIP:    session.ClientIP,
			UserAgent:   session.UserAgent,
			LastLoginAt: session.LastLoginAt.Format(time.RFC3339),
			ExpiresAt:   session.ExpiresAt.Format(time.RFC3339),
			Status:      session.Status,
		})
	}
	return items, nil
}

func (u *usecase) RevokeSession(ctx context.Context, userID string, sessionID string) error {
	// check session exists and belongs to the user
	session, err := u.userSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.ID == "" || session.UserID != userID {
		return pkgErr.NotFound("session not found")
	}

	return u.userSessionRepo.InactiveSessionByID(ctx, sessionID)
}
