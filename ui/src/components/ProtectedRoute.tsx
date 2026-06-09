"use client";

import { useEffect, useRef, useState } from "react";
import { Navigate } from "react-router-dom";
import { Box, Spinner, Stack } from "@chakra-ui/react";
import { getTokens, isTokenExpired, saveTokens } from "@/utils/axios";
import { refreshToken } from "@/apis/auth";
import { useUser } from "@/hooks/useUser";
import { usePermissions } from "@/hooks/usePermissions";

interface ProtectedRouteProps {
	children: React.ReactNode;
	// When set, the user must hold this permission (in `appCode`, default "isme")
	// after auth resolves — otherwise they're redirected to /welcome. UX gating
	// only; the backend still enforces access.
	requiredPermission?: string;
	appCode?: string;
}

export const ProtectedRoute = ({ children, requiredPermission, appCode = "isme" }: ProtectedRouteProps) => {
	const { user, loading, error, refetch } = useUser();
	const { can } = usePermissions();

	// One-shot recovery guard: refresh the token at most once per mount
	// before declaring the session unauthenticated.
	const recoveryAttempted = useRef(false);
	const [recoveryDone, setRecoveryDone] = useState(false);

	useEffect(() => {
		if (loading || user || recoveryAttempted.current) return;

		const tokens = getTokens();
		if (!tokens.access_token) return;
		recoveryAttempted.current = true;

		const recover = async () => {
			try {
				if (error || isTokenExpired()) {
					// Context fetch 401ed or the token expired before any fetch —
					// refresh once, then refetch /me with the new token.
					if (!tokens.refresh_token) return;
					const response = await refreshToken({ refresh_token: tokens.refresh_token });
					saveTokens(response.data);
					await refetch();
				} else {
					// Token present but no user yet (e.g. session created after the
					// provider's initial fetch settled) — fetch once.
					await refetch();
				}
			} catch {
				// Recovery failed — fall through to the unauthenticated redirect.
			} finally {
				setRecoveryDone(true);
			}
		};
		void recover();
	}, [loading, user, error, refetch]);

	const hasAccessToken = !!getTokens().access_token;
	const recoveryPending = hasAccessToken && !user && !recoveryDone;

	if (loading || recoveryPending) {
		return (
			<Box w="full" h="100vh" display="flex" alignItems="center" justifyContent="center">
				<Stack align="center" gap="4">
					<Spinner size="xl" color="accent" />
				</Stack>
			</Box>
		);
	}

	if (!user) {
		return <Navigate to="/login" replace />;
	}

	// Authenticated but lacks the permission this route requires — bounce to the
	// overview rather than letting them reach a screen that would only 403.
	if (requiredPermission && !can(requiredPermission, appCode)) {
		return <Navigate to="/welcome" replace />;
	}

	return <>{children}</>;
};
