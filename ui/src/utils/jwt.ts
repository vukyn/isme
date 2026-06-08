/**
 * JWT payload decoding (no signature verification — display/gating only,
 * the backend always re-checks permissions server-side).
 */

/** Per-app permission grant inside resource_access, keyed by app_code. */
export interface AppAccess {
	/** Permission codes as "resource:action" (e.g. "role:assign"). */
	perms?: string[];
}

/** Claims embedded in the isme access token (kuery/claims keys). */
export interface AccessTokenClaims {
	/** Token ID (jti). */
	jti?: string;
	/** User ID. */
	uid?: string;
	email?: string;
	/** Unix expiry (seconds). */
	exp?: number;
	/** App codes this token is valid for (aud). */
	aud?: string[] | string;
	/** Per-app permission grants, keyed by app_code (e.g. resource_access.isme.perms). */
	resource_access?: Record<string, AppAccess>;
}

/**
 * Decode the payload of a JWT without verifying the signature.
 * Returns null when the token is missing or malformed.
 */
export const decodeJwtPayload = <T = AccessTokenClaims>(token: string | null | undefined): T | null => {
	if (!token) return null;
	const parts = token.split(".");
	if (parts.length !== 3) return null;
	try {
		const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
		const json = decodeURIComponent(
			atob(base64)
				.split("")
				.map((c) => `%${c.charCodeAt(0).toString(16).padStart(2, "0")}`)
				.join("")
		);
		return JSON.parse(json) as T;
	} catch {
		return null;
	}
};
