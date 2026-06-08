package models

import (
	"errors"
	"regexp"

	pkgBase "github.com/vukyn/kuery/http/base"
)

var roleCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type CreateRequest struct {
	AppID           string `json:"app_id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CloneFromRoleID string `json:"clone_from_role_id"`
}

func (r CreateRequest) Validate() error {
	if r.AppID == "" {
		return errors.New("app_id is required")
	}
	if r.Code == "" {
		return errors.New("code is required")
	}
	if !roleCodePattern.MatchString(r.Code) {
		return errors.New("code must be a lowercase slug (a-z, 0-9, hyphen)")
	}
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// ListRequest filters the role catalog by owning app (empty = all apps).
type ListRequest struct {
	AppID   string `json:"app_id" query:"app_id"`
	AppCode string `json:"app_code" query:"app_code"`
}

func (r ListRequest) Validate() error {
	return nil
}

// ListPermissionsRequest filters the permission catalog by owning app (empty = all apps).
type ListPermissionsRequest struct {
	AppID   string `json:"app_id" query:"app_id"`
	AppCode string `json:"app_code" query:"app_code"`
}

func (r ListPermissionsRequest) Validate() error {
	return nil
}

type CreateResponse struct {
	ID string `json:"id"`
}

type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r UpdateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type SetPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}

func (r SetPermissionsRequest) Validate() error {
	for _, permissionID := range r.PermissionIDs {
		if permissionID <= 0 {
			return errors.New("permission_ids must be positive")
		}
	}
	return nil
}

type AddMembersRequest struct {
	UserIDs      []string `json:"user_ids"`
	AppServiceID *string  `json:"app_service_id"`
}

func (r AddMembersRequest) Validate() error {
	if len(r.UserIDs) == 0 {
		return errors.New("user_ids is required")
	}
	for _, userID := range r.UserIDs {
		if userID == "" {
			return errors.New("user_ids must not contain empty values")
		}
	}
	if r.AppServiceID != nil && *r.AppServiceID == "" {
		return errors.New("app_service_id must not be empty when set")
	}
	return nil
}

type ListMembersRequest struct {
	pkgBase.Pagination
	Query string `json:"query" query:"query"`
}

func (r ListMembersRequest) Validate() error {
	return nil
}

type RoleListItem struct {
	ID           string `json:"id"`
	AppID        string `json:"app_id"`
	AppCode      string `json:"app_code"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsSystem     bool   `json:"is_system"`
	MembersCount int    `json:"members_count"`
}

type PermissionItem struct {
	ID       int64  `json:"id"`
	AppID    string `json:"app_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type RoleDetailResponse struct {
	ID          string           `json:"id"`
	AppID       string           `json:"app_id"`
	AppCode     string           `json:"app_code"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	IsSystem    bool             `json:"is_system"`
	Permissions []PermissionItem `json:"permissions"`
}

// UserAppRole is one app-scoped role a user holds, used by the user list to
// render app:role chips. It carries both codes and display names.
type UserAppRole struct {
	AppCode  string `json:"app_code"`
	AppName  string `json:"app_name"`
	RoleCode string `json:"role_code"`
	RoleName string `json:"role_name"`
}

type MemberItem struct {
	UserID       string  `json:"user_id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	AppServiceID *string `json:"app_service_id"`
	CreatedAt    string  `json:"created_at"`
}

type ListMembersResponse struct {
	Items []MemberItem `json:"items"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
}
