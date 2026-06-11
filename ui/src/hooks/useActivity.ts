import { useCallback, useEffect, useState } from "react";
import type { ActivityItem } from "@/types";
import { getMyActivity } from "@/apis";

interface UseActivityResult {
	activity: ActivityItem[];
	loading: boolean;
	error: Error | null;
	/** Re-fetch the feed. */
	refetch: () => Promise<void>;
}

/** Loads the signed-in user's recent activity feed, mirroring the useSessions
 *  shape (data + loading/error) and exposing a refetch. `limit` is clamped
 *  server-side to the configured max (50); omit it for the default short feed. */
export const useActivity = (limit?: number): UseActivityResult => {
	const [activity, setActivity] = useState<ActivityItem[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<Error | null>(null);

	const load = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const items = await getMyActivity(limit);
			setActivity(items);
		} catch (err) {
			setError(err instanceof Error ? err : new Error("Failed to load activity"));
		} finally {
			setLoading(false);
		}
	}, [limit]);

	useEffect(() => {
		void load();
	}, [load]);

	return { activity, loading, error, refetch: load };
};
