/**
 * Cookie utilities for managing session cookies
 */

/**
 * Set a cookie
 * @param name - Cookie name
 * @param value - Cookie value
 * @param days - Number of days until expiration (optional, if not provided, creates session cookie)
 */
export const setCookie = (name: string, value: string, days?: number): void => {
	let expires = "";
	if (days) {
		const date = new Date();
		date.setTime(date.getTime() + days * 24 * 60 * 60 * 1000);
		expires = `; expires=${date.toUTCString()}`;
	}
	// Session cookie if no days provided
	document.cookie = `${name}=${value || ""}${expires}; path=/; SameSite=Lax`;
};

/**
 * Get a cookie value by name
 * @param name - Cookie name
 * @returns Cookie value or null if not found
 */
export const getCookie = (name: string): string | null => {
	const nameEQ = `${name}=`;
	const ca = document.cookie.split(";");
	for (let i = 0; i < ca.length; i++) {
		let c = ca[i];
		while (c.charAt(0) === " ") c = c.substring(1, c.length);
		if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
	}
	return null;
};

/**
 * Delete a cookie
 * @param name - Cookie name
 */
export const deleteCookie = (name: string): void => {
	document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
};
