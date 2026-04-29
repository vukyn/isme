/**
 * Shared aurora CTA gradient. Mirrors theme.gradients.auroraPrimary
 * but used directly via `css={{ background }}` because Chakra v3
 * `bgGradient="auroraPrimary"` token resolution for the full gradient
 * string is fragile across versions in this project.
 */
export const AURORA_CTA_BG = "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)";
export const AURORA_CTA_STYLE = {
	background: AURORA_CTA_BG,
	backgroundSize: "200% 200%",
} as const;
