package models

import (
	"errors"
	"regexp"

	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"

	pkgBase "github.com/vukyn/kuery/http/base"
)

var roleCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// permissionTokenPattern validates a permission resource or action segment.
// Lowercase letters, digits and underscores only — the ":" separator is
// reserved for joining resource and action into a claim code.
var permissionTokenPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*$`)

type CreateRequest struct {
	AppID           string `json:"app_id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CloneFromRoleID string `json:"clone_from_role_id"`
	// Icon is an optional icon key (allowlist in roleConstants); empty = neutral
	// default in the UI.
	Icon string `json:"icon"`
	// Color is an optional color palette key (allowlist in roleConstants);
	// empty = neutral fallback in the UI.
	Color string `json:"color"`
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
	if !roleConstants.IsValidIcon(r.Icon) {
		return errors.New("icon is not a known icon key")
	}
	if !roleConstants.IsValidColor(r.Color) {
		return errors.New("color is not a known color key")
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

// PermissionPair is one resource:action permission to create for an app. Icon
// is an optional per-resource icon key (allowlist in roleConstants); empty =
// neutral default. When the resource already exists in the app, the repository
// reuses that resource's existing icon and ignores this value.
type PermissionPair struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Icon     string `json:"icon"`
	// Color is an optional per-resource color palette key (allowlist in
	// roleConstants); empty = neutral fallback. When the resource already
	// exists in the app, the repository reuses that resource's existing color
	// and ignores this value (mirrors Icon).
	Color string `json:"color"`
}

// CreatePermissionsRequest creates one or many resource:action permissions for
// an app's catalog. Creation is idempotent — existing pairs are returned with
// their ids.
type CreatePermissionsRequest struct {
	AppID       string           `json:"app_id"`
	Permissions []PermissionPair `json:"permissions"`
}

func (r CreatePermissionsRequest) Validate() error {
	if r.AppID == "" {
		return errors.New("app_id is required")
	}
	if len(r.Permissions) == 0 {
		return errors.New("permissions is required")
	}
	for _, permission := range r.Permissions {
		if permission.Resource == "" {
			return errors.New("resource is required")
		}
		if permission.Action == "" {
			return errors.New("action is required")
		}
		if !permissionTokenPattern.MatchString(permission.Resource) {
			return errors.New("resource must be lowercase (a-z, 0-9, underscore) with no ':'")
		}
		if !permissionTokenPattern.MatchString(permission.Action) {
			return errors.New("action must be lowercase (a-z, 0-9, underscore) with no ':'")
		}
		if !roleConstants.IsValidPermissionIcon(permission.Icon) {
			return errors.New("icon is not a known icon key")
		}
		if !roleConstants.IsValidColor(permission.Color) {
			return errors.New("color is not a known color key")
		}
	}
	return nil
}

type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Icon is the role's icon key (allowlist in roleConstants); empty = neutral
	// default in the UI.
	Icon string `json:"icon"`
	// Color is the role's color palette key (allowlist in roleConstants);
	// empty = neutral fallback in the UI.
	Color string `json:"color"`
}

func (r UpdateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if !roleConstants.IsValidIcon(r.Icon) {
		return errors.New("icon is not a known icon key")
	}
	if !roleConstants.IsValidColor(r.Color) {
		return errors.New("color is not a known color key")
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
	Icon         string `json:"icon"`
	Color        string `json:"color"`
	IsSystem     bool   `json:"is_system"`
	MembersCount int    `json:"members_count"`
}

type PermissionItem struct {
	ID       int64  `json:"id"`
	AppID    string `json:"app_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	// Icon is the per-resource icon key shared by all rows of the same
	// (app_id, resource); empty = neutral default in the UI.
	Icon string `json:"icon"`
	// Color is the per-resource color palette key shared by all rows of the
	// same (app_id, resource); empty = neutral fallback in the UI.
	Color string `json:"color"`
}

type RoleDetailResponse struct {
	ID          string           `json:"id"`
	AppID       string           `json:"app_id"`
	AppCode     string           `json:"app_code"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Icon        string           `json:"icon"`
	Color       string           `json:"color"`
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
