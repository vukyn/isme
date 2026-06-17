import type {
	ActivityCleanupConfig,
	DatabaseBackupConfig,
	RotationCleanupConfig,
	SessionRevokeConfig,
	UpdateActivityCleanupConfigRequest,
	UpdateDatabaseBackupConfigRequest,
	UpdateRotationCleanupConfigRequest,
	UpdateSessionRevokeConfigRequest,
} from "@/types";
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

/** Raw backend shape (snake_case) for the rotation-cleanup config. */
interface RotationCleanupConfigDTO {
	enabled: boolean;
	cron: string;
	retention_hours: number;
	last_run_at: number | null;
	last_cleaned_count: number | null;
}

export const getRotationCleanupConfig = async (): Promise<RotationCleanupConfig> => {
	const response = await apiClient.get<Envelope<RotationCleanupConfigDTO>>(API_ENDPOINTS.SETTINGS_ROTATION_CLEANUP);
	const dto = response.data.data;
	return {
		enabled: dto.enabled,
		cron: dto.cron,
		retentionHours: dto.retention_hours,
		lastRunAt: dto.last_run_at ?? null,
		lastCleanedCount: dto.last_cleaned_count ?? null,
	};
};

export const updateRotationCleanupConfig = async (request: UpdateRotationCleanupConfigRequest): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.SETTINGS_ROTATION_CLEANUP, {
		enabled: request.enabled,
		cron: request.cron,
		retention_hours: request.retentionHours,
	});
};

/** Raw backend shape (snake_case) for the activity-cleanup config. */
interface ActivityCleanupConfigDTO {
	enabled: boolean;
	cron: string;
	retention_days: number;
	last_run_at: number | null;
	last_pruned_count: number | null;
}

export const getActivityCleanupConfig = async (): Promise<ActivityCleanupConfig> => {
	const response = await apiClient.get<Envelope<ActivityCleanupConfigDTO>>(API_ENDPOINTS.SETTINGS_ACTIVITY_CLEANUP);
	const dto = response.data.data;
	return {
		enabled: dto.enabled,
		cron: dto.cron,
		retentionDays: dto.retention_days,
		lastRunAt: dto.last_run_at ?? null,
		lastPrunedCount: dto.last_pruned_count ?? null,
	};
};

export const updateActivityCleanupConfig = async (request: UpdateActivityCleanupConfigRequest): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.SETTINGS_ACTIVITY_CLEANUP, {
		enabled: request.enabled,
		cron: request.cron,
		retention_days: request.retentionDays,
	});
};

/** Raw backend shape (snake_case) for the database-backup config. */
interface DatabaseBackupConfigDTO {
	enabled: boolean;
	cron: string;
	retain_count: number;
	last_run_at: number | null;
	last_backup_path: string | null;
	last_kept_count: number | null;
}

export const getDatabaseBackupConfig = async (): Promise<DatabaseBackupConfig> => {
	const response = await apiClient.get<Envelope<DatabaseBackupConfigDTO>>(API_ENDPOINTS.SETTINGS_DATABASE_BACKUP);
	const dto = response.data.data;
	return {
		enabled: dto.enabled,
		cron: dto.cron,
		retainCount: dto.retain_count,
		lastRunAt: dto.last_run_at ?? null,
		lastBackupPath: dto.last_backup_path ?? null,
		lastKeptCount: dto.last_kept_count ?? null,
	};
};

export const updateDatabaseBackupConfig = async (request: UpdateDatabaseBackupConfigRequest): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.SETTINGS_DATABASE_BACKUP, {
		enabled: request.enabled,
		cron: request.cron,
		retain_count: request.retainCount,
	});
};
