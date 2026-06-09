package models

import (
	"errors"

	"github.com/vukyn/kuery/validator"
)

type CreateRequest struct {
	Name  string
	Email string
	// RoleID + AppServiceID assign an app-scoped role to the new user at creation.
	// Both must be set together (or both empty for no initial role assignment).
	RoleID       string
	AppServiceID string
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
	if (r.RoleID == "") != (r.AppServiceID == "") {
		return errors.New("role_id and app_service_id must be set together")
	}
	return nil
}

// ListRequest for listing users with pagination and filters
type ListRequest struct {
	Page     int    `json:"page" query:"page"`
	PageSize int    `json:"size" query:"size"`
	Search   string `json:"query" query:"query"`   // search by name or email
	Status   int32  `json:"status" query:"status"` // 1=active, 2=inactive
	// AppCode scopes the role filter to one app_service (app-owned RBAC).
	AppCode string `json:"app" query:"app"`
	// RoleID carries a role *code* (not an id); app-scoped, only honoured when AppCode is set.
	RoleID   string `json:"role" query:"role"`         // filter by role code (requires AppCode)
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

// AppRole is one app-scoped role a user holds, rendered as an app:role chip.
type AppRole struct {
	AppCode  string `json:"app_code"`
	AppName  string `json:"app_name"`
	RoleCode string `json:"role_code"`
	RoleName string `json:"role_name"`
}

// UserListItem represents a user in the list response
type UserListItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Status        int32     `json:"status"`
	IsVerified    bool      `json:"is_verified"`
	Roles         []AppRole `json:"roles"` // full set of app-scoped roles
	SessionsCount int       `json:"sessions_count"`
	LastLoginAt   string    `json:"last_login_at"`
	CreatedAt     string    `json:"created_at"`
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
