import { useCallback, useMemo } from "react";
import { getTokens, decodeJwtPayload } from "@/utils";
import type { AccessTokenClaims } from "@/utils";

interface UsePermissionsReturn {
	/** True when the current user holds the permission (admins bypass). */
	can: (permission: string) => boolean;
	isAdmin: boolean;
	loading: boolean;
}

/**
 * Permission gate derived from the access token JWT claims
 * (`adm` boolean + `perms` string array — the backend is the
 * single source of authorization, no API call needed).
 * Permission codes are "resource:action" — e.g. can("role:assign").
 */
export const usePermissions = (): UsePermissionsReturn => {
	const claims = useMemo<AccessTokenClaims | null>(() => {
		const { access_token } = getTokens();
		return decodeJwtPayload(access_token);
	}, []);

	const can = useCallback(
		(permission: string) => {
			if (!claims) return false;
			if (claims.adm) return true;
			return (claims.perms ?? []).includes(permission);
		},
		[claims]
	);

	return { can, isAdmin: !!claims?.adm, loading: false };
};
