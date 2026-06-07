"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getCurrentUser } from "@/apis/auth";
import { getTokens } from "@/utils/axios";
import type { GetMeResponse } from "@/types";
import { UserContext } from "./userContext";

export const UserProvider = ({ children }: { children: React.ReactNode }) => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<Error | null>(null);

	// Dedup concurrent/StrictMode-double fetches into a single in-flight request.
	const inflight = useRef<Promise<GetMeResponse | null> | null>(null);

	// Returns the fetched user (or null) so callers like ProtectedRoute can
	// act on the fresh result instead of stale context state.
	const fetchUser = useCallback(async (): Promise<GetMeResponse | null> => {
		if (inflight.current) return inflight.current;

		const run = (async () => {
			// Yield so state updates run in promise context, not synchronously
			// inside the mounting effect (react-hooks/set-state-in-effect).
			await Promise.resolve();

			// Skip the network call when no session exists (e.g. login page),
			// otherwise /me 401s and the axios interceptor bounces to /login.
			if (!getTokens().access_token) {
				setUser(null);
				setError(null);
				setLoading(false);
				inflight.current = null;
				return null;
			}

			try {
				setLoading(true);
				setError(null);
				const response = await getCurrentUser();
				setUser(response.data);
				return response.data;
			} catch (err) {
				setError(err instanceof Error ? err : new Error("Failed to fetch user"));
				setUser(null);
				return null;
			} finally {
				setLoading(false);
				inflight.current = null;
			}
		})();

		inflight.current = run;
		return run;
	}, []);

	// Drop the cached user (e.g. on logout) without firing a network call.
	const clear = useCallback(() => {
		setUser(null);
		setError(null);
		setLoading(false);
	}, []);

	useEffect(() => {
		void fetchUser();
	}, [fetchUser]);

	return (
		<UserContext.Provider value={{ user, loading, error, refetch: fetchUser, clear }}>
			{children}
		</UserContext.Provider>
	);
};
