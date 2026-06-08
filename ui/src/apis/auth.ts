import type {
	LoginRequest,
	LoginResponse,
	SignupRequest,
	SignupResponse,
	RefreshTokenRequest,
	RefreshTokenResponse,
	LogoutResponse,
	GetMeResponse,
	InviteDetailResponse,
	AcceptInviteRequest,
} from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

export const login = async (data: LoginRequest): Promise<LoginResponse> => {
	const response = await apiClient.post<LoginResponse>(API_ENDPOINTS.AUTH_LOGIN, data);
	return response.data;
};

export const signup = async (data: SignupRequest): Promise<SignupResponse> => {
	const response = await apiClient.post<SignupResponse>(API_ENDPOINTS.AUTH_SIGNUP, data);
	return response.data;
};

export const getCurrentUser = async (): Promise<{ data: GetMeResponse }> => {
	const response = await apiClient.get<{ data: GetMeResponse }>(API_ENDPOINTS.AUTH_ME);
	return response.data;
};

export const refreshToken = async (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
	const response = await apiClient.post<RefreshTokenResponse>(API_ENDPOINTS.AUTH_REFRESH, data);
	return response.data;
};

export const logout = async (): Promise<LogoutResponse> => {
	const response = await apiClient.post<LogoutResponse>(API_ENDPOINTS.AUTH_LOGOUT);
	return response.data;
};

/** Public — resolves an invite token to its email + role name (404 when invalid/expired/used). */
export const getInvitationByToken = async (token: string): Promise<InviteDetailResponse> => {
	const response = await apiClient.get<{ data: InviteDetailResponse }>(API_ENDPOINTS.AUTH_INVITE_DETAIL(token));
	return response.data.data;
};

/** Public — consumes the one-time invite and creates the account (verified + active). */
export const acceptInvite = async (data: AcceptInviteRequest): Promise<void> => {
	await apiClient.post(API_ENDPOINTS.AUTH_ACCEPT_INVITE, data);
};
