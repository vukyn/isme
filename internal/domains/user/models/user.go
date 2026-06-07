package models

import (
	"errors"

	"github.com/vukyn/kuery/validator"
)

type CreateRequest struct {
	Name  string
	Email string
}

func (r CreateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if !validator.IsEmail(r.Email) {
		return errors.New("invalid email")
	}
	return nil
}

// ListRequest for listing users with pagination and filters
type ListRequest struct {
	Page     int    `json:"page" query:"page"`
	PageSize int    `json:"size" query:"size"`
	Search   string `json:"query" query:"query"`       // search by name or email
	Status   int32  `json:"status" query:"status"`     // 1=active, 2=inactive
	RoleID   string `json:"role" query:"role"`         // filter by role
	IsAdmin  *bool  `json:"is_admin" query:"is_admin"` // filter by admin status
	Verified *bool  `json:"verified" query:"verified"` // filter by verification state (nil=all)
}

func (r ListRequest) Validate() error {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 10
	}
	return nil
}

// UserListItem represents a user in the list response
type UserListItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Status        int32  `json:"status"`
	IsAdmin       bool   `json:"is_admin"`
	IsVerified    bool   `json:"is_verified"`
	Role          string `json:"role"` // global role code
	SessionsCount int    `json:"sessions_count"`
	LastLoginAt   string `json:"last_login_at"`
	CreatedAt     string `json:"created_at"`
}

// ListResponse for user list endpoint
type ListResponse struct {
	Items []UserListItem `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
}

// UpdateStatusRequest for updating user status
type UpdateStatusRequest struct {
	Status int32 `json:"status"`
}

func (r UpdateStatusRequest) Validate() error {
	if r.Status != 1 && r.Status != 2 {
		return errors.New("invalid status, must be 1 (active) or 2 (inactive)")
	}
	return nil
}

// SessionItem represents an active user session
type SessionItem struct {
	ID          string `json:"id"`
	ClientIP    string `json:"client_ip"`
	UserAgent   string `json:"user_agent"`
	LastLoginAt string `json:"last_login_at"`
	ExpiresAt   string `json:"expires_at"`
	Status      int32  `json:"status"`
}
