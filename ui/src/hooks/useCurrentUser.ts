import { useEffect, useState } from "react";
import { getCurrentUser } from "@/apis/auth";
import type { GetMeResponse } from "@/types";

let cached: GetMeResponse | null = null;
let inFlight: Promise<GetMeResponse> | null = null;

const fetchOnce = (): Promise<GetMeResponse> => {
	if (cached) return Promise.resolve(cached);
	if (inFlight) return inFlight;
	inFlight = getCurrentUser()
		.then((r) => {
			cached = (r as { data: GetMeResponse }).data;
			return cached;
		})
		.finally(() => {
			inFlight = null;
		});
	return inFlight;
};

export const clearCurrentUser = () => {
	cached = null;
};

interface UseCurrentUserReturn {
	user: GetMeResponse | null;
	loading: boolean;
	error: string | null;
}

export const useCurrentUser = (): UseCurrentUserReturn => {
	const [user, setUser] = useState<GetMeResponse | null>(cached);
	const [loading, setLoading] = useState(!cached);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		if (cached) return;
		let active = true;
		fetchOnce()
			.then((u) => {
				if (active) setUser(u);
			})
			.catch((err: unknown) => {
				if (active) setError(err instanceof Error ? err.message : "Failed to load user data");
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, []);

	return { user, loading, error };
};
