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

/** One app-scoped role a user holds — rendered as an `app_code:role_code` chip.
 *  Maps to internal/domains/user/models.AppRole. */
export interface AppRole {
	app_code: string;
	app_name: string;
	role_code: string;
	role_name: string;
}

/** Maps to internal/domains/user/models.UserListItem.
 *  is_admin is GONE (app-owned RBAC): platform admin = the `admin` role on the
 *  isme app, surfaced via `roles` below. */
export interface UserListItem {
	id: string;
	name: string;
	email: string;
	status: UserStatus;
	/** One-way verification flag — false blocks login ("account pending verification"). */
	is_verified: boolean;
	/** Full set of app-scoped roles the user holds (models.UserListItem.Roles).
	 *  Empty = the user holds no roles. The "Roles (by app)" column renders one
	 *  chip per entry. */
	roles: AppRole[];
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
	/** App-scope for the role filter (app_code). Role codes are not unique across
	 *  apps under app-owned RBAC, so a role filter is only honoured with an app. */
	app?: string;
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

/** One app-scoped role assignment carried by an invitation — pairs a role with
 *  the app it is scoped to. Maps to models.RoleAssignment. */
export interface InvitationRoleAssignment {
	role_id: string;
	app_service_id: string;
}

/** Maps to internal/domains/user_invitation/models.CreateRequest — email plus
 *  one or more app-scoped role assignments. */
export interface CreateInvitationRequest {
	email: string;
	assignments: InvitationRoleAssignment[];
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

/** One assigned role on a listed invitation. Maps to models.AssignmentItem. */
export interface InvitationAssignmentItem {
	role_id: string;
	role_name: string;
	role_code: string;
	app_service_id: string;
	app_code: string;
	app_name: string;
}

/** Maps to internal/domains/user_invitation/models.InvitationListItem. */
export interface InvitationListItem {
	id: string;
	email: string;
	/** Full set of app-scoped roles this invitation grants on accept. */
	assignments: InvitationAssignmentItem[];
	status: InvitationStatus;
	expires_at: string;
	/** RFC3339; empty when not accepted. */
	accepted_at: string;
	created_at: string;
}
