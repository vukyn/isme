import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { login as loginApi, signup as signupApi, logout as logoutApi } from "@/apis/auth";
import { saveTokens, clearTokens, getTokens } from "@/utils/axios";
import { useUser } from "@/hooks/useUser";
import type { LoginRequest, LoginResponse, SignupRequest } from "@/types";

interface ApiErrorLike {
	response?: { data?: { message?: string } };
	message?: string;
}

const getApiErrorMessage = (err: unknown, fallback: string): string => {
	const apiError = err as ApiErrorLike | null | undefined;
	return apiError?.response?.data?.message || apiError?.message || fallback;
};

interface UseAuthReturn {
	login: (data: LoginRequest) => Promise<LoginResponse>;
	signup: (data: SignupRequest) => Promise<void>;
	logout: () => Promise<void>;
	refreshToken: () => Promise<void>;
	isAuthenticated: boolean;
	loading: boolean;
	error: Error | null;
}

export const useAuth = (): UseAuthReturn => {
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<Error | null>(null);
	const navigate = useNavigate();
	const { refetch: refetchUser, clear: clearUser } = useUser();

	const isAuthenticated = !!getTokens().access_token;

	const login = useCallback(
		async (data: LoginRequest): Promise<LoginResponse> => {
			try {
				setLoading(true);
				setError(null);
				const response = await loginApi(data);
				saveTokens(response.data);
				// Only navigate if this is not an SSO login (no redirect_url);
				// SSO logins leave the app, so skip the /me prefetch.
				if (!response.data.redirect_url) {
					await refetchUser();
					navigate("/welcome");
				}
				return response;
			} catch (err) {
				const errorMessage = getApiErrorMessage(err, "Login failed. Please check your credentials.");
				const error = new Error(errorMessage);
				setError(error);
				throw error;
			} finally {
				setLoading(false);
			}
		},
		[navigate, refetchUser]
	);

	const signup = useCallback(
		async (data: SignupRequest) => {
			try {
				setLoading(true);
				setError(null);
				await signupApi(data);
				navigate("/login");
			} catch (err) {
				const errorMessage = getApiErrorMessage(err, "Signup failed. Please try again.");
				const error = new Error(errorMessage);
				setError(error);
				throw error;
			} finally {
				setLoading(false);
			}
		},
		[navigate]
	);

	const logout = useCallback(async () => {
		try {
			setLoading(true);
			setError(null);
			await logoutApi();
		} catch (err) {
			// Even if API call fails, we should still clear tokens and log out locally
			const errorMessage = getApiErrorMessage(err, "Logout failed");
			setError(new Error(errorMessage));
		} finally {
			clearTokens();
			clearUser();
			navigate("/login");
			setLoading(false);
		}
	}, [navigate, clearUser]);

	const refreshToken = useCallback(async () => {
		try {
			const tokens = getTokens();
			if (!tokens.refresh_token) {
				throw new Error("No refresh token available");
			}
			// This is handled by axios interceptor, but we can expose it if needed
		} catch (err) {
			const error = err instanceof Error ? err : new Error("Token refresh failed");
			setError(error);
			throw error;
		}
	}, []);

	return {
		login,
		signup,
		logout,
		refreshToken,
		isAuthenticated,
		loading,
		error,
	};
};
