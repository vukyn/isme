/** Maps to internal/domains/activity/models.ActivityItem.
 *  A single recent-activity entry. The server emits a STRUCTURED record
 *  (type + meta); the frontend composes the display copy/icon/tone from the
 *  type — there is no server-composed body. */
export interface ActivityItem {
	id: string;
	/** One of: sign_in | sign_out | password_changed | invitation_sent. */
	type: string;
	/** Type-specific payload (e.g. { device, client_ip } for sign_in). */
	meta: Record<string, unknown>;
	/** RFC3339; empty = unknown. */
	created_at: string;
}
