package models

import (
	"errors"
	"regexp"

	pkgBase "github.com/vukyn/kuery/http/base"
)

var roleCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type CreateRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CloneFromRoleID string `json:"clone_from_role_id"`
}

func (r CreateRequest) Validate() error {
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
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsSystem     bool   `json:"is_system"`
	MembersCount int    `json:"members_count"`
}

type PermissionItem struct {
	ID       int64  `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type RoleDetailResponse struct {
	ID          string           `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	IsSystem    bool             `json:"is_system"`
	Permissions []PermissionItem `json:"permissions"`
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
