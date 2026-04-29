import { createSystem, defaultConfig, defineConfig } from "@chakra-ui/react";

const config = defineConfig({
	theme: {
		tokens: {
			colors: {
				canvas: {
					0: { value: "#07071A" },
					1: { value: "#0B0B23" },
					2: { value: "#12122E" },
					3: { value: "#04040E" },
				},
				aurora: {
					cyan: { value: "#22D3EE" },
					blue: { value: "#6366F1" },
					violet: { value: "#8B5CF6" },
					magenta: { value: "#EC4899" },
					amber: { value: "#F59E0B" },
					mint: { value: "#34D399" },
				},
				ink: {
					DEFAULT: { value: "#F4F5FF" },
					soft: { value: "#C7CAE8" },
					mute: { value: "#8A8FB5" },
				},
				glass: {
					fill: { value: "rgba(255,255,255,0.06)" },
					fillHi: { value: "rgba(255,255,255,0.10)" },
					line: { value: "rgba(255,255,255,0.10)" },
					lineHi: { value: "rgba(255,255,255,0.18)" },
				},
				glow: {
					violet: { value: "rgba(139,92,246,0.45)" },
					violetSoft: { value: "rgba(139,92,246,0.25)" },
					blue: { value: "rgba(99,102,241,0.45)" },
					magenta: { value: "rgba(236,72,153,0.45)" },
				},
			},
			gradients: {
				auroraPrimary: {
					value: "linear-gradient(135deg, {colors.aurora.blue} 0%, {colors.aurora.violet} 50%, {colors.aurora.magenta} 100%)",
				},
				auroraText: {
					value: "linear-gradient(135deg, {colors.aurora.cyan}, {colors.aurora.blue} 40%, {colors.aurora.violet} 70%, {colors.aurora.magenta})",
				},
				auroraStat: {
					value: "linear-gradient(135deg, {colors.aurora.cyan}, {colors.aurora.violet})",
				},
				conicLogo: {
					value: "conic-gradient(from 180deg at 50% 50%, {colors.aurora.cyan}, {colors.aurora.blue}, {colors.aurora.violet}, {colors.aurora.magenta}, {colors.aurora.cyan})",
				},
			},
			shadows: {
				glassSoft: {
					value: "0 1px 0 rgba(255,255,255,0.06) inset, 0 8px 32px rgba(0,0,0,0.45)",
				},
				ctaGlow: {
					value: "0 10px 30px {colors.glow.blue}, 0 0 0 1px rgba(255,255,255,0.10) inset",
				},
				ctaGlowHi: { value: "0 14px 40px {colors.glow.violet}" },
				focusRing: {
					value: "0 0 0 4px {colors.glow.violetSoft}, 0 0 24px {colors.glow.violetSoft}",
				},
			},
			radii: {
				glass: { value: "16px" },
				glassSm: { value: "12px" },
			},
			durations: { ui: { value: "180ms" } },
			easings: { ui: { value: "cubic-bezier(.2,.8,.2,1)" } },
		},
		semanticTokens: {
			colors: {
				bg: { value: "{colors.canvas.0}" },
				"bg.subtle": { value: "{colors.canvas.1}" },
				"bg.muted": { value: "{colors.canvas.2}" },
				"bg.glass": { value: "{colors.glass.fill}" },
				"bg.glassHi": { value: "{colors.glass.fillHi}" },
				fg: { value: "{colors.ink.DEFAULT}" },
				"fg.muted": { value: "{colors.ink.mute}" },
				"fg.subtle": { value: "{colors.ink.soft}" },
				border: { value: "{colors.glass.line}" },
				"border.strong": { value: "{colors.glass.lineHi}" },
				success: { value: "{colors.aurora.mint}" },
				warning: { value: "{colors.aurora.amber}" },
				accent: { value: "{colors.aurora.violet}" },
				accentAlt: { value: "{colors.aurora.cyan}" },
				danger: { value: "#F87171" },
			},
		},
	},
});

export const system = createSystem(defaultConfig, config);
