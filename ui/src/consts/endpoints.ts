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

	// User
	USER_ME: "/api/user/me",
} as const;
