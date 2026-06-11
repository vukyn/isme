/**
 * API Endpoints
 */
export const API_ENDPOINTS = {
	// Auth
	AUTH_LOGIN: "/api/v1/auth/login",
	AUTH_SIGNUP: "/api/v1/auth/signup",
	AUTH_ME: "/api/v1/auth/me",
	AUTH_REFRESH: "/api/v1/auth/refresh",
	AUTH_LOGOUT: "/api/v1/auth/logout",
	AUTH_SSO_CHECK: "/api/v1/auth/sso/check",
	AUTH_SSO_CONSENT: "/api/v1/auth/sso/consent",
	AUTH_INVITE_DETAIL: (token: string) => `/api/v1/auth/invites/${encodeURIComponent(token)}`,
	AUTH_ACCEPT_INVITE: "/api/v1/auth/accept-invite",
	// Self-service session management for the signed-in user.
	AUTH_MY_SESSIONS: "/api/v1/auth/sessions",
	AUTH_MY_SESSIONS_COUNT: "/api/v1/auth/sessions/count",
	AUTH_MY_SESSION_REVOKE: (sessionId: string) => `/api/v1/auth/sessions/${encodeURIComponent(sessionId)}`,
	AUTH_MY_OTHER_SESSIONS_REVOKE: "/api/v1/auth/sessions/others",
	// Self-service recent-activity feed for the signed-in user.
	AUTH_MY_ACTIVITY: "/api/v1/auth/me/activity",

	// User
	USER_ME: "/api/user/me",
	USERS: "/api/v1/users",
	USER_DETAIL: (userId: string) => `/api/v1/users/${userId}`,
	USER_STATUS: (userId: string) => `/api/v1/users/${userId}/status`,
	USER_VERIFY: (userId: string) => `/api/v1/users/${userId}/verify`,
	USER_SESSIONS: (userId: string) => `/api/v1/users/${userId}/sessions`,
	USER_SESSION_REVOKE: (userId: string, sessionId: string) => `/api/v1/users/${userId}/sessions/${sessionId}/revoke`,
	USER_INVITES: "/api/v1/users/invites",
	USER_INVITE_REVOKE: (invitationId: string) => `/api/v1/users/invites/${invitationId}/revoke`,

	// App service
	APP_SERVICES: "/api/v1/app-service",
	APP_SERVICE_REGISTER: "/api/v1/app-service/register",
	APP_SERVICE_VERIFY: "/api/v1/app-service/verify",
	APP_SERVICE_REFRESH: "/api/v1/app-service/refresh",
	APP_SERVICE_DETAIL: (appServiceId: string) => `/api/v1/app-service/${appServiceId}`,
	APP_SERVICE_STATUS: (appServiceId: string) => `/api/v1/app-service/${appServiceId}/status`,

	// Role / RBAC
	ROLES: "/api/v1/roles",
	ROLE_DETAIL: (roleId: string) => `/api/v1/roles/${roleId}`,
	ROLE_PERMISSIONS: (roleId: string) => `/api/v1/roles/${roleId}/permissions`,
	ROLE_MEMBERS: (roleId: string) => `/api/v1/roles/${roleId}/members`,
	ROLE_MEMBER_DETAIL: (roleId: string, userId: string) => `/api/v1/roles/${roleId}/members/${userId}`,
	PERMISSIONS: "/api/v1/permissions",
	PERMISSION_APPEARANCE: "/api/v1/permissions/appearance",
	PERMISSION_DETAIL: (permissionId: number) => `/api/v1/permissions/${permissionId}`,

	// Settings
	SETTINGS_SESSION_REVOKE: "/api/v1/settings/session-revoke",
	SETTINGS_ROTATION_CLEANUP: "/api/v1/settings/rotation-cleanup",
} as const;
