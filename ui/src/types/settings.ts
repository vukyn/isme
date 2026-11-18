/**
 * Session auto-revoke configuration — the typed config that drives the backend
 * cronjob revoking expired user sessions. Persisted as { enabled, cron } on the
 * SSO service; last_run_at / last_revoked_count are read-only run history.
 */
export interface SessionRevokeConfig {
	enabled: boolean;
	/** 5-field standard cron expression (minute hour day month weekday). */
	cron: string;
	/** Unix seconds of the last scheduler run, or null if it never ran. */
	lastRunAt: number | null;
	/** Sessions revoked on the last run, or null if it never ran. */
	lastRevokedCount: number | null;
}

export interface UpdateSessionRevokeConfigRequest {
	enabled: boolean;
	cron: string;
}

/**
 * Rotation-cleanup configuration — the typed config that drives the backend
 * cronjob pruning token_rotation_events older than a retention window. Persisted
 * as { enabled, cron, retentionHours } on the SSO service; last_run_at /
 * last_cleaned_count are read-only run history.
 */
export interface RotationCleanupConfig {
	enabled: boolean;
	/** 5-field standard cron expression (minute hour day month weekday). */
	cron: string;
	/** Retention window in hours; rows older than this are pruned (floor 24). */
	retentionHours: number;
	/** Unix seconds of the last scheduler run, or null if it never ran. */
	lastRunAt: number | null;
	/** Rotation events pruned on the last run, or null if it never ran. */
	lastCleanedCount: number | null;
}

export interface UpdateRotationCleanupConfigRequest {
	enabled: boolean;
	cron: string;
	retentionHours: number;
}

/**
 * Activity-cleanup configuration — the typed config that drives the backend
 * cronjob pruning activity_events older than a retention window measured in
 * DAYS (not hours like rotation cleanup). Persisted as
 * { enabled, cron, retentionDays } on the SSO service; last_run_at /
 * last_pruned_count are read-only run history.
 */
export interface ActivityCleanupConfig {
	enabled: boolean;
	/** 5-field standard cron expression (minute hour day month weekday). */
	cron: string;
	/** Retention window in DAYS; rows older than this are pruned (floor 7). */
	retentionDays: number;
	/** Unix seconds of the last scheduler run, or null if it never ran. */
	lastRunAt: number | null;
	/** Activity events pruned on the last run, or null if it never ran. */
	lastPrunedCount: number | null;
}

export interface UpdateActivityCleanupConfigRequest {
	enabled: boolean;
	cron: string;
	retentionDays: number;
}

/**
 * Invitation-cleanup configuration — the typed config that drives the backend
 * cronjob marking pending invitations past their expiry as expired. Persisted
 * as { enabled, cron } on the SSO service; last_run_at / last_expired_count are
 * read-only run history.
 */
export interface InvitationCleanupConfig {
	enabled: boolean;
	/** 5-field standard cron expression (minute hour day month weekday). */
	cron: string;
	/** Unix seconds of the last scheduler run, or null if it never ran. */
	lastRunAt: number | null;
	/** Invitations marked expired on the last run, or null if it never ran. */
	lastExpiredCount: number | null;
}

export interface UpdateInvitationCleanupConfigRequest {
	enabled: boolean;
	cron: string;
}

/**
 * Unverified-user-cleanup configuration — the typed config that drives the
 * backend cronjob soft-deleting accounts that never finished sign-up once they
 * are older than a minimum age in DAYS. Persisted as
 * { enabled, cron, minAgeDays } on the SSO service; last_run_at /
 * last_deleted_count are read-only run history.
 */
export interface UnverifiedUserCleanupConfig {
	enabled: boolean;
	/** 5-field standard cron expression (minute hour day month weekday). */
	cron: string;
	/** Minimum account age in DAYS before deletion (floor 1). */
	minAgeDays: number;
	/** Unix seconds of the last scheduler run, or null if it never ran. */
	lastRunAt: number | null;
	/** Accounts deleted on the last run, or null if it never ran. */
	lastDeletedCount: number | null;
}

export interface UpdateUnverifiedUserCleanupConfigRequest {
	enabled: boolean;
	cron: string;
	minAgeDays: number;
}
