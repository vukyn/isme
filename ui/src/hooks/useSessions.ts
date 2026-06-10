import { useCallback, useEffect, useState } from "react";
import type { MySessionItem } from "@/types";
import { getMySessions } from "@/apis";

interface UseSessionsResult {
	sessions: MySessionItem[];
	loading: boolean;
	error: Error | null;
	/** Re-fetch the list — call after a revoke so the UI reflects the change. */
	refetch: () => Promise<void>;
}

/** Loads the signed-in user's active sessions, mirroring the useUser shape
 *  (data + loading/error) and exposing a refetch for post-revoke refresh. */
export const useSessions = (): UseSessionsResult => {
	const [sessions, setSessions] = useState<MySessionItem[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<Error | null>(null);

	const load = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const items = await getMySessions();
			setSessions(items);
		} catch (err) {
			setError(err instanceof Error ? err : new Error("Failed to load sessions"));
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void load();
	}, [load]);

	return { sessions, loading, error, refetch: load };
};
