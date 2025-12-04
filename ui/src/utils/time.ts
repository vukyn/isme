/**
 * Format timestamp to relative time (e.g., "5 minutes ago")
 * This is a simple implementation - in production, you might want to use a library like date-fns
 */
export const formatRelativeTime = (timestamp: string | Date): string => {
	// If already formatted as relative time string, return as is
	if (typeof timestamp === "string" && timestamp.includes("ago")) {
		return timestamp;
	}

	const date = typeof timestamp === "string" ? new Date(timestamp) : timestamp;
	const now = new Date();
	const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

	if (diffInSeconds < 60) {
		return "just now";
	}

	const diffInMinutes = Math.floor(diffInSeconds / 60);
	if (diffInMinutes < 60) {
		return `${diffInMinutes} minute${diffInMinutes > 1 ? "s" : ""} ago`;
	}

	const diffInHours = Math.floor(diffInMinutes / 60);
	if (diffInHours < 24) {
		return `${diffInHours} hour${diffInHours > 1 ? "s" : ""} ago`;
	}

	const diffInDays = Math.floor(diffInHours / 24);
	if (diffInDays < 7) {
		return `${diffInDays} day${diffInDays > 1 ? "s" : ""} ago`;
	}

	const diffInWeeks = Math.floor(diffInDays / 7);
	if (diffInWeeks < 4) {
		return `${diffInWeeks} week${diffInWeeks > 1 ? "s" : ""} ago`;
	}

	const diffInMonths = Math.floor(diffInDays / 30);
	return `${diffInMonths} month${diffInMonths > 1 ? "s" : ""} ago`;
};

