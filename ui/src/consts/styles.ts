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
	// background shorthand resets repeat to `repeat`; with 200% size the
	// neighbouring tile's magenta (100% stop) seams onto the left edge as a
	// thin red sliver. Pin position + disable repeat to kill the seam.
	backgroundRepeat: "no-repeat",
	backgroundPosition: "0% 0%",
} as const;

const DIALOG_BUTTON_BASE = {
	h: "11",
	px: "4.5",
	borderRadius: "glassSm",
	fontSize: "sm",
	fontWeight: "semibold",
	variant: "outline",
} as const;

/** Mock .btn-ghost — Cancel / Close. */
export const GHOST_DIALOG_BUTTON_PROPS = {
	...DIALOG_BUTTON_BASE,
	borderColor: "border.strong",
	bg: "bg.glass",
	color: "fg",
	_hover: { bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" },
	_focusVisible: { outline: "none", boxShadow: "0 0 0 3px rgba(139,92,246,0.55)", borderColor: "aurora.violet" },
} as const;

/** Mock .btn-danger — the footer's destructive primary action. */
export const DANGER_DIALOG_BUTTON_PROPS = {
	...DIALOG_BUTTON_BASE,
	borderColor: "rgba(236,72,153,0.55)",
	bg: "rgba(236,72,153,0.16)",
	color: "#FFD3E8",
	_hover: { bg: "rgba(236,72,153,0.26)", borderColor: "rgba(236,72,153,0.75)" },
	_focusVisible: { outline: "none", boxShadow: "0 0 0 3px rgba(236,72,153,0.45)" },
} as const;
