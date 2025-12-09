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

export interface GetMeResponse {
	id: string;
	name: string;
	email: string;
}
