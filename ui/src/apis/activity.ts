import type { ActivityItem } from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/** kuery http/base.Response envelope — every endpoint wraps its payload in `data`. */
interface Envelope<T> {
	code: number;
	message: string;
	data: T;
}

/** Recent activity of the signed-in user (newest first). The list is `[]` when
 *  none. `limit` is clamped server-side to the configured default/max. */
export const getMyActivity = async (limit?: number): Promise<ActivityItem[]> => {
	const response = await apiClient.get<Envelope<ActivityItem[]>>(API_ENDPOINTS.AUTH_MY_ACTIVITY, {
		params: { limit },
	});
	return response.data.data ?? [];
};
