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
	/** Stored appearance keys for the app tile (empty → fallback look). */
	icon: string;
	color: string;
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

/** Maps to internal/domains/auth/models.SSOCheckRequest — read-only silent-authorize probe. */
export interface SSOCheckRequest {
	session_id: string;
	access_token?: string;
	refresh_token?: string;
}

/** One consent line item rendered on the SSO consent screen. Maps to models.SSOScope. */
export interface SSOScope {
	title: string;
	description: string;
}

/** Maps to internal/domains/auth/models.SSOCheckResponse.
 *  When `valid` is false the page falls back to the password form. */
export interface SSOCheckResponse {
	data: {
		valid: boolean;
		user: {
			name: string;
			email: string;
		};
		app: {
			name: string;
			redirect_url: string;
			app_code: string;
			/** Stored appearance keys for the requesting-app tile (empty → fallback). */
			icon: string;
			color: string;
		};
		scopes: SSOScope[];
		/** Short-TTL single-use CSRF nonce required by /sso/consent. */
		nonce: string;
	};
}

/** Maps to internal/domains/auth/models.SSOConsentRequest. */
export interface SSOConsentRequest {
	session_id: string;
	access_token?: string;
	refresh_token?: string;
	nonce: string;
}

/** Maps to internal/domains/auth/models.SSOConsentResponse. */
export interface SSOConsentResponse {
	data: {
		redirect_url: string;
		authorization_code: string;
	};
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
	/** medioa object URL or pasted external link for the profile photo; empty = no avatar. */
	avatar_url: string;
	/** RFC3339 account-creation time; drives the Welcome "member since" stat. */
	created_at: string;
}

/** Maps to internal/domains/auth/models.UpdateMeRequest — self-service profile update. */
export interface UpdateMeRequest {
	name: string;
	avatar_url: string;
}

/** Maps to internal/domains/auth/models.ChangePasswordRequest. */
export interface ChangePasswordRequest {
	old_password: string;
	new_password: string;
}

/** Maps to internal/domains/media/models.UploadResponse — proxied medioa upload result. */
export interface MediaUploadResponse {
	data: {
		url: string;
		file_id: string;
		file_name: string;
		file_size: number;
	};
}
