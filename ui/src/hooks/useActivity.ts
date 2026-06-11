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
 *  shape (data + loading/error) and exposing a refetch. */
export const useActivity = (): UseActivityResult => {
	const [activity, setActivity] = useState<ActivityItem[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<Error | null>(null);

	const load = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const items = await getMyActivity();
			setActivity(items);
		} catch (err) {
			setError(err instanceof Error ? err : new Error("Failed to load activity"));
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void load();
	}, [load]);

	return { activity, loading, error, refetch: load };
};
