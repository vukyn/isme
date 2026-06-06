import type {
	AddRoleMembersRequest,
	CreateRoleRequest,
	CreateRoleResponse,
	ListRoleMembersRequest,
	ListRoleMembersResponse,
	PermissionItem,
	RoleDetailResponse,
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

export const listRoles = async (): Promise<RoleListItem[]> => {
	const response = await apiClient.get<Envelope<RoleListItem[]>>(API_ENDPOINTS.ROLES);
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

export const listPermissions = async (): Promise<PermissionItem[]> => {
	const response = await apiClient.get<Envelope<PermissionItem[]>>(API_ENDPOINTS.PERMISSIONS);
	return response.data.data ?? [];
};

export const setRolePermissions = async (roleId: string, permissionIds: number[]): Promise<void> => {
	await apiClient.put(API_ENDPOINTS.ROLE_PERMISSIONS(roleId), { permission_ids: permissionIds });
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
 * Bulk-assign a role (by code) to users — resolves the role id via the roles
 * list, then adds members globally. Used by the Users screen bulk bar.
 */
export const assignUserRole = async (userIds: string[], roleCode: string): Promise<void> => {
	const roles = await listRoles();
	const role = roles.find((item) => item.code === roleCode);
	if (!role) {
		throw new Error(`role "${roleCode}" not found`);
	}
	await addRoleMembers(role.id, { user_ids: userIds });
};
