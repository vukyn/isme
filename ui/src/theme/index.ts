import { createSystem, defaultConfig, defineConfig } from "@chakra-ui/react";

const config = defineConfig({
	theme: {
		tokens: {
			colors: {
				// Brand colors - Purple palette for accents
				brand: {
					50: { value: "#faf5ff" },
					100: { value: "#f3e8ff" },
					200: { value: "#e9d5ff" },
					300: { value: "#d8b4fe" },
					400: { value: "#c4b5fd" },
					500: { value: "#a855f7" },
					600: { value: "#9333ea" },
					700: { value: "#7c3aed" },
					800: { value: "#6b46c1" },
					900: { value: "#581c87" },
					950: { value: "#3b0764" },
					1000: { value: "#1e002b" },
				},
				// Custom gray palette for dark theme
				darkGray: {
					50: { value: "#f8fafc" },
					100: { value: "#f1f5f9" },
					200: { value: "#e2e8f0" },
					300: { value: "#cbd5e1" },
					400: { value: "#94a3b8" },
					500: { value: "#64748b" },
					600: { value: "#475569" },
					700: { value: "#334155" },
					800: { value: "#1e293b" },
					900: { value: "#0f172a" },
					950: { value: "#020617" },
				},
				// Surface colors for different elevations
				surface: {
					dark: {
						primary: { value: "#0f0f0f" },
						secondary: { value: "#1a1a1a" },
						elevated: { value: "#2d2d2d" },
						hover: { value: "#404040" },
					},
					light: {
						primary: { value: "#ffffff" },
						secondary: { value: "#f8fafc" },
						elevated: { value: "#f1f5f9" },
						hover: { value: "#e2e8f0" },
					},
				},
			},
			gradients: {
				brandPrimary: {
					value: "linear-gradient(135deg, {colors.brand.600}, {colors.brand.500})",
				},
				brandSecondary: {
					value: "linear-gradient(135deg, {colors.brand.700}, {colors.brand.600})",
				},
				brandSubtle: {
					value: "linear-gradient(135deg, {colors.brand.200}, {colors.brand.300})",
				},
			},
		},
		semanticTokens: {
			colors: {
				// Background semantic tokens
				bg: {
					value: {
						_light: "{colors.surface.light.primary}",
						_dark: "{colors.surface.dark.primary}",
					},
				},
				"bg.subtle": {
					value: {
						_light: "{colors.surface.light.secondary}",
						_dark: "{colors.surface.dark.secondary}",
					},
				},
				"bg.muted": {
					value: {
						_light: "{colors.surface.light.elevated}",
						_dark: "{colors.surface.dark.elevated}",
					},
				},
				"bg.emphasized": {
					value: {
						_light: "{colors.surface.light.hover}",
						_dark: "{colors.surface.dark.hover}",
					},
				},
				// Foreground semantic tokens
				fg: {
					value: {
						_light: "{colors.darkGray.900}",
						_dark: "{colors.white}",
					},
				},
				"fg.muted": {
					value: {
						_light: "{colors.darkGray.500}",
						_dark: "{colors.darkGray.400}",
					},
				},
				"fg.subtle": {
					value: {
						_light: "{colors.darkGray.600}",
						_dark: "{colors.darkGray.300}",
					},
				},
				// Border semantic tokens
				border: {
					value: {
						_light: "{colors.darkGray.200}",
						_dark: "{colors.darkGray.700}",
					},
				},
				"border.muted": {
					value: {
						_light: "{colors.darkGray.100}",
						_dark: "{colors.darkGray.800}",
					},
				},
				"border.emphasized": {
					value: {
						_light: "{colors.darkGray.300}",
						_dark: "{colors.darkGray.600}",
					},
				},
				// Brand semantic tokens
				brand: {
					solid: {
						value: {
							_light: "{colors.brand.500}",
							_dark: "{colors.surface.dark.elevated}",
						},
					},
					contrast: {
						value: {
							_light: "{colors.white}",
							_dark: "{colors.white}",
						},
					},
					fg: {
						value: {
							_light: "{colors.brand.600}",
							_dark: "{colors.brand.300}",
						},
					},
					muted: {
						value: {
							_light: "{colors.brand.100}",
							_dark: "{colors.brand.900}",
						},
					},
					subtle: {
						value: {
							_light: "{colors.brand.50}",
							_dark: "{colors.brand.950}",
						},
					},
					emphasized: {
						value: {
							_light: "{colors.brand.600}",
							_dark: "{colors.brand.400}",
						},
					},
					focusRing: {
						value: {
							_light: "{colors.brand.500}",
							_dark: "{colors.brand.400}",
						},
					},
				},
				// Button semantic tokens for dark theme consistency
				button: {
					solid: {
						value: {
							_light: "{colors.brand.500}",
							_dark: "{colors.brand.600}",
						},
					},
					contrast: {
						value: {
							_light: "{colors.white}",
							_dark: "{colors.white}",
						},
					},
					outline: {
						value: {
							_light: "{colors.brand.500}",
							_dark: "{colors.darkGray.600}",
						},
					},
					outlineText: {
						value: {
							_light: "{colors.white}",
							_dark: "{colors.white}",
						},
					},
					outlineHover: {
						value: {
							_light: "{colors.brand.600}",
							_dark: "{colors.darkGray.500}",
						},
					},
					// Search button (purple theme)
					search: {
						bg: {
							value: {
								_light: "{colors.purple.500/20}",
								_dark: "{colors.purple.500/20}",
							},
						},
						color: {
							value: {
								_light: "{colors.purple.600}",
								_dark: "{colors.purple.300}",
							},
						},
						border: {
							value: {
								_light: "{colors.purple.500/30}",
								_dark: "{colors.purple.500/30}",
							},
						},
						hoverBg: {
							value: {
								_light: "{colors.purple.500/30}",
								_dark: "{colors.purple.500/30}",
							},
						},
						hoverColor: {
							value: {
								_light: "{colors.purple.200}",
								_dark: "{colors.purple.200}",
							},
						},
						hoverBorder: {
							value: {
								_light: "{colors.purple.500/50}",
								_dark: "{colors.purple.500/50}",
							},
						},
					},
					// View button (purple theme)
					view: {
						bg: {
							value: {
								_light: "{colors.purple.500/20}",
								_dark: "{colors.purple.500/20}",
							},
						},
						color: {
							value: {
								_light: "{colors.purple.600}",
								_dark: "{colors.purple.300}",
							},
						},
						border: {
							value: {
								_light: "{colors.purple.500/30}",
								_dark: "{colors.purple.500/30}",
							},
						},
						hoverBg: {
							value: {
								_light: "{colors.purple.500/30}",
								_dark: "{colors.purple.500/30}",
							},
						},
						hoverColor: {
							value: {
								_light: "{colors.purple.200}",
								_dark: "{colors.purple.200}",
							},
						},
						hoverBorder: {
							value: {
								_light: "{colors.purple.500/50}",
								_dark: "{colors.purple.500/50}",
							},
						},
					},
					// Run button (blue theme)
					run: {
						bg: {
							value: {
								_light: "{colors.blue.600/20}",
								_dark: "{colors.blue.500/20}",
							},
						},
						color: {
							value: {
								_light: "{colors.blue.400}",
								_dark: "{colors.blue.300}",
							},
						},
						border: {
							value: {
								_light: "{colors.blue.600/30}",
								_dark: "{colors.blue.500/30}",
							},
						},
						hoverBg: {
							value: {
								_light: "{colors.blue.600/30}",
								_dark: "{colors.blue.500/30}",
							},
						},
						hoverColor: {
							value: {
								_light: "{colors.blue.500}",
								_dark: "{colors.blue.200}",
							},
						},
						hoverBorder: {
							value: {
								_light: "{colors.blue.600/50}",
								_dark: "{colors.blue.500/50}",
							},
						},
					},
					// Stop button (red theme)
					stop: {
						bg: {
							value: {
								_light: "{colors.red.600/20}",
								_dark: "{colors.red.500/20}",
							},
						},
						color: {
							value: {
								_light: "{colors.red.400}",
								_dark: "{colors.red.300}",
							},
						},
						border: {
							value: {
								_light: "{colors.red.600/30}",
								_dark: "{colors.red.500/30}",
							},
						},
						hoverBg: {
							value: {
								_light: "{colors.red.600/30}",
								_dark: "{colors.red.500/30}",
							},
						},
						hoverColor: {
							value: {
								_light: "{colors.red.500}",
								_dark: "{colors.red.200}",
							},
						},
						hoverBorder: {
							value: {
								_light: "{colors.red.600/50}",
								_dark: "{colors.red.500/50}",
							},
						},
					},
					// Legacy tokens for backward compatibility
					success: {
						value: {
							_light: "{colors.green.500}",
							_dark: "{colors.green.600}",
						},
					},
					successHover: {
						value: {
							_light: "{colors.green.600}",
							_dark: "{colors.green.500}",
						},
					},
					error: {
						value: {
							_light: "{colors.red.500}",
							_dark: "{colors.red.600}",
						},
					},
					errorHover: {
						value: {
							_light: "{colors.red.600}",
							_dark: "{colors.red.500}",
						},
					},
				},
				// Gray semantic tokens for neutral buttons
				gray: {
					solid: {
						value: {
							_light: "{colors.darkGray.800}",
							_dark: "{colors.darkGray.700}",
						},
					},
					contrast: {
						value: {
							_light: "{colors.white}",
							_dark: "{colors.white}",
						},
					},
					fg: {
						value: {
							_light: "{colors.darkGray.600}",
							_dark: "{colors.darkGray.300}",
						},
					},
					muted: {
						value: {
							_light: "{colors.darkGray.100}",
							_dark: "{colors.darkGray.800}",
						},
					},
					subtle: {
						value: {
							_light: "{colors.darkGray.50}",
							_dark: "{colors.darkGray.900}",
						},
					},
					emphasized: {
						value: {
							_light: "{colors.darkGray.700}",
							_dark: "{colors.darkGray.600}",
						},
					},
					focusRing: {
						value: {
							_light: "{colors.darkGray.500}",
							_dark: "{colors.darkGray.400}",
						},
					},
				},
				// Status colors
				error: {
					solid: {
						value: {
							_light: "{colors.red.500}",
							_dark: "{colors.red.400}",
						},
					},
					muted: {
						value: {
							_light: "{colors.red.50}",
							_dark: "{colors.red.950}",
						},
					},
					fg: {
						value: {
							_light: "{colors.red.600}",
							_dark: "{colors.red.300}",
						},
					},
				},
				success: {
					solid: {
						value: {
							_light: "{colors.green.500}",
							_dark: "{colors.green.400}",
						},
					},
					muted: {
						value: {
							_light: "{colors.green.50}",
							_dark: "{colors.green.950}",
						},
					},
					fg: {
						value: {
							_light: "{colors.green.600}",
							_dark: "{colors.green.300}",
						},
					},
				},
				warning: {
					solid: {
						value: {
							_light: "{colors.yellow.500}",
							_dark: "{colors.yellow.400}",
						},
					},
					muted: {
						value: {
							_light: "{colors.yellow.50}",
							_dark: "{colors.yellow.950}",
						},
					},
					fg: {
						value: {
							_light: "{colors.yellow.600}",
							_dark: "{colors.yellow.300}",
						},
					},
				},
				info: {
					solid: {
						value: {
							_light: "{colors.blue.500}",
							_dark: "{colors.blue.400}",
						},
					},
					muted: {
						value: {
							_light: "{colors.blue.50}",
							_dark: "{colors.blue.950}",
						},
					},
					fg: {
						value: {
							_light: "{colors.blue.600}",
							_dark: "{colors.blue.300}",
						},
					},
				},
				// Slider semantic tokens with high contrast
				slider: {
					track: {
						value: {
							_light: "{colors.darkGray.200}",
							_dark: "{colors.darkGray.700}",
						},
					},
					range: {
						value: {
							_light: "{colors.brand.500}",
							_dark: "{colors.brand.500}",
						},
					},
					thumb: {
						value: {
							_light: "{colors.brand.600}",
							_dark: "{colors.brand.500}",
						},
					},
					thumbBg: {
						value: {
							_light: "{colors.white}",
							_dark: "{colors.white}",
						},
					},
				},
			},
			gradients: {
				brandPrimary: {
					value: {
						_light: "{gradients.brandPrimary}",
						_dark: "{gradients.brandSecondary}",
					},
				},
				brandHeader: {
					value: {
						_light: "{gradients.brandSubtle}",
						_dark: "{gradients.brandPrimary}",
					},
				},
			},
		},
	},
});

export const system = createSystem(defaultConfig, config);
