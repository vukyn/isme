/** Maps to internal/domains/auth/models.MySessionItem.
 *  Self-scoped active session of the signed-in user. The backend never exposes
 *  refresh_token / token_id. */
export interface MySessionItem {
	id: string;
	client_ip: string;
	user_agent: string;
	/** RFC3339; empty = unknown. */
	last_login_at: string;
	/** RFC3339; empty = unknown. */
	expires_at: string;
	/** Owning app service id; empty = first-party isme session. */
	app_service_id: string;
	/** Owning app display name; empty for first-party isme or a deleted app. */
	app_name: string;
	/** Owning app stored icon key; empty for first-party isme or a deleted app. */
	app_icon: string;
	/** Owning app stored color key; empty for first-party isme or a deleted app. */
	app_color: string;
	/** True for the session the current request is authenticated with. */
	current: boolean;
	/** Lifetime token-rotation count for this session. */
	refresh_count: number;
	/** RFC3339 of the most recent rotation; empty = never refreshed. */
	last_refreshed_at: string;
}

/** Maps to internal/domains/auth/models.MySessionCount. */
export interface MySessionCount {
	count: number;
	new_in_24h: number;
	/** Accurate sliding-24h token-rotation count for the user. */
	rotations_24h: number;
	/** RFC3339 of the user's most recent rotation; empty = no refreshes yet. */
	last_refreshed_at: string;
}
