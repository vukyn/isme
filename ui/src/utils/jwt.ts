/**
 * JWT payload decoding (no signature verification — display/gating only,
 * the backend always re-checks permissions server-side).
 */

/** Claims embedded in the isme access token (kuery/claims keys). */
export interface AccessTokenClaims {
	/** Token ID (jti). */
	jti?: string;
	/** User ID. */
	uid?: string;
	email?: string;
	/** Unix expiry (seconds). */
	exp?: number;
	/** Admin flag — admins bypass all permission checks. */
	adm?: boolean;
	/** Permission codes as "resource:action" (e.g. "role:assign"). */
	perms?: string[];
	/** Role codes. */
	roles?: string[];
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
