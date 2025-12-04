/**
 * Helper function to build query string
 */
export const buildQueryString = (params: Record<string, string | number>): string => {
	const query = new URLSearchParams();
	Object.entries(params).forEach(([key, value]) => {
		query.append(key, String(value));
	});
	return query.toString();
};

/**
 * Helper function to build endpoint with query params
 */
export const buildEndpoint = (endpoint: string, params?: Record<string, string | number>): string => {
	if (!params || Object.keys(params).length === 0) {
		return endpoint;
	}
	return `${endpoint}?${buildQueryString(params)}`;
};
