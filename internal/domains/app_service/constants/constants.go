package constants

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

// APP_SERVICE_COLOR_KEYS is the fixed aurora palette allowlist for an app tile
// color. The value is a palette key (not a hex); the frontend maps it to a hex.
// An empty color is always allowed and resolves to a neutral fallback in the UI.
var APP_SERVICE_COLOR_KEYS = map[string]struct{}{
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
	_, ok := APP_SERVICE_COLOR_KEYS[color]
	return ok
}
