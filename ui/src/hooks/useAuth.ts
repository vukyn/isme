import { useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { login as loginApi, signup as signupApi, getCurrentUser, logout as logoutApi } from "@/apis/auth";
import { saveTokens, clearTokens, getTokens } from "@/utils/axios";
import type { LoginRequest, SignupRequest } from "@/types";

interface UseAuthReturn {
	login: (data: LoginRequest) => Promise<void>;
	signup: (data: SignupRequest) => Promise<void>;
	logout: () => Promise<void>;
	refreshToken: () => Promise<void>;
	getCurrentUser: () => Promise<unknown>;
	isAuthenticated: boolean;
	loading: boolean;
	error: Error | null;
}

export const useAuth = (): UseAuthReturn => {
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<Error | null>(null);
	const navigate = useNavigate();

	const isAuthenticated = !!getTokens().access_token;

	const login = useCallback(
		async (data: LoginRequest) => {
			try {
				setLoading(true);
				setError(null);
				const response = await loginApi(data);
				saveTokens(response.data);
				navigate("/welcome");
			} catch (err: any) {
				const errorMessage = err?.response?.data?.message || err?.message || "Login failed. Please check your credentials.";
				const error = new Error(errorMessage);
				setError(error);
				throw error;
			} finally {
				setLoading(false);
			}
		},
		[navigate]
	);

	const signup = useCallback(
		async (data: SignupRequest) => {
			try {
				setLoading(true);
				setError(null);
				await signupApi(data);
				navigate("/login");
			} catch (err: any) {
				const errorMessage = err?.response?.data?.message || err?.message || "Signup failed. Please try again.";
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
		} catch (err: any) {
			// Even if API call fails, we should still clear tokens and log out locally
			const errorMessage = err?.response?.data?.message || err?.message || "Logout failed";
			setError(new Error(errorMessage));
		} finally {
			clearTokens();
			navigate("/login");
			setLoading(false);
		}
	}, [navigate]);

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

	const getCurrentUserData = useCallback(async () => {
		try {
			setLoading(true);
			setError(null);
			return await getCurrentUser();
		} catch (err) {
			const error = err instanceof Error ? err : new Error("Failed to get user data");
			setError(error);
			throw error;
		} finally {
			setLoading(false);
		}
	}, []);

	return {
		login,
		signup,
		logout,
		refreshToken,
		getCurrentUser: getCurrentUserData,
		isAuthenticated,
		loading,
		error,
	};
};
