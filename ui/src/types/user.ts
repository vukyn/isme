export interface User {
	id: string;
	name: string;
	email: string;
	avatar: string;
}

/**
 * Maps to internal/domains/user/entity.User.status:
 * 1 = active, 2 = inactive.
 * 3 (pending) is forward-looking — invite sent, no password set yet;
 * the backend never returns it today (and rejects it as a list filter).
 */
export type UserStatus = 1 | 2 | 3;

/** Maps to internal/domains/user/models.UserListItem. */
export interface UserListItem {
	id: string;
	name: string;
	email: string;
	status: UserStatus;
	is_admin: boolean;
	/** One-way verification flag — false blocks login ("account pending verification"). */
	is_verified: boolean;
	/** Global RBAC role code; omitted when the user has no global role. */
	role?: string;
	sessions_count: number;
	/** RFC3339; empty / zero time = never logged in. */
	last_login_at: string;
	created_at: string;
}

/** Query params for GET /api/v1/users (kuery http/base.Pagination + filters). */
export interface ListUsersRequest {
	page: number;
	size: number;
	query?: string;
	status?: UserStatus;
	role?: string;
	/** Filter by verification state — omitted = all. */
	verified?: boolean;
}

export interface ListUsersResponse {
	items: UserListItem[];
	total: number;
	page: number;
}

/** Maps to internal/domains/user_session/entity.UserSession. */
export interface UserSessionItem {
	id: string;
	client_ip: string;
	user_agent: string;
	last_login_at: string;
	expires_at: string;
	status: number;
}

/** Maps to internal/domains/user_invitation/models.CreateRequest. */
export interface CreateInvitationRequest {
	email: string;
	role_id: string;
}

/** One-time invite link — the raw token is never re-displayable after creation. */
export interface CreateInvitationResponse {
	id: string;
	invite_link: string;
}

/**
 * Stored invitation status (1 pending / 2 accepted / 3 revoked).
 * "expired" is never stored — it is derived client-side from expires_at
 * on pending rows.
 */
export type InvitationStatus = 1 | 2 | 3;

export type InvitationDisplayStatus = "pending" | "accepted" | "revoked" | "expired";

export const INVITATION_STATUS_LABELS: Record<InvitationStatus, InvitationDisplayStatus> = {
	1: "pending",
	2: "accepted",
	3: "revoked",
};

/** Maps to internal/domains/user_invitation/models.InvitationListItem. */
export interface InvitationListItem {
	id: string;
	email: string;
	role_id: string;
	role_name: string;
	status: InvitationStatus;
	expires_at: string;
	/** RFC3339; empty when not accepted. */
	accepted_at: string;
	created_at: string;
}
