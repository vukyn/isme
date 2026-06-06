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
	Page      int    `json:"page" query:"page"`
	PageSize  int    `json:"pageSize" query:"pageSize"`
	Search    string `json:"search" query:"search"`       // search by name or email
	Status    int32  `json:"status" query:"status"`       // 1=active, 2=inactive
	RoleID    string `json:"roleID" query:"roleID"`       // filter by role
	IsAdmin   *bool  `json:"isAdmin" query:"isAdmin"`     // filter by admin status
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
	IsAdmin       bool   `json:"isAdmin"`
	Role          string `json:"role"`       // global role code
	SessionsCount int    `json:"sessionsCount"`
	LastLoginAt   string `json:"lastLoginAt"`
	CreatedAt     string `json:"createdAt"`
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
	ClientIP    string `json:"clientIP"`
	UserAgent   string `json:"userAgent"`
	LastLoginAt string `json:"lastLoginAt"`
	ExpiresAt   string `json:"expiresAt"`
	Status      int32  `json:"status"`
}
