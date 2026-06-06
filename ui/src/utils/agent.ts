/**
 * Best-effort user-agent → "Browser · OS" label (e.g. "Chrome · macOS").
 * Falls back to the raw string when nothing is recognized.
 */
export const parseUserAgent = (userAgent: string): string => {
	if (!userAgent) return "Unknown device";

	let browser = "";
	if (/edg(a|ios)?\//i.test(userAgent)) browser = "Edge";
	else if (/opr\/|opera/i.test(userAgent)) browser = "Opera";
	else if (/firefox|fxios/i.test(userAgent)) browser = "Firefox";
	else if (/chrome|crios/i.test(userAgent)) browser = "Chrome";
	else if (/safari/i.test(userAgent)) browser = "Safari";

	let os = "";
	if (/windows/i.test(userAgent)) os = "Windows";
	else if (/iphone|ipad|ipod/i.test(userAgent)) os = "iOS";
	else if (/mac os x|macintosh/i.test(userAgent)) os = "macOS";
	else if (/android/i.test(userAgent)) os = "Android";
	else if (/linux/i.test(userAgent)) os = "Linux";

	if (!browser && !os) return userAgent;
	return `${browser || "Unknown"} · ${os || "Unknown"}`;
};
