/**
 * RBAC types — mirror internal/domains/role/models/role.go JSON tags.
 */

/** Maps to models.RoleListItem (GET /api/v1/roles). RBAC is app-owned: every
 *  role belongs to one app_service (app_id/app_code). */
export interface RoleListItem {
	id: string;
	app_id: string;
	app_code: string;
	code: string;
	name: string;
	description: string;
	/** Icon key (allowlist in consts/permissionIcons); empty = neutral default. */
	icon: string;
	/** Color palette key (allowlist in consts/appColors); empty = neutral fallback. */
	color: string;
	is_system: boolean;
	members_count: number;
}

/** Maps to models.PermissionItem (GET /api/v1/permissions + role detail).
 *  app_id ties the permission to its owning app; resource:action is the claim
 *  (never app-prefixed). */
export interface PermissionItem {
	id: number;
	app_id: string;
	resource: string;
	action: string;
	/** Per-resource icon key (allowlist in consts/permissionIcons). Shared by all
	 *  rows of the same (app_id, resource); empty = neutral default. */
	icon: string;
	/** Per-resource color palette key (allowlist in consts/appColors). Shared by
	 *  all rows of the same (app_id, resource); empty = neutral fallback. */
	color: string;
}

/** Maps to models.RoleDetailResponse (GET /api/v1/roles/:id). */
export interface RoleDetailResponse {
	id: string;
	app_id: string;
	app_code: string;
	code: string;
	name: string;
	description: string;
	/** Icon key (allowlist in consts/permissionIcons); empty = neutral default. */
	icon: string;
	/** Color palette key (allowlist in consts/appColors); empty = neutral fallback. */
	color: string;
	is_system: boolean;
	permissions: PermissionItem[];
}

/** Optional app filter for the role + permission list endpoints (empty = all apps). */
export interface RoleListFilter {
	app_id?: string;
	app_code?: string;
}

/** Maps to models.CreateRequest (POST /api/v1/roles). The new role belongs to
 *  the selected app (app_id is required and immutable). */
export interface CreateRoleRequest {
	app_id: string;
	code: string;
	name: string;
	description: string;
	clone_from_role_id?: string;
	/** Optional icon key (allowlist in consts/permissionIcons); empty = neutral default. */
	icon?: string;
	/** Optional color palette key (allowlist in consts/appColors); empty = neutral fallback. */
	color?: string;
}

/** Maps to models.CreateResponse. */
export interface CreateRoleResponse {
	id: string;
}

/** Maps to models.UpdateRequest (PUT /api/v1/roles/:id). */
export interface UpdateRoleRequest {
	name: string;
	description: string;
	/** Icon key (allowlist in consts/permissionIcons); empty = neutral default. */
	icon?: string;
	/** Color palette key (allowlist in consts/appColors); empty = neutral fallback. */
	color?: string;
}

/** Maps to models.SetPermissionsRequest (PUT /api/v1/roles/:id/permissions). */
export interface SetRolePermissionsRequest {
	permission_ids: number[];
}

/** One resource:action pair to create — maps to models.PermissionPair. */
export interface PermissionPair {
	resource: string;
	action: string;
	/** Optional per-resource icon key. When the resource already exists in the
	 *  app, the backend reuses that resource's stored icon and ignores this. */
	icon?: string;
	/** Optional per-resource color palette key. When the resource already exists
	 *  in the app, the backend reuses that resource's stored color and ignores this. */
	color?: string;
}

/** Maps to models.CreatePermissionsRequest (POST /api/v1/permissions). Creates
 *  resource:action permissions in an app's catalog (rejected for the isme app). */
export interface CreatePermissionsRequest {
	app_id: string;
	permissions: PermissionPair[];
}

/** Maps to models.UpdatePermissionAppearanceRequest (PUT /api/v1/permissions/appearance).
 *  Changes a resource's per-resource icon + color across all its catalog rows
 *  (rejected for the isme system app). icon/color are allowlist keys; empty allowed. */
export interface UpdatePermissionAppearanceRequest {
	app_id: string;
	resource: string;
	icon: string;
	color: string;
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
