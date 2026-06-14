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
	SSOCheckRequest,
	SSOCheckResponse,
	SSOConsentRequest,
	SSOConsentResponse,
	UpdateMeRequest,
	ChangePasswordRequest,
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

/** Self-service profile update (display name + avatar URL). Returns the fresh
 *  GetMe payload so callers can refresh the UserProvider cache. */
export const updateProfile = async (data: UpdateMeRequest): Promise<GetMeResponse> => {
	const response = await apiClient.patch<{ data: GetMeResponse }>(API_ENDPOINTS.AUTH_UPDATE_ME, data);
	return response.data.data;
};

/** Self-service password change. The backend revokes all sessions on success,
 *  but the current device stays signed in via its existing tokens. */
export const changePassword = async (data: ChangePasswordRequest): Promise<void> => {
	await apiClient.post(API_ENDPOINTS.AUTH_CHANGE_PASSWORD, data);
};

export const refreshToken = async (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
	const response = await apiClient.post<RefreshTokenResponse>(API_ENDPOINTS.AUTH_REFRESH, data);
	return response.data;
};

export const logout = async (): Promise<LogoutResponse> => {
	const response = await apiClient.post<LogoutResponse>(API_ENDPOINTS.AUTH_LOGOUT);
	return response.data;
};

/** Public — read-only silent-authorize probe. Returns {valid:false} (not an error)
 *  when no live isme session exists, so the page can fall back to the password form. */
export const ssoCheck = async (data: SSOCheckRequest): Promise<SSOCheckResponse> => {
	const response = await apiClient.post<SSOCheckResponse>(API_ENDPOINTS.AUTH_SSO_CHECK, data);
	return response.data;
};

/** Public — authorize step. Requires the single-use nonce from ssoCheck; mints a
 *  one-time authorization_code + redirect_url for the requesting app. */
export const ssoConsent = async (data: SSOConsentRequest): Promise<SSOConsentResponse> => {
	const response = await apiClient.post<SSOConsentResponse>(API_ENDPOINTS.AUTH_SSO_CONSENT, data);
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
