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
