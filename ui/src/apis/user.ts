import type {
	CreateInvitationRequest,
	CreateInvitationResponse,
	InvitationListItem,
	ListUsersRequest,
	ListUsersResponse,
	UserSessionItem,
	UserStatus,
} from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/** kuery http/base.Response envelope — every endpoint wraps its payload in `data`. */
interface Envelope<T> {
	code: number;
	message: string;
	data: T;
}

const STUB_DELAY_MS = 250;

const stub = <T>(data: T): Promise<T> =>
	new Promise((resolve) => {
		setTimeout(() => resolve(data), STUB_DELAY_MS);
	});

export const listUsers = async (request: ListUsersRequest): Promise<ListUsersResponse> => {
	const response = await apiClient.get<Envelope<ListUsersResponse>>(API_ENDPOINTS.USERS, {
		params: {
			page: request.page,
			size: request.size,
			query: request.query || undefined,
			status: request.status || undefined,
			app: request.app || undefined,
			role: request.role || undefined,
			verified: request.verified,
		},
	});
	const data = response.data.data;
	return { items: data.items ?? [], total: data.total, page: data.page };
};

export const updateUserStatus = async (userId: string, status: UserStatus): Promise<void> => {
	await apiClient.patch(API_ENDPOINTS.USER_STATUS(userId), { status });
};

/** One-way: flips is_verified to true and unblocks login — there is no unverify. */
export const verifyUser = async (userId: string): Promise<void> => {
	await apiClient.post(API_ENDPOINTS.USER_VERIFY(userId));
};

export const softDeleteUser = async (userId: string): Promise<void> => {
	await apiClient.delete(API_ENDPOINTS.USER_DETAIL(userId));
};

export const listUserSessions = async (userId: string): Promise<UserSessionItem[]> => {
	const response = await apiClient.get<Envelope<UserSessionItem[]>>(API_ENDPOINTS.USER_SESSIONS(userId));
	return response.data.data ?? [];
};

export const revokeUserSession = async (userId: string, sessionId: string): Promise<void> => {
	await apiClient.post(API_ENDPOINTS.USER_SESSION_REVOKE(userId, sessionId));
};

export const resetUserPassword = async (userId: string): Promise<void> => {
	// TODO(backend): deferred by user decision — needs email infra to send the
	// reset link. The Users screen disables this row action until it lands.
	void userId;
	return stub(undefined);
};

/** The returned invite_link carries the raw token EXACTLY ONCE — it can never be re-displayed. */
export const createInvitation = async (request: CreateInvitationRequest): Promise<CreateInvitationResponse> => {
	const response = await apiClient.post<Envelope<CreateInvitationResponse>>(API_ENDPOINTS.USER_INVITES, request);
	return response.data.data;
};

export const listInvitations = async (): Promise<InvitationListItem[]> => {
	const response = await apiClient.get<Envelope<{ items: InvitationListItem[] }>>(API_ENDPOINTS.USER_INVITES);
	return response.data.data?.items ?? [];
};

export const revokeInvitation = async (invitationId: string): Promise<void> => {
	await apiClient.post(API_ENDPOINTS.USER_INVITE_REVOKE(invitationId));
};
