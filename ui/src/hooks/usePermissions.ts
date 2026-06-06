import { useCallback } from "react";
import { useCurrentUser } from "./useCurrentUser";

interface UsePermissionsReturn {
	/** True when the current user holds the permission (admins bypass). */
	can: (permission: string) => boolean;
	isAdmin: boolean;
	loading: boolean;
}

/**
 * Permission gate derived from GET /me (roles + perms claims).
 * Permission codes are "resource:action" — e.g. can("role:assign").
 */
export const usePermissions = (): UsePermissionsReturn => {
	const { user, loading } = useCurrentUser();

	const can = useCallback(
		(permission: string) => {
			if (!user) return false;
			if (user.is_admin) return true;
			return (user.perms ?? []).includes(permission);
		},
		[user]
	);

	return { can, isAdmin: !!user?.is_admin, loading };
};
