package constants

import (
	"github.com/vukyn/kuery/rbac"
)

// Permission codes (must match the catalog seeded by 010_seed_rbac)
var (
	PERM_USER_READ           = rbac.Perm("user", "read")
	PERM_USER_CREATE         = rbac.Perm("user", "create")
	PERM_USER_UPDATE         = rbac.Perm("user", "update")
	PERM_USER_DELETE         = rbac.Perm("user", "delete")
	PERM_USER_RESET_PASSWORD = rbac.Perm("user", "reset_password")

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

// System role codes
const (
	ROLE_CODE_ADMIN  = "admin"
	ROLE_CODE_MEMBER = "member"
	ROLE_CODE_VIEWER = "viewer"
)

// System role IDs (deterministic, seeded by 010_seed_rbac)
const (
	ROLE_ID_ADMIN  = "rol_admin"
	ROLE_ID_MEMBER = "rol_member"
	ROLE_ID_VIEWER = "rol_viewer"
)
