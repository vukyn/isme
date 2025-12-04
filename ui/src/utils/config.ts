/**
 * Runtime configuration interface
 * This is injected by the Go server at runtime via window.__APP_CONFIG__
 */
interface RuntimeConfig {
	API_BASE_URL?: string;
}

declare global {
	interface Window {
		__APP_CONFIG__?: RuntimeConfig;
	}
}

/**
 * Get API base URL from runtime config or fall back to build-time env variable
 * Runtime config takes precedence over build-time env variables
 */
export function getApiBaseUrl(): string {
	if (typeof window !== "undefined" && window.__APP_CONFIG__?.API_BASE_URL) {
		if (window.__APP_CONFIG__?.API_BASE_URL !== "{{.APIBaseURL}}") {
			// Try to get from runtime config (injected by Go server)
			return window.__APP_CONFIG__.API_BASE_URL;
		}
	}

	// Fall back to build-time environment variable
	return import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8081";
}
