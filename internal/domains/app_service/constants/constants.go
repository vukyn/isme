package constants

import (
	roleConstants "github.com/vukyn/isme/internal/domains/role/constants"
)

const (
	AppServiceStatusActive     = 1
	AppServiceStatusInactive   = 2
	AppServiceStatusTerminated = 3

	CtxInfoAuthen     = "authen"
	CtxInfoAppService = "app_service"

	// PlatformAppID is the seeded isme self-app (see migration 014). It owns the
	// original RBAC catalog and is the platform's own auth service — it is
	// READ-ONLY: status, secret rotation, and appearance edits are all rejected.
	PlatformAppID   = "app_isme"
	PlatformAppCode = "isme"
)

// IsPlatformApp reports whether the app service is the read-only isme self-app
// (matched by either id or app_code), which may not be mutated.
func IsPlatformApp(idOrCode string) bool {
	return idOrCode == PlatformAppID || idOrCode == PlatformAppCode
}

var AllowedCtxInfos = map[string]struct{}{
	CtxInfoAuthen:     {},
	CtxInfoAppService: {},
}

// APP_SERVICE_COLOR_KEYS aliases the shared color allowlist hoisted into
// role/constants (the canonical source, shared with role + permission
// appearance). Kept for the app_service call sites.
var APP_SERVICE_COLOR_KEYS = roleConstants.COLOR_KEYS

// IsValidColor aliases the shared roleConstants.IsValidColor — kept for the
// app_service call sites. The empty string (neutral fallback) is valid.
func IsValidColor(color string) bool {
	return roleConstants.IsValidColor(color)
}
