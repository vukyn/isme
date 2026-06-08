package entity

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// UserInvitationRole is one app-scoped role assignment carried by an invitation.
// An invitation may hold several of these (e.g. medioa2->Editor + rainy->Viewer);
// on accept each becomes an app-scoped user_roles row.
type UserInvitationRole struct {
	bun.BaseModel `bun:"table:user_invitation_roles,alias:uir"`
	ID            string    `bun:"id,pk,notnull"`
	InvitationID  string    `bun:"invitation_id,notnull"`
	RoleID        string    `bun:"role_id,notnull"`
	AppServiceID  string    `bun:"app_service_id,notnull"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
	CreatedBy     string    `bun:"created_by,nullzero"`
	UpdatedAt     time.Time `bun:"updated_at,default:current_timestamp"`
	UpdatedBy     string    `bun:"updated_by,nullzero"`
	DeletedAt     time.Time `bun:"deleted_at,soft_delete,nullzero"`
	DeletedBy     string    `bun:"deleted_by,nullzero"`
}

// === Hooks ===

func (u *UserInvitationRole) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch q := query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now().UTC()
	case *bun.UpdateQuery:
		q.Column("updated_at")
		u.UpdatedAt = time.Now().UTC()
	}
	return nil
}
