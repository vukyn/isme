/**
 * RBAC types — mirror internal/domains/role/models/role.go JSON tags.
 */

/** Maps to models.RoleListItem (GET /api/v1/roles). */
export interface RoleListItem {
	id: string;
	code: string;
	name: string;
	description: string;
	is_system: boolean;
	members_count: number;
}

/** Maps to models.PermissionItem (GET /api/v1/permissions + role detail). */
export interface PermissionItem {
	id: number;
	resource: string;
	action: string;
}

/** Maps to models.RoleDetailResponse (GET /api/v1/roles/:id). */
export interface RoleDetailResponse {
	id: string;
	code: string;
	name: string;
	description: string;
	is_system: boolean;
	permissions: PermissionItem[];
}

/** Maps to models.CreateRequest (POST /api/v1/roles). */
export interface CreateRoleRequest {
	code: string;
	name: string;
	description: string;
	clone_from_role_id?: string;
}

/** Maps to models.CreateResponse. */
export interface CreateRoleResponse {
	id: string;
}

/** Maps to models.UpdateRequest (PUT /api/v1/roles/:id). */
export interface UpdateRoleRequest {
	name: string;
	description: string;
}

/** Maps to models.SetPermissionsRequest (PUT /api/v1/roles/:id/permissions). */
export interface SetRolePermissionsRequest {
	permission_ids: number[];
}

/** Maps to models.AddMembersRequest (POST /api/v1/roles/:id/members). */
export interface AddRoleMembersRequest {
	user_ids: string[];
	/** Scope of the assignment — omitted/null = global (user_roles.app_service_id NULL). */
	app_service_id?: string | null;
}

/** Maps to models.MemberItem. */
export interface RoleMemberItem {
	user_id: string;
	name: string;
	email: string;
	/** null = global assignment. */
	app_service_id: string | null;
	created_at: string;
}

/** Maps to models.ListMembersRequest query params. */
export interface ListRoleMembersRequest {
	page: number;
	size: number;
	query?: string;
}

/** Maps to models.ListMembersResponse (GET /api/v1/roles/:id/members). */
export interface ListRoleMembersResponse {
	items: RoleMemberItem[];
	total: number;
	page: number;
}
