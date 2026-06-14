package constants

const (
	// Auth
	AUTH_GROUP_NAME               = "auth"
	AUTH_ENDPOINT_LOGIN           = "/login"
	AUTH_ENDPOINT_REFRESH         = "/refresh"
	AUTH_ENDPOINT_ME              = "/me"
	AUTH_ENDPOINT_LOGOUT          = "/logout"
	AUTH_ENDPOINT_CHANGE_PASSWORD = "/change-password"
	AUTH_ENDPOINT_REQUEST_LOGIN   = "/request-login"
	AUTH_ENDPOINT_EXCHANGE_CODE   = "/exchange-code"
	AUTH_ENDPOINT_SSO_CHECK       = "/sso/check"
	AUTH_ENDPOINT_SSO_CONSENT     = "/sso/consent"
	AUTH_ENDPOINT_INVITE_DETAIL   = "/invites/:token"
	AUTH_ENDPOINT_ACCEPT_INVITE   = "/accept-invite"
	// Self-service session management. Register the static /sessions/others and
	// /sessions/count BEFORE /sessions/:id so Fiber's in-order matcher does not
	// swallow them as the :id param.
	AUTH_ENDPOINT_MY_SESSIONS              = "/sessions"
	AUTH_ENDPOINT_MY_SESSIONS_COUNT        = "/sessions/count"
	AUTH_ENDPOINT_REVOKE_MY_OTHER_SESSIONS = "/sessions/others"
	AUTH_ENDPOINT_REVOKE_MY_SESSION        = "/sessions/:id"
	AUTH_ENDPOINT_MY_ACTIVITY              = "/me/activity"

	// App service
	APP_SERVICE_GROUP_NAME        = "app-service"
	APP_SERVICE_ENDPOINT_ROOT     = ""
	APP_SERVICE_ENDPOINT_REGISTER = "/register"
	APP_SERVICE_ENDPOINT_VERIFY   = "/verify"
	APP_SERVICE_ENDPOINT_REFRESH  = "/refresh"
	APP_SERVICE_ENDPOINT_DETAIL   = "/:appServiceID"
	APP_SERVICE_ENDPOINT_STATUS   = "/:appServiceID/status"

	// User
	USER_GROUP_NAME              = "/users"
	USER_ENDPOINT_ROOT           = ""
	USER_ENDPOINT_DETAIL         = "/:userID"
	USER_ENDPOINT_STATUS         = "/:userID/status"
	USER_ENDPOINT_VERIFY         = "/:userID/verify"
	USER_ENDPOINT_SESSIONS       = "/:userID/sessions"
	USER_ENDPOINT_SESSION_REVOKE = "/:userID/sessions/:sessionID/revoke"
	USER_ENDPOINT_INVITES        = "/invites"
	USER_ENDPOINT_INVITE_REVOKE  = "/invites/:invitationID/revoke"

	// Role
	ROLE_GROUP_NAME             = "/roles"
	ROLE_ENDPOINT_ROOT          = ""
	ROLE_ENDPOINT_DETAIL        = "/:roleID"
	ROLE_ENDPOINT_PERMISSIONS   = "/:roleID/permissions"
	ROLE_ENDPOINT_MEMBERS       = "/:roleID/members"
	ROLE_ENDPOINT_MEMBER_DETAIL = "/:roleID/members/:userID"

	// Permissions
	PERMISSION_ENDPOINT_CATALOG    = "/permissions"
	PERMISSION_ENDPOINT_APPEARANCE = "/permissions/appearance"
	PERMISSION_ENDPOINT_DETAIL     = "/permissions/:permissionID"

	// Media (self-service avatar upload proxy to medioa)
	MEDIA_GROUP_NAME      = "/media"
	MEDIA_ENDPOINT_UPLOAD = "/upload"

	// Settings
	SETTINGS_GROUP_NAME                = "/settings"
	SETTINGS_ENDPOINT_SESSION_REVOKE   = "/session-revoke"
	SETTINGS_ENDPOINT_ROTATION_CLEANUP = "/rotation-cleanup"
	SETTINGS_ENDPOINT_ACTIVITY_CLEANUP = "/activity-cleanup"
)
