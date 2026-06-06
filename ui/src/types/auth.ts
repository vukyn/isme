export interface LoginRequest {
	email: string;
	password: string;
	session_id?: string;
}

export interface LoginResponse {
	data: {
		access_token: string;
		refresh_token: string;
		expires_at: string;
		redirect_url?: string;
		authorization_code?: string;
	};
}

export interface SignupRequest {
	name: string;
	email: string;
	password: string;
}

export interface SignupResponse {
	data: {
		id: string;
	};
}

export interface RefreshTokenRequest {
	refresh_token: string;
}

export interface RefreshTokenResponse {
	data: {
		access_token: string;
		refresh_token: string;
		expires_at: string;
	};
}

export interface LogoutResponse {
	data: {
		message: string;
	};
}

export interface AuthTokens {
	access_token: string;
	refresh_token: string;
	expires_at: string;
}

/** Maps to internal/domains/auth/models.GetMeResponse. */
export interface GetMeResponse {
	id: string;
	name: string;
	email: string;
	/** omitempty on the backend — absent for non-admins. */
	is_admin?: boolean;
	/** Global role codes of the current user (RBAC). */
	roles: string[];
	/** Permission codes as "resource:action" (e.g. "role:assign"). Admins bypass. */
	perms: string[];
}
