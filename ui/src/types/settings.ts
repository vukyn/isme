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
 * Database-backup configuration — the typed config that drives the backend
 * cronjob snapshotting the SQLite database (VACUUM INTO) and keeping the N most
 * recent backup files. Retention here is a COUNT of files (not a time window).
 * Persisted as { enabled, cron, retainCount } on the SSO service; last_run_at /
 * last_backup_path / last_kept_count are read-only run history.
 */
export interface DatabaseBackupConfig {
	enabled: boolean;
	/** 5-field standard cron expression (minute hour day month weekday). */
	cron: string;
	/** Number of backup files to keep; older ones are pruned (1–90). */
	retainCount: number;
	/** Unix seconds of the last scheduler run, or null if it never ran. */
	lastRunAt: number | null;
	/** Path of the most recent backup file, or null if it never ran. */
	lastBackupPath: string | null;
	/** Backup files kept after the last run, or null if it never ran. */
	lastKeptCount: number | null;
}

export interface UpdateDatabaseBackupConfigRequest {
	enabled: boolean;
	cron: string;
	retainCount: number;
}
