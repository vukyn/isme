package constants

const (
	// Auth
	AUTH_GROUP_NAME               = "auth"
	AUTH_ENDPOINT_LOGIN           = "/login"
	AUTH_ENDPOINT_SIGNUP          = "/signup"
	AUTH_ENDPOINT_REFRESH         = "/refresh"
	AUTH_ENDPOINT_ME              = "/me"
	AUTH_ENDPOINT_LOGOUT          = "/logout"
	AUTH_ENDPOINT_CHANGE_PASSWORD = "/change-password"
	AUTH_ENDPOINT_REQUEST_LOGIN   = "/request-login"
	AUTH_ENDPOINT_EXCHANGE_CODE   = "/exchange-code"

	// App service
	APP_SERVICE_GROUP_NAME        = "app-service"
	APP_SERVICE_ENDPOINT_ROOT     = ""
	APP_SERVICE_ENDPOINT_REGISTER = "/register"
	APP_SERVICE_ENDPOINT_VERIFY   = "/verify"
	APP_SERVICE_ENDPOINT_REFRESH  = "/refresh"
	APP_SERVICE_ENDPOINT_STATUS   = "/:appServiceID/status"

	// User
	USER_GROUP_NAME              = "/users"
	USER_ENDPOINT_ROOT           = ""
	USER_ENDPOINT_DETAIL         = "/:userID"
	USER_ENDPOINT_STATUS         = "/:userID/status"
	USER_ENDPOINT_SESSIONS       = "/:userID/sessions"
	USER_ENDPOINT_SESSION_REVOKE = "/:userID/sessions/:sessionID/revoke"

	// Role
	ROLE_GROUP_NAME             = "/roles"
	ROLE_ENDPOINT_ROOT          = ""
	ROLE_ENDPOINT_DETAIL        = "/:roleID"
	ROLE_ENDPOINT_PERMISSIONS   = "/:roleID/permissions"
	ROLE_ENDPOINT_MEMBERS       = "/:roleID/members"
	ROLE_ENDPOINT_MEMBER_DETAIL = "/:roleID/members/:userID"

	// Permissions
	PERMISSION_ENDPOINT_CATALOG = "/permissions"
)
