import type { MySessionItem, MySessionCount } from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/** kuery http/base.Response envelope — every endpoint wraps its payload in `data`. */
interface Envelope<T> {
	code: number;
	message: string;
	data: T;
}

/** Active sessions of the signed-in user. The list is `[]` when none. */
export const getMySessions = async (): Promise<MySessionItem[]> => {
	const response = await apiClient.get<Envelope<MySessionItem[]>>(API_ENDPOINTS.AUTH_MY_SESSIONS);
	return response.data.data ?? [];
};

/** Total active count + how many appeared in the last 24h (Welcome stat card). */
export const getMySessionsCount = async (): Promise<MySessionCount> => {
	const response = await apiClient.get<Envelope<MySessionCount>>(API_ENDPOINTS.AUTH_MY_SESSIONS_COUNT);
	return response.data.data ?? { count: 0, new_in_24h: 0 };
};

/** Revoke a single OTHER session owned by the caller (ownership-checked server-side). */
export const revokeMySession = async (sessionId: string): Promise<void> => {
	await apiClient.delete(API_ENDPOINTS.AUTH_MY_SESSION_REVOKE(sessionId));
};

/** Revoke every active session except the current one. */
export const revokeMyOtherSessions = async (): Promise<void> => {
	await apiClient.delete(API_ENDPOINTS.AUTH_MY_OTHER_SESSIONS_REVOKE);
};
