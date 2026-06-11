import type {
	AddRoleMembersRequest,
	CreatePermissionsRequest,
	UpdatePermissionAppearanceRequest,
	CreateRoleRequest,
	CreateRoleResponse,
	ListRoleMembersRequest,
	ListRoleMembersResponse,
	PermissionItem,
	PermissionPair,
	RoleDetailResponse,
	RoleListFilter,
	RoleListItem,
	UpdateRoleRequest,
} from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/** kuery http/base.Response envelope — every endpoint wraps its payload in `data`. */
interface Envelope<T> {
	code: number;
	message: string;
	data: T;
}

/** Lists roles, optionally scoped to one owning app (app_id/app_code).
 *  Empty filter returns roles across all apps. */
export const listRoles = async (filter?: RoleListFilter): Promise<RoleListItem[]> => {
	const response = await apiClient.get<Envelope<RoleListItem[]>>(API_ENDPOINTS.ROLES, {
		params: filter?.app_id || filter?.app_code ? { app_id: filter.app_id, app_code: filter.app_code } : undefined,
	});
	return response.data.data ?? [];
};

export const getRole = async (roleId: string): Promise<RoleDetailResponse> => {
	const response = await apiClient.get<Envelope<RoleDetailResponse>>(API_ENDPOINTS.ROLE_DETAIL(roleId));
	return response.data.data;
};

export const createRole = async (request: CreateRoleRequest): Promise<CreateRoleResponse> => {
	const response = await apiClient.post<Envelope<CreateRoleResponse>>(API_ENDPOINTS.ROLES, request);
	return response.data.data;
};

export const updateRole = async (roleId: string, request: UpdateRoleRequest): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.ROLE_DETAIL(roleId), request);
};

export const deleteRole = async (roleId: string): Promise<void> => {
	await apiClient.delete(API_ENDPOINTS.ROLE_DETAIL(roleId));
};

/** Lists the permission catalog, optionally scoped to one owning app. */
export const listPermissions = async (filter?: RoleListFilter): Promise<PermissionItem[]> => {
	const response = await apiClient.get<Envelope<PermissionItem[]>>(API_ENDPOINTS.PERMISSIONS, {
		params: filter?.app_id || filter?.app_code ? { app_id: filter.app_id, app_code: filter.app_code } : undefined,
	});
	return response.data.data ?? [];
};

export const setRolePermissions = async (roleId: string, permissionIds: number[]): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.ROLE_PERMISSIONS(roleId), { permission_ids: permissionIds });
};

/** Creates resource:action permissions in an app's catalog (rejected for the
 *  isme system app). Returns the created/resulting permission items with ids. */
export const createPermissions = async (appId: string, permissions: PermissionPair[]): Promise<PermissionItem[]> => {
	const request: CreatePermissionsRequest = { app_id: appId, permissions };
	const response = await apiClient.post<Envelope<PermissionItem[]>>(API_ENDPOINTS.PERMISSIONS, request);
	return response.data.data ?? [];
};

/** Deletes a permission from an app's catalog (rejected for the isme system
 *  app). The backend clears any role grants referencing it first. */
export const deletePermission = async (permissionId: number): Promise<void> => {
	await apiClient.delete(API_ENDPOINTS.PERMISSION_DETAIL(permissionId));
};

/** Updates a resource's per-resource appearance (icon + color) across every
 *  catalog row of that (app_id, resource). Rejected for the isme system app. */
export const updatePermissionAppearance = async (request: UpdatePermissionAppearanceRequest): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.PERMISSION_APPEARANCE, request);
};

export const listRoleMembers = async (
	roleId: string,
	request: ListRoleMembersRequest
): Promise<ListRoleMembersResponse> => {
	const response = await apiClient.get<Envelope<ListRoleMembersResponse>>(API_ENDPOINTS.ROLE_MEMBERS(roleId), {
		params: request,
	});
	return response.data.data;
};

export const addRoleMembers = async (roleId: string, request: AddRoleMembersRequest): Promise<void> => {
	await apiClient.post(API_ENDPOINTS.ROLE_MEMBERS(roleId), request);
};

export const removeRoleMember = async (
	roleId: string,
	userId: string,
	appServiceId?: string | null
): Promise<void> => {
	await apiClient.delete(API_ENDPOINTS.ROLE_MEMBER_DETAIL(roleId, userId), {
		params: appServiceId ? { app_service_id: appServiceId } : undefined,
	});
};

/**
 * Bulk-assign an isme-app role (by code) to users — resolves the role id via
 * the isme-scoped roles list, then adds members globally. Used by the Users
 * screen bulk bar (platform roles live on the isme app).
 */
export const assignUserRole = async (userIds: string[], roleCode: string): Promise<void> => {
	const roles = await listRoles({ app_code: "isme" });
	const role = roles.find((item) => item.code === roleCode);
	if (!role) {
		throw new Error(`role "${roleCode}" not found`);
	}
	await addRoleMembers(role.id, { user_ids: userIds });
};
