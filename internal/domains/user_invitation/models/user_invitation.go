package models

import (
	"errors"

	"github.com/vukyn/kuery/validator"
)

// RoleAssignment pairs a role with the app it is scoped to. An invitation
// carries one or more of these; on accept each becomes an app-scoped user role.
type RoleAssignment struct {
	RoleID       string `json:"role_id"`
	AppServiceID string `json:"app_service_id"`
}

type CreateRequest struct {
	Email string `json:"email"`
	// Assignments are the app-scoped roles the invitee receives on accept.
	Assignments []RoleAssignment `json:"assignments"`
}

func (r CreateRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if !validator.IsEmail(r.Email) {
		return errors.New("invalid email")
	}
	if len(r.Assignments) == 0 {
		return errors.New("at least one role assignment is required")
	}
	for _, assignment := range r.Assignments {
		if assignment.RoleID == "" {
			return errors.New("role_id is required for each assignment")
		}
		if assignment.AppServiceID == "" {
			return errors.New("app_service_id is required for each assignment")
		}
	}
	return nil
}

type CreateResponse struct {
	ID string `json:"id"`
	// One-time link — the raw token is never persisted, only its hash
	InviteLink string `json:"invite_link"`
}

// AssignmentItem describes one assigned role on a listed invitation.
type AssignmentItem struct {
	RoleID       string `json:"role_id"`
	RoleName     string `json:"role_name"`
	RoleCode     string `json:"role_code"`
	AppServiceID string `json:"app_service_id"`
	AppCode      string `json:"app_code"`
	AppName      string `json:"app_name"`
}

type InvitationListItem struct {
	ID          string           `json:"id"`
	Email       string           `json:"email"`
	Status      int32            `json:"status"`
	Assignments []AssignmentItem `json:"assignments"`
	ExpiresAt   string           `json:"expires_at"`
	AcceptedAt  string           `json:"accepted_at"`
	CreatedAt   string           `json:"created_at"`
}

type ListResponse struct {
	Items []InvitationListItem `json:"items"`
}

// InviteAssignmentDetail is the public, pre-auth view of one assigned role.
// It exposes only the invited role's app, role identity, and the permission
// preview (resource:action codes) the role grants — nothing else.
type InviteAssignmentDetail struct {
	AppCode  string `json:"app_code"`
	AppName  string `json:"app_name"`
	RoleName string `json:"role_name"`
	RoleCode string `json:"role_code"`
	// Permissions previews the resource:action codes this role grants.
	Permissions []string `json:"permissions"`
}

// InviteDetailResponse is the public payload the AcceptInvite page renders.
// Status + DisplayStatus let the page distinguish valid / expired / used /
// revoked without leaking anything beyond the invited roles.
type InviteDetailResponse struct {
	Email string `json:"email"`
	// Status is the raw invitation status (1=pending, 2=accepted, 3=revoked).
	Status int32 `json:"status"`
	// DisplayStatus is a stable string the UI can switch on:
	// "valid" | "expired" | "accepted" | "revoked".
	DisplayStatus string                   `json:"display_status"`
	Assignments   []InviteAssignmentDetail `json:"assignments"`
}

type AcceptRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (r AcceptRequest) Validate() error {
	if r.Token == "" {
		return errors.New("token is required")
	}
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}
