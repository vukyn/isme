package constants

import (
	"github.com/vukyn/kuery/rbac"
)

// Permission codes (must match the catalog seeded by 010_seed_rbac + 012_add_user_verification — 19 permissions)
var (
	PERM_USER_READ           = rbac.Perm("user", "read")
	PERM_USER_CREATE         = rbac.Perm("user", "create")
	PERM_USER_UPDATE         = rbac.Perm("user", "update")
	PERM_USER_DELETE         = rbac.Perm("user", "delete")
	PERM_USER_RESET_PASSWORD = rbac.Perm("user", "reset_password")
	PERM_USER_VERIFY         = rbac.Perm("user", "verify")

	PERM_USER_SESSION_READ   = rbac.Perm("user_session", "read")
	PERM_USER_SESSION_DELETE = rbac.Perm("user_session", "delete")
	PERM_USER_SESSION_REVOKE = rbac.Perm("user_session", "revoke")

	PERM_APP_SERVICE_READ          = rbac.Perm("app_service", "read")
	PERM_APP_SERVICE_CREATE        = rbac.Perm("app_service", "create")
	PERM_APP_SERVICE_UPDATE        = rbac.Perm("app_service", "update")
	PERM_APP_SERVICE_DELETE        = rbac.Perm("app_service", "delete")
	PERM_APP_SERVICE_ROTATE_SECRET = rbac.Perm("app_service", "rotate_secret")

	PERM_ROLE_READ   = rbac.Perm("role", "read")
	PERM_ROLE_CREATE = rbac.Perm("role", "create")
	PERM_ROLE_UPDATE = rbac.Perm("role", "update")
	PERM_ROLE_DELETE = rbac.Perm("role", "delete")
	PERM_ROLE_ASSIGN = rbac.Perm("role", "assign")
)

// isme self-app identifiers. isme is itself an app_service that owns the
// original permission catalog and system roles (seeded by 014/015).
const (
	APP_ID_ISME   = "app_isme"
	APP_CODE_ISME = "isme"
)

// System role codes
const (
	ROLE_CODE_ADMIN  = "admin"
	ROLE_CODE_MEMBER = "member"
	ROLE_CODE_VIEWER = "viewer"
)

// System role IDs (deterministic, owned by app_isme — seeded by 010_seed_rbac)
const (
	ROLE_ID_ADMIN  = "rol_admin"
	ROLE_ID_MEMBER = "rol_member"
	ROLE_ID_VIEWER = "rol_viewer"
)

// ICON_KEYS is the shared allowlist of icon keys a caller may store on any
// entity that carries an icon (permission resource, app_service tile). Shared
// intent with the frontend icon registry (ui/src/consts/permissionIcons.tsx).
// An empty icon is always allowed and resolves to a neutral default in the UI.
var ICON_KEYS = map[string]struct{}{
	"file":       {},
	"folder":     {},
	"database":   {},
	"box":        {},
	"image":      {},
	"music":      {},
	"video":      {},
	"shield":     {},
	"key":        {},
	"user":       {},
	"users":      {},
	"globe":      {},
	"tag":        {},
	"lock":       {},
	"bell":       {},
	"star":       {},
	"settings":   {},
	"server":     {},
	"cloud":      {},
	"code":       {},
	"layers":     {},
	"cloud-rain": {},
	"isme":       {},
}

// IsValidIcon reports whether the icon key is allowed. The empty string
// (neutral default) is valid; any other value must be in the allowlist.
func IsValidIcon(icon string) bool {
	if icon == "" {
		return true
	}
	_, ok := ICON_KEYS[icon]
	return ok
}

// PERMISSION_ICON_KEYS aliases ICON_KEYS — kept for the permission call sites.
var PERMISSION_ICON_KEYS = ICON_KEYS

// IsValidPermissionIcon aliases IsValidIcon — kept for the permission call sites.
func IsValidPermissionIcon(icon string) bool {
	return IsValidIcon(icon)
}

// COLOR_KEYS is the shared allowlist of color palette keys a caller may store on
// any entity that carries a color (app_service tile, role chip, permission
// resource badge). The value is a palette key (not a hex); the frontend maps it
// to a hex (ui/src/consts/appColors.ts). An empty color is always allowed and
// resolves to a neutral fallback in the UI.
var COLOR_KEYS = map[string]struct{}{
	"violet":  {},
	"indigo":  {},
	"cyan":    {},
	"sky":     {},
	"teal":    {},
	"mint":    {},
	"amber":   {},
	"rose":    {},
	"magenta": {},
}

// IsValidColor reports whether the color palette key is allowed. The empty
// string (neutral fallback) is valid; any other value must be in the allowlist.
func IsValidColor(color string) bool {
	if color == "" {
		return true
	}
	_, ok := COLOR_KEYS[color]
	return ok
}
