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

/** One assigned role in the public invite detail — app identity, role identity,
 *  and the resource:action permission preview. Maps to models.InviteAssignmentDetail. */
export interface InviteAssignmentDetail {
	app_code: string;
	app_name: string;
	role_name: string;
	role_code: string;
	/** resource:action codes this role grants. */
	permissions: string[];
}

/** Display state driving the accept page — independent of the raw status int. */
export type InviteDisplayStatus = "valid" | "expired" | "accepted" | "revoked";

/** Maps to internal/domains/user_invitation/models.InviteDetailResponse. */
export interface InviteDetailResponse {
	email: string;
	/** Raw invitation status (1=pending, 2=accepted, 3=revoked). */
	status: number;
	/** Stable string the page switches on: valid | expired | accepted | revoked. */
	display_status: InviteDisplayStatus;
	assignments: InviteAssignmentDetail[];
}

/** Maps to internal/domains/user_invitation/models.AcceptRequest. */
export interface AcceptInviteRequest {
	token: string;
	name: string;
	password: string;
}

export interface AuthTokens {
	access_token: string;
	refresh_token: string;
	expires_at: string;
}

/**
 * Maps to internal/domains/auth/models.GetMeResponse.
 * Authorization data (admin flag, permissions) is NOT here —
 * it lives in the access token JWT claims (see utils/jwt.ts).
 */
export interface GetMeResponse {
	id: string;
	name: string;
	email: string;
}
