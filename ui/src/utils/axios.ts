import axios, { type AxiosError, type InternalAxiosRequestConfig } from "axios";
import { API_ENDPOINTS } from "@/consts";
import { getCookie, setCookie, deleteCookie } from "./cookies";

const STORAGE_KEYS = {
	ACCESS_TOKEN: "access_token",
	REFRESH_TOKEN: "refresh_token",
	EXPIRES_AT: "expires_at",
} as const;

/**
 * Get tokens from cookies
 */
export const getTokens = () => {
	return {
		access_token: getCookie(STORAGE_KEYS.ACCESS_TOKEN),
		refresh_token: getCookie(STORAGE_KEYS.REFRESH_TOKEN),
		expires_at: getCookie(STORAGE_KEYS.EXPIRES_AT),
	};
};

/**
 * Save tokens to cookies (session cookies)
 */
export const saveTokens = (tokens: { access_token: string; refresh_token: string; expires_at: string }) => {
	setCookie(STORAGE_KEYS.ACCESS_TOKEN, tokens.access_token);
	setCookie(STORAGE_KEYS.REFRESH_TOKEN, tokens.refresh_token);
	setCookie(STORAGE_KEYS.EXPIRES_AT, tokens.expires_at);
};

/**
 * Clear tokens from cookies
 */
export const clearTokens = () => {
	deleteCookie(STORAGE_KEYS.ACCESS_TOKEN);
	deleteCookie(STORAGE_KEYS.REFRESH_TOKEN);
	deleteCookie(STORAGE_KEYS.EXPIRES_AT);
};

/**
 * Check if token is expired
 */
export const isTokenExpired = (): boolean => {
	const expiresAt = getCookie(STORAGE_KEYS.EXPIRES_AT);
	if (!expiresAt) return true;
	return new Date(expiresAt) < new Date();
};

/**
 * Create axios instance
 */
export const apiClient = axios.create({
	baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080",
	headers: {
		"Content-Type": "application/json",
	},
});

let isRefreshing = false;
let failedQueue: Array<{
	resolve: (value?: unknown) => void;
	reject: (error?: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null, token: string | null = null) => {
	failedQueue.forEach((prom) => {
		if (error) {
			prom.reject(error);
		} else {
			prom.resolve(token);
		}
	});
	failedQueue = [];
};

/**
 * Request interceptor - Add access token to headers
 */
apiClient.interceptors.request.use(
	(config: InternalAxiosRequestConfig) => {
		const tokens = getTokens();
		if (tokens.access_token && config.headers) {
			config.headers.Authorization = `Bearer ${tokens.access_token}`;
		}
		return config;
	},
	(error) => {
		return Promise.reject(error);
	}
);

/**
 * Response interceptor - Handle token refresh on 401
 */
apiClient.interceptors.response.use(
	(response) => response,
	async (error: AxiosError) => {
		const originalRequest = error.config as InternalAxiosRequestConfig & {
			_retry?: boolean;
		};

		// If error is 401 and we haven't retried yet
		if (error.response?.status === 401 && !originalRequest._retry) {
			if (isRefreshing) {
				// If already refreshing, queue this request
				return new Promise((resolve, reject) => {
					failedQueue.push({ resolve, reject });
				})
					.then((token) => {
						if (originalRequest.headers) {
							originalRequest.headers.Authorization = `Bearer ${token}`;
						}
						return apiClient(originalRequest);
					})
					.catch((err) => {
						return Promise.reject(err);
					});
			}

			originalRequest._retry = true;
			isRefreshing = true;

			const tokens = getTokens();
			if (!tokens.refresh_token) {
				clearTokens();
				processQueue(error, null);
				isRefreshing = false;
				window.location.href = "/login";
				return Promise.reject(error);
			}

			try {
				const response = await axios.post(
					`${import.meta.env.VITE_API_BASE_URL || "http://localhost:8080"}${API_ENDPOINTS.AUTH_REFRESH}`,
					{
						refresh_token: tokens.refresh_token,
					}
				);

				const { access_token, refresh_token, expires_at } = response.data.data;
				saveTokens({ access_token, refresh_token, expires_at });

				if (originalRequest.headers) {
					originalRequest.headers.Authorization = `Bearer ${access_token}`;
				}

				processQueue(null, access_token);
				isRefreshing = false;

				return apiClient(originalRequest);
			} catch (refreshError) {
				processQueue(refreshError as AxiosError, null);
				isRefreshing = false;
				clearTokens();
				window.location.href = "/login";
				return Promise.reject(refreshError);
			}
		}

		return Promise.reject(error);
	}
);
