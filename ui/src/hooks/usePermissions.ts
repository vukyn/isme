import { useCallback, useMemo } from "react";
import { getTokens, decodeJwtPayload } from "@/utils";
import type { AccessTokenClaims } from "@/utils";

interface UsePermissionsReturn {
	/**
	 * True when the current user holds the permission for the given app.
	 * Defaults to the "isme" app, since the admin console manages isme's
	 * own resources (users, roles, app_services).
	 */
	can: (permission: string, appCode?: string) => boolean;
}

/**
 * Permission gate derived from the access token JWT claims
 * (`resource_access` map, keyed by app_code — the backend is the single
 * source of authorization, no API call needed). Permission codes are
 * "resource:action" — e.g. can("role:assign").
 */
export const usePermissions = (): UsePermissionsReturn => {
	const claims = useMemo<AccessTokenClaims | null>(() => {
		const { access_token } = getTokens();
		return decodeJwtPayload(access_token);
	}, []);

	const can = useCallback(
		(permission: string, appCode = "isme") => {
			const perms = claims?.resource_access?.[appCode]?.perms;
			return (perms ?? []).includes(permission);
		},
		[claims]
	);

	return { can };
};
