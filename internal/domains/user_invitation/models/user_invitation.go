package models

import (
	"errors"

	"github.com/vukyn/kuery/validator"
)

type CreateRequest struct {
	Email  string `json:"email"`
	RoleID string `json:"role_id"`
}

func (r CreateRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if !validator.IsEmail(r.Email) {
		return errors.New("invalid email")
	}
	if r.RoleID == "" {
		return errors.New("role_id is required")
	}
	return nil
}

type CreateResponse struct {
	ID string `json:"id"`
	// One-time link — the raw token is never persisted, only its hash
	InviteLink string `json:"invite_link"`
}

type InvitationListItem struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	RoleID     string `json:"role_id"`
	RoleName   string `json:"role_name"`
	Status     int32  `json:"status"`
	ExpiresAt  string `json:"expires_at"`
	AcceptedAt string `json:"accepted_at"`
	CreatedAt  string `json:"created_at"`
}

type ListResponse struct {
	Items []InvitationListItem `json:"items"`
}

type InviteDetailResponse struct {
	Email    string `json:"email"`
	RoleName string `json:"role_name"`
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
