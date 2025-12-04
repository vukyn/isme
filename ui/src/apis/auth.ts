import type {
	LoginRequest,
	LoginResponse,
	SignupRequest,
	SignupResponse,
	RefreshTokenRequest,
	RefreshTokenResponse,
} from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

export const login = async (data: LoginRequest): Promise<LoginResponse> => {
	const response = await apiClient.post<LoginResponse>(
		API_ENDPOINTS.AUTH_LOGIN,
		data,
	);
	return response.data;
};

export const signup = async (data: SignupRequest): Promise<SignupResponse> => {
	const response = await apiClient.post<SignupResponse>(
		API_ENDPOINTS.AUTH_SIGNUP,
		data,
	);
	return response.data;
};

export const getCurrentUser = async () => {
	const response = await apiClient.get(API_ENDPOINTS.AUTH_ME);
	return response.data;
};

export const refreshToken = async (
	data: RefreshTokenRequest,
): Promise<RefreshTokenResponse> => {
	const response = await apiClient.post<RefreshTokenResponse>(
		API_ENDPOINTS.AUTH_REFRESH,
		data,
	);
	return response.data;
};

