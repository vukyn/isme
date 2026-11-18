package usecase

import (
	"context"

	"github.com/vukyn/isme/internal/domains/user_invitation/models"
)

type IUseCase interface {
	// Create a pending invitation and return the one-time invite link
	Create(ctx context.Context, req models.CreateRequest) (models.CreateResponse, error)
	// List all invitations
	List(ctx context.Context) (models.ListResponse, error)
	// Revoke a pending invitation
	Revoke(ctx context.Context, id string) error
	// Resolve a raw token to the invitation detail (public, pre-accept check)
	GetByToken(ctx context.Context, token string) (models.InviteDetailResponse, error)
	// Accept an invitation: create the user, set password, verify, assign role
	Accept(ctx context.Context, req models.AcceptRequest) error
}

// TxRunner runs fn inside ONE database transaction that every repository call
// made with fn's context joins (kuery/bun/tx). It carries no bun type, so the
// usecase layer states "these writes are atomic" without holding a handle.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
