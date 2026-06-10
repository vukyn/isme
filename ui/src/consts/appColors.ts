/**
 * Fixed aurora color palette for an app_service appearance.
 *
 * The backend stores a palette KEY string (see app_service constants
 * APP_SERVICE_COLOR_KEYS + IsValidColor) — never a free hex. This map resolves
 * each allowed key to its hex + rgb triplet so the UI can render tinted-glass
 * tiles. Hexes come straight from the aurora design tokens; rose/sky/teal are
 * aurora-consistent extension hues for palette breadth. Keep the key set in
 * sync with the backend allowlist.
 */

/** The allowlist of color palette keys — must match APP_SERVICE_COLOR_KEYS. */
export const APP_COLOR_KEYS = ["violet", "indigo", "cyan", "sky", "teal", "mint", "amber", "rose", "magenta"] as const;

export type AppColorKey = (typeof APP_COLOR_KEYS)[number];

interface AppColor {
	/** Full hex (e.g. "#8B5CF6"). */
	hex: string;
	/** Comma-separated rgb triplet (e.g. "139,92,246") for rgba() tints. */
	rgb: string;
}

/** key → hex + rgb. Pulled from the edit-design mock PALETTE. */
export const APP_COLORS: Record<AppColorKey, AppColor> = {
	violet: { hex: "#8B5CF6", rgb: "139,92,246" },
	indigo: { hex: "#6366F1", rgb: "99,102,241" },
	cyan: { hex: "#22D3EE", rgb: "34,211,238" },
	sky: { hex: "#38BDF8", rgb: "56,189,248" },
	teal: { hex: "#2DD4BF", rgb: "45,212,191" },
	mint: { hex: "#34D399", rgb: "52,211,153" },
	amber: { hex: "#F59E0B", rgb: "245,158,11" },
	rose: { hex: "#FB7185", rgb: "251,113,133" },
	magenta: { hex: "#EC4899", rgb: "236,72,153" },
};

/** The default palette key when a color is empty/unknown. */
export const DEFAULT_APP_COLOR_KEY: AppColorKey = "violet";

/**
 * Resolves a stored color key to its hex/rgb. Empty or unknown keys fall back
 * to the default (violet) so a row that has never been styled still renders.
 */
export const resolveAppColor = (key: string | undefined): AppColor => {
	if (key && key in APP_COLORS) {
		return APP_COLORS[key as AppColorKey];
	}
	return APP_COLORS[DEFAULT_APP_COLOR_KEY];
};
