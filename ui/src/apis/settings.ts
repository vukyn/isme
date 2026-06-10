import type { SessionRevokeConfig, UpdateSessionRevokeConfigRequest } from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/** kuery http/base.Response envelope — every endpoint wraps its payload in `data`. */
interface Envelope<T> {
	code: number;
	message: string;
	data: T;
}

/** Raw backend shape (snake_case) for the session auto-revoke config. */
interface SessionRevokeConfigDTO {
	enabled: boolean;
	cron: string;
	last_run_at: number | null;
	last_revoked_count: number | null;
}

export const getSessionRevokeConfig = async (): Promise<SessionRevokeConfig> => {
	const response = await apiClient.get<Envelope<SessionRevokeConfigDTO>>(API_ENDPOINTS.SETTINGS_SESSION_REVOKE);
	const dto = response.data.data;
	return {
		enabled: dto.enabled,
		cron: dto.cron,
		lastRunAt: dto.last_run_at ?? null,
		lastRevokedCount: dto.last_revoked_count ?? null,
	};
};

export const updateSessionRevokeConfig = async (request: UpdateSessionRevokeConfigRequest): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.SETTINGS_SESSION_REVOKE, request);
};
