# Aurora UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace existing purple/light Chakra UI v3 frontend with Aurora UI: dark-only theme, animated mesh-gradient background, glass cards, iridescent text accents. Add a real Welcome dashboard (mock data), forgot-password stub, and stubs for `/sessions`, `/team`, `/settings`. No backend or new APIs.

**Architecture:** Hybrid organization — shared `AuthLayout` + `AppShell` layouts plus reusable atoms (`AuroraBackground`, `BrandMark`, `BrandPanel`, `GlassCard`, `FeatureList`, `PasswordField`, `PasswordStrength`, `StatCard`, `ActivityRow`, `Topbar`, `UserChip`). Theme rewritten clean (`canvas`, `aurora`, `ink`, `glass`, `glow` tokens + `auroraPrimary`/`auroraText`/`auroraStat`/`conicLogo` gradients). `<AuroraBackground/>` mounted once at the App root.

**Tech Stack:** Vite 7 + React 19 + TypeScript 5.9 + Chakra UI v3.30 + `@emotion/react` + `next-themes` + `react-router-dom` v7 + `react-icons/lu` + `zod`.

**Verification model:** No test runner installed in `ui/`. Each task verified by `npm run build` + `npm run lint` and (when UI-visible) a manual smoke walk via `make run-ui`. Spec at `docs/superpowers/specs/2026-04-29-aurora-ui-redesign-design.md`.

**Working directory:** `/Users/vuky10/vukyn/repo/isme`. All UI commands assume `cd ui &&` prefix or run from `ui/`.

**Branch:** `feat/aurora-ui-redesign` (already created, contains spec + mockup commit `5045cf2`).

---

## File Map

| Path | Op | Owns |
|------|----|------|
| `ui/src/theme/index.ts` | rewrite | Aurora tokens, gradients, shadows, semantic tokens |
| `ui/src/components/ui/provider.tsx` | edit | Force dark mode via `next-themes` |
| `ui/src/components/ui/button.tsx` | edit | Drop legacy `brand` colorPalette mapping |
| `ui/src/index.css` | rewrite | Aurora `@keyframes`, reduced-motion guard, base bg/fg |
| `ui/src/App.css` | rewrite | Trim — remove logo-spin/h1 leftovers |
| `ui/src/App.tsx` | edit | Mount `<AuroraBackground/>`, add 4 new routes |
| `ui/src/components/ui/aurora-background.tsx` | new | Fixed mesh-gradient bg + animation |
| `ui/src/components/ui/brand-mark.tsx` | new | Conic-gradient logo |
| `ui/src/components/ui/glass-card.tsx` | new | Glass surface card |
| `ui/src/components/ui/brand-panel.tsx` | new | Auth left side composition |
| `ui/src/components/ui/feature-list.tsx` | new | Glass row list (icon/title/desc) |
| `ui/src/components/ui/password-field.tsx` | new | Password Field + eye toggle |
| `ui/src/components/ui/password-strength.tsx` | new | 4-bar strength meter |
| `ui/src/components/ui/topbar.tsx` | new | AppShell topbar (logo + nav + notif + UserChip) |
| `ui/src/components/ui/user-chip.tsx` | new | Avatar + meta + logout menu |
| `ui/src/components/ui/stat-card.tsx` | new | Dashboard stat tile |
| `ui/src/components/ui/activity-row.tsx` | new | Activity feed row |
| `ui/src/layouts/AuthLayout.tsx` | new | Split brand+form shell |
| `ui/src/layouts/AppShell.tsx` | new | Topbar + main shell |
| `ui/src/consts/mock.ts` | new | MOCK_STATS + MOCK_ACTIVITY |
| `ui/src/pages/Login.tsx` | rewrite | Aurora Login |
| `ui/src/pages/Signup.tsx` | rewrite | Aurora Signup, drop T&C |
| `ui/src/pages/Welcome.tsx` | rewrite | AppShell + dashboard |
| `ui/src/pages/ForgotPassword.tsx` | new | Stub form |
| `ui/src/pages/Sessions.tsx` | new | Stub page |
| `ui/src/pages/Team.tsx` | new | Stub page |
| `ui/src/pages/Settings.tsx` | new | Stub page |
| `ui/src/pages/SSOLogin.tsx` | edit | Token swap only, keep layout |
| `ui/src/pages/NotFound.tsx` | edit | Aurora retheme |
| `ui/src/pages/Home.tsx` | edit | Aurora retheme |
| `ui/src/components/ProtectedRoute.tsx` | edit | Spinner color → `accent` |

---

## Phase 1 — Theme + Foundation

### Task 1: Rewrite theme tokens

**Files:**
- Modify: `ui/src/theme/index.ts` (full rewrite)

- [ ] **Step 1:** Replace the entire file contents with the spec theme.

`ui/src/theme/index.ts`:
```ts
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
```

- [ ] **Step 2:** Run build to confirm no token resolution errors.

```bash
cd ui && npm run build
```

Expected: build passes (existing pages still reference legacy `brand.*` tokens — this is OK, Chakra falls back silently. They will be rewritten in Phase 3).

- [ ] **Step 3:** Commit.

```bash
git add ui/src/theme/index.ts
git commit -m "chore(ui): rewrite theme to aurora dark tokens

Drop brand/darkGray/surface palettes. Add canvas, aurora, ink,
glass, glow tokens plus auroraPrimary/Text/Stat/conicLogo
gradients and glassSoft/ctaGlow/focusRing shadow tokens.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: Force dark mode via provider

**Files:**
- Modify: `ui/src/components/ui/provider.tsx`

- [ ] **Step 1:** Replace `provider.tsx` to force dark.

`ui/src/components/ui/provider.tsx`:
```tsx
"use client";

import { ChakraProvider } from "@chakra-ui/react";
import { ColorModeProvider, type ColorModeProviderProps } from "./color-mode";
import { system } from "@/theme";

export function Provider(props: ColorModeProviderProps) {
	return (
		<ChakraProvider value={system}>
			<ColorModeProvider forcedTheme="dark" defaultTheme="dark" {...props} />
		</ChakraProvider>
	);
}
```

- [ ] **Step 2:** Build to verify TS still compiles.

```bash
cd ui && npm run build
```

Expected: build passes. `next-themes` accepts `forcedTheme` and `defaultTheme` on `ThemeProvider`.

- [ ] **Step 3:** Commit.

```bash
git add ui/src/components/ui/provider.tsx
git commit -m "chore(ui): force dark theme in Provider

Aurora UI is dark-only. forcedTheme=dark prevents user toggle
and disables next-themes light fallback.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: Update Button wrapper (drop legacy brand palette)

**Files:**
- Modify: `ui/src/components/ui/button.tsx`

- [ ] **Step 1:** Replace Button wrapper — pass-through only, no `colorPalette="brand"`.

`ui/src/components/ui/button.tsx`:
```tsx
"use client";

import { Button as ChakraButton, type ButtonProps } from "@chakra-ui/react";
import { forwardRef } from "react";

export interface ButtonComponentProps extends ButtonProps {}

export const Button = forwardRef<HTMLButtonElement, ButtonComponentProps>(function Button(props, ref) {
	return <ChakraButton ref={ref} {...props} />;
});
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** Commit.

```bash
git add ui/src/components/ui/button.tsx
git commit -m "chore(ui): simplify Button wrapper, drop brand palette mapping

Pages now drive variant/colorPalette directly. Legacy primary
shorthand removed alongside brand token deletion.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: Rewrite global CSS

**Files:**
- Modify: `ui/src/index.css`
- Modify: `ui/src/App.css`

- [ ] **Step 1:** Replace `index.css` with aurora keyframes + reduced-motion guard.

`ui/src/index.css`:
```css
:root {
	font-family: "Inter", "DM Sans", system-ui, -apple-system, sans-serif;
	line-height: 1.5;
	font-weight: 400;
	font-synthesis: none;
	text-rendering: optimizeLegibility;
	-webkit-font-smoothing: antialiased;
	-moz-osx-font-smoothing: grayscale;
	color-scheme: dark;
}

html,
body {
	margin: 0;
	padding: 0;
	min-width: 320px;
	min-height: 100vh;
	background: #07071a;
	color: #f4f5ff;
}

#root {
	width: 100%;
	min-height: 100vh;
}

@keyframes auroraFlow {
	0% {
		background-position: 0% 0%, 100% 0%, 100% 100%, 0% 100%;
	}
	50% {
		background-position: 30% 20%, 60% 30%, 70% 70%, 30% 60%;
	}
	100% {
		background-position: 100% 100%, 0% 100%, 0% 0%, 100% 0%;
	}
}

@media (prefers-reduced-motion: reduce) {
	*,
	*::before,
	*::after {
		animation: none !important;
		transition: none !important;
	}
}
```

- [ ] **Step 2:** Trim `App.css` — drop logo-spin and h1.

`ui/src/App.css`:
```css
#root {
	width: 100%;
	min-height: 100vh;
	margin: 0;
	padding: 0;
}
```

- [ ] **Step 3:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 4:** Commit.

```bash
git add ui/src/index.css ui/src/App.css
git commit -m "chore(ui): aurora base CSS + reduced-motion guard

index.css: Inter/DM Sans stack, dark color-scheme, auroraFlow
keyframes, prefers-reduced-motion override. App.css trimmed.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase 2 — Layouts + Atoms

### Task 5: AuroraBackground

**Files:**
- Create: `ui/src/components/ui/aurora-background.tsx`

- [ ] **Step 1:** Create the component.

`ui/src/components/ui/aurora-background.tsx`:
```tsx
import { Box } from "@chakra-ui/react";

export const AuroraBackground = () => {
	return (
		<>
			<Box
				position="fixed"
				inset="0"
				zIndex="-2"
				pointerEvents="none"
				css={{
					background: [
						"radial-gradient(60% 50% at 12% 10%, rgba(99,102,241,0.45), transparent 60%)",
						"radial-gradient(45% 40% at 88% 18%, rgba(34,211,238,0.35), transparent 60%)",
						"radial-gradient(55% 45% at 75% 92%, rgba(236,72,153,0.32), transparent 60%)",
						"radial-gradient(50% 40% at 18% 80%, rgba(139,92,246,0.32), transparent 60%)",
						"linear-gradient(180deg, #07071A 0%, #0B0B23 60%, #04040E 100%)",
					].join(","),
					backgroundSize: "200% 200%",
					animation: "auroraFlow 14s ease-in-out infinite alternate",
					filter: "saturate(1.15)",
				}}
			/>
			<Box
				position="fixed"
				inset="0"
				zIndex="-1"
				pointerEvents="none"
				css={{
					backgroundImage:
						"radial-gradient(rgba(255,255,255,0.03) 1px, transparent 1px)",
					backgroundSize: "3px 3px",
					mixBlendMode: "overlay",
					opacity: 0.45,
				}}
			/>
		</>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet — batched at end of Phase 2.

---

### Task 6: BrandMark

**Files:**
- Create: `ui/src/components/ui/brand-mark.tsx`

- [ ] **Step 1:** Create.

`ui/src/components/ui/brand-mark.tsx`:
```tsx
import { Box, Center } from "@chakra-ui/react";

interface BrandMarkProps {
	size?: "sm" | "md";
}

export const BrandMark = ({ size = "md" }: BrandMarkProps) => {
	const outer = size === "sm" ? "8" : "10";
	const inner = size === "sm" ? "6" : "8";
	const radii = size === "sm" ? "9px" : "10px";
	const innerRadii = size === "sm" ? "7px" : "9px";
	const iconSize = size === "sm" ? 14 : 16;

	return (
		<Center
			w={outer}
			h={outer}
			borderRadius={radii}
			boxShadow="0 8px 24px rgba(99,102,241,0.45)"
			css={{
				background:
					"conic-gradient(from 180deg at 50% 50%, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
			}}
		>
			<Center w={inner} h={inner} borderRadius={innerRadii} bg="canvas.0">
				<svg
					width={iconSize}
					height={iconSize}
					viewBox="0 0 24 24"
					fill="none"
					stroke="white"
					strokeWidth={2.4}
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M12 2 4 7l8 5 8-5-8-5Z" />
					<path d="m4 17 8 5 8-5" />
					<path d="m4 12 8 5 8-5" />
				</svg>
			</Center>
		</Center>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 7: GlassCard

**Files:**
- Create: `ui/src/components/ui/glass-card.tsx`

- [ ] **Step 1:** Create.

`ui/src/components/ui/glass-card.tsx`:
```tsx
import { Box, type BoxProps } from "@chakra-ui/react";
import { forwardRef } from "react";

export interface GlassCardProps extends BoxProps {
	tone?: "default" | "strong";
}

export const GlassCard = forwardRef<HTMLDivElement, GlassCardProps>(function GlassCard(
	{ tone = "default", children, ...rest },
	ref
) {
	return (
		<Box
			ref={ref}
			bg={tone === "strong" ? "bg.glassHi" : "bg.glass"}
			borderWidth="1px"
			borderColor={tone === "strong" ? "border.strong" : "border"}
			borderRadius="2xl"
			boxShadow="glassSoft"
			css={{ backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" }}
			{...rest}
		>
			{children}
		</Box>
	);
});
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 8: FeatureList

**Files:**
- Create: `ui/src/components/ui/feature-list.tsx`

- [ ] **Step 1:** Create.

`ui/src/components/ui/feature-list.tsx`:
```tsx
import { HStack, Stack, Box, Text, Center } from "@chakra-ui/react";
import type { ReactNode } from "react";

export interface FeatureItem {
	icon: ReactNode;
	title: string;
	desc: string;
}

interface FeatureListProps {
	items: FeatureItem[];
}

export const FeatureList = ({ items }: FeatureListProps) => {
	return (
		<Stack gap="3" mt="7" position="relative" zIndex={2}>
			{items.map((item, i) => (
				<HStack
					key={i}
					gap="3.5"
					align="flex-start"
					p="3.5"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border"
					borderRadius="xl"
					css={{ backdropFilter: "blur(20px) saturate(1.2)", WebkitBackdropFilter: "blur(20px) saturate(1.2)" }}
				>
					<Center
						w="8"
						h="8"
						borderRadius="lg"
						borderWidth="1px"
						borderColor="border.strong"
						color="accentAlt"
						css={{
							background:
								"linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.25))",
						}}
						flex="0 0 auto"
					>
						{item.icon}
					</Center>
					<Box>
						<Text fontSize="sm" fontWeight="semibold" color="fg" mb="0.5">
							{item.title}
						</Text>
						<Text fontSize="xs" color="fg.muted">
							{item.desc}
						</Text>
					</Box>
				</HStack>
			))}
		</Stack>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 9: BrandPanel

**Files:**
- Create: `ui/src/components/ui/brand-panel.tsx`

- [ ] **Step 1:** Create.

`ui/src/components/ui/brand-panel.tsx`:
```tsx
import { Box, Flex, HStack, Text, Heading } from "@chakra-ui/react";
import type { ReactNode } from "react";
import { BrandMark } from "./brand-mark";
import { FeatureList, type FeatureItem } from "./feature-list";

interface BrandPanelProps {
	pill: string;
	pillTone?: "violet" | "cyan" | "mint";
	titleLead: string;
	titleGrad: string;
	sub: string;
	features: FeatureItem[];
	footerLeft?: string;
	footerRight?: string;
}

const pillToneColor: Record<NonNullable<BrandPanelProps["pillTone"]>, string> = {
	violet: "#FBBF24",
	cyan: "#22D3EE",
	mint: "#34D399",
};

export const BrandPanel = ({
	pill,
	pillTone = "violet",
	titleLead,
	titleGrad,
	sub,
	features,
	footerLeft = "© isme 2026",
	footerRight = "v1.0 · main",
}: BrandPanelProps) => {
	const dot = pillToneColor[pillTone];

	return (
		<Flex
			position="relative"
			direction="column"
			justify="space-between"
			p="11"
			borderRightWidth={{ base: 0, md: "1px" }}
			borderBottomWidth={{ base: "1px", md: 0 }}
			borderColor="border"
			minH={{ base: "320px", md: "auto" }}
			overflow="hidden"
		>
			{/* glass orbs */}
			<Box position="absolute" w="320px" h="320px" right="-80px" top="-80px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(40px)", background: "radial-gradient(circle, rgba(99,102,241,0.6), transparent 70%)" }} />
			<Box position="absolute" w="260px" h="260px" left="-60px" bottom="40px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(40px)", background: "radial-gradient(circle, rgba(236,72,153,0.55), transparent 70%)" }} />
			<Box position="absolute" w="220px" h="220px" right="80px" bottom="-60px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(40px)", background: "radial-gradient(circle, rgba(34,211,238,0.45), transparent 70%)" }} />

			<HStack gap="3" position="relative" zIndex={2}>
				<BrandMark />
				<Text fontWeight="bold" fontSize="lg" letterSpacing="-0.01em" color="fg">
					isme
				</Text>
			</HStack>

			<Box position="relative" zIndex={2} maxW="480px">
				<HStack
					display="inline-flex"
					gap="2"
					px="3"
					py="1.5"
					borderRadius="full"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border.strong"
					css={{ backdropFilter: "blur(20px)", WebkitBackdropFilter: "blur(20px)" }}
				>
					<Box w="1.5" h="1.5" borderRadius="full" css={{ background: dot, boxShadow: `0 0 10px ${dot}` }} />
					<Text fontSize="xs" fontWeight="medium" color="fg.subtle">
						{pill}
					</Text>
				</HStack>
				<Heading
					as="h1"
					mt="4.5"
					mb="3.5"
					fontSize={{ base: "30px", md: "44px" }}
					lineHeight="1.05"
					letterSpacing="-0.025em"
					color="fg"
				>
					{titleLead}{" "}
					<Text
						as="span"
						css={{
							backgroundImage:
								"linear-gradient(135deg, #22D3EE 0%, #6366F1 40%, #8B5CF6 70%, #EC4899 100%)",
							backgroundSize: "200% 100%",
							WebkitBackgroundClip: "text",
							backgroundClip: "text",
							color: "transparent",
							animation: "auroraFlow 8s ease-in-out infinite alternate",
						}}
					>
						{titleGrad}
					</Text>
				</Heading>
				<Text color="fg.muted" fontSize="md" maxW="440px">
					{sub}
				</Text>
				<FeatureList items={features} />
			</Box>

			<HStack justify="space-between" position="relative" zIndex={2} color="fg.muted" fontSize="xs">
				<Text>{footerLeft}</Text>
				<Text>{footerRight}</Text>
			</HStack>
		</Flex>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 10: PasswordField + PasswordStrength

**Files:**
- Create: `ui/src/components/ui/password-field.tsx`
- Create: `ui/src/components/ui/password-strength.tsx`

- [ ] **Step 1:** Create PasswordField.

`ui/src/components/ui/password-field.tsx`:
```tsx
import { Field, IconButton, Input, InputGroup } from "@chakra-ui/react";
import { useState } from "react";
import { LuEye, LuEyeOff, LuLock } from "react-icons/lu";

interface PasswordFieldProps {
	label: string;
	value: string;
	onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
	error?: string;
	autoComplete?: "current-password" | "new-password";
	name?: string;
	placeholder?: string;
}

export const PasswordField = ({
	label,
	value,
	onChange,
	error,
	autoComplete = "current-password",
	name = "password",
	placeholder = "Enter password",
}: PasswordFieldProps) => {
	const [show, setShow] = useState(false);

	return (
		<Field.Root invalid={!!error}>
			<Field.Label>{label}</Field.Label>
			<InputGroup
				startElement={<LuLock />}
				endElement={
					<IconButton
						aria-label={show ? "Hide password" : "Show password"}
						variant="ghost"
						size="xs"
						onClick={() => setShow((s) => !s)}
						color="fg.muted"
						_hover={{ color: "fg" }}
					>
						{show ? <LuEyeOff /> : <LuEye />}
					</IconButton>
				}
			>
				<Input
					name={name}
					type={show ? "text" : "password"}
					autoComplete={autoComplete}
					placeholder={placeholder}
					value={value}
					onChange={onChange}
				/>
			</InputGroup>
			{error && <Field.ErrorText>{error}</Field.ErrorText>}
		</Field.Root>
	);
};
```

- [ ] **Step 2:** Create PasswordStrength.

`ui/src/components/ui/password-strength.tsx`:
```tsx
import { Box, HStack } from "@chakra-ui/react";

interface PasswordStrengthProps {
	value: string;
}

const score = (pw: string): number => {
	let s = 0;
	if (pw.length >= 8) s++;
	if (/[A-Z]/.test(pw)) s++;
	if (/[0-9]/.test(pw)) s++;
	if (/[^A-Za-z0-9]/.test(pw)) s++;
	return s;
};

export const PasswordStrength = ({ value }: PasswordStrengthProps) => {
	const lit = score(value);

	return (
		<HStack gap="1.5" mt="2.5" aria-hidden="true">
			{[0, 1, 2, 3].map((i) => (
				<Box
					key={i}
					flex="1"
					h="1"
					borderRadius="full"
					bg={i < lit ? undefined : "bg.glassHi"}
					css={
						i < lit
							? {
									background: "linear-gradient(90deg, #22D3EE, #8B5CF6)",
									boxShadow: "0 0 8px rgba(139,92,246,0.5)",
							  }
							: undefined
					}
				/>
			))}
		</HStack>
	);
};
```

- [ ] **Step 3:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 4:** No commit yet.

---

### Task 11: Topbar + UserChip

**Files:**
- Create: `ui/src/components/ui/user-chip.tsx`
- Create: `ui/src/components/ui/topbar.tsx`

- [ ] **Step 1:** Create UserChip.

`ui/src/components/ui/user-chip.tsx`:
```tsx
import { Box, Center, HStack, Menu, Portal, Text } from "@chakra-ui/react";
import { LuChevronDown, LuLogOut } from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";

interface UserChipProps {
	name: string;
	email: string;
}

const initials = (name: string) =>
	name
		.split(" ")
		.filter(Boolean)
		.map((s) => s[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

export const UserChip = ({ name, email }: UserChipProps) => {
	const { logout } = useAuth();

	return (
		<Menu.Root>
			<Menu.Trigger asChild>
				<HStack
					as="button"
					gap="2.5"
					pl="1"
					pr="2.5"
					py="1"
					borderWidth="1px"
					borderColor="border.strong"
					borderRadius="full"
					bg="bg.glass"
					cursor="pointer"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
					_hover={{ bg: "bg.glassHi" }}
				>
					<Center
						w="8"
						h="8"
						borderRadius="full"
						color="white"
						fontWeight="bold"
						fontSize="xs"
						boxShadow="0 0 16px rgba(139,92,246,0.45)"
						css={{
							background:
								"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
						}}
					>
						{initials(name)}
					</Center>
					<Box textAlign="left" lineHeight="1.15">
						<Text fontSize="sm" fontWeight="semibold" color="fg">
							{name}
						</Text>
						<Text fontSize="xs" color="fg.muted">
							{email}
						</Text>
					</Box>
					<Box color="fg.muted" ml="1">
						<LuChevronDown size={14} />
					</Box>
				</HStack>
			</Menu.Trigger>
			<Portal>
				<Menu.Positioner>
					<Menu.Content
						bg="bg.subtle"
						borderWidth="1px"
						borderColor="border.strong"
						boxShadow="glassSoft"
						css={{ backdropFilter: "blur(20px)", WebkitBackdropFilter: "blur(20px)" }}
					>
						<Menu.Item value="logout" onClick={() => logout()}>
							<LuLogOut /> Log out
						</Menu.Item>
					</Menu.Content>
				</Menu.Positioner>
			</Portal>
		</Menu.Root>
	);
};
```

- [ ] **Step 2:** Create Topbar.

`ui/src/components/ui/topbar.tsx`:
```tsx
import { Box, Flex, HStack, IconButton } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { LuBell } from "react-icons/lu";
import { BrandMark } from "./brand-mark";
import { UserChip } from "./user-chip";

export type TopbarTab = "overview" | "sessions" | "team" | "settings";

interface TopbarProps {
	active: TopbarTab;
	user: { name: string; email: string };
}

const NAV: { key: TopbarTab; label: string; to: string }[] = [
	{ key: "overview", label: "Overview", to: "/welcome" },
	{ key: "sessions", label: "Sessions", to: "/sessions" },
	{ key: "team", label: "Team", to: "/team" },
	{ key: "settings", label: "Settings", to: "/settings" },
];

export const Topbar = ({ active, user }: TopbarProps) => {
	return (
		<Flex
			as="header"
			h="16"
			px="7"
			align="center"
			justify="space-between"
			borderBottomWidth="1px"
			borderColor="border"
			bg="rgba(7,7,26,0.55)"
			css={{ backdropFilter: "blur(20px) saturate(1.2)", WebkitBackdropFilter: "blur(20px) saturate(1.2)" }}
		>
			<HStack gap="6">
				<HStack gap="3">
					<BrandMark size="sm" />
					<Box fontWeight="bold" fontSize="lg" letterSpacing="-0.01em" color="fg">
						isme
					</Box>
				</HStack>
				<HStack as="nav" gap="1" aria-label="Primary">
					{NAV.map((n) => {
						const isActive = n.key === active;
						return (
							<Box
								as={RouterLink}
								key={n.key}
								to={n.to}
								aria-current={isActive ? "page" : undefined}
								px="3.5"
								py="2"
								fontSize="sm"
								fontWeight="medium"
								borderRadius="md"
								color={isActive ? "fg" : "fg.muted"}
								bg={isActive ? "bg.glass" : "transparent"}
								borderWidth={isActive ? "1px" : "0px"}
								borderColor="border.strong"
								boxShadow={isActive ? "0 0 20px rgba(99,102,241,0.20)" : "none"}
								_hover={{ color: "fg" }}
								textDecoration="none"
							>
								{n.label}
							</Box>
						);
					})}
				</HStack>
			</HStack>
			<HStack gap="2.5">
				<IconButton
					aria-label="Notifications"
					variant="ghost"
					size="md"
					color="fg.subtle"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border.strong"
					borderRadius="md"
					position="relative"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
					_hover={{ color: "fg", borderColor: "rgba(255,255,255,0.28)" }}
				>
					<LuBell />
					<Box
						position="absolute"
						top="1.5"
						right="1.5"
						w="2"
						h="2"
						borderRadius="full"
						bg="aurora.magenta"
						css={{ boxShadow: "0 0 8px #EC4899" }}
					/>
				</IconButton>
				<UserChip name={user.name} email={user.email} />
			</HStack>
		</Flex>
	);
};
```

- [ ] **Step 3:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 4:** No commit yet.

---

### Task 12: StatCard + ActivityRow

**Files:**
- Create: `ui/src/components/ui/stat-card.tsx`
- Create: `ui/src/components/ui/activity-row.tsx`

- [ ] **Step 1:** Create StatCard.

`ui/src/components/ui/stat-card.tsx`:
```tsx
import { Box, Center, Heading, Text, HStack } from "@chakra-ui/react";
import type { ReactNode } from "react";
import { GlassCard } from "./glass-card";

export type StatTone = "cyan" | "violet" | "magenta";

interface StatCardProps {
	icon: ReactNode;
	title: string;
	desc: string;
	stat: string;
	delta: string;
	tone?: StatTone;
}

const toneColor: Record<StatTone, string> = {
	cyan: "#22D3EE",
	violet: "#8B5CF6",
	magenta: "#EC4899",
};

export const StatCard = ({ icon, title, desc, stat, delta, tone = "cyan" }: StatCardProps) => {
	return (
		<GlassCard p="5.5" position="relative" overflow="hidden">
			<Center
				w="9.5"
				h="9.5"
				borderRadius="lg"
				borderWidth="1px"
				borderColor="border.strong"
				color={toneColor[tone]}
				mb="3.5"
				css={{
					background:
						"linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))",
				}}
			>
				{icon}
			</Center>
			<Text fontSize="md" fontWeight="semibold" color="fg" mb="0.5">
				{title}
			</Text>
			<Text fontSize="sm" color="fg.muted" mb="3.5">
				{desc}
			</Text>
			<Heading
				as="div"
				fontSize="32px"
				fontWeight="bold"
				letterSpacing="-0.02em"
				css={{
					backgroundImage: "linear-gradient(135deg, #22D3EE, #8B5CF6)",
					WebkitBackgroundClip: "text",
					backgroundClip: "text",
					color: "transparent",
				}}
			>
				{stat}
			</Heading>
			<HStack gap="1" mt="1" fontSize="xs" color="success" fontWeight="medium">
				<Text as="span">{delta}</Text>
			</HStack>
			<Box
				position="absolute"
				inset="0"
				pointerEvents="none"
				borderRadius="inherit"
				css={{
					background: "linear-gradient(135deg, transparent 60%, rgba(99,102,241,0.15))",
				}}
			/>
		</GlassCard>
	);
};
```

- [ ] **Step 2:** Create ActivityRow.

`ui/src/components/ui/activity-row.tsx`:
```tsx
import { Box, Center, Grid, Text } from "@chakra-ui/react";
import type { ReactNode } from "react";

export type ActivityTone = "ok" | "violet" | "magenta";

interface ActivityRowProps {
	tone: ActivityTone;
	icon: ReactNode;
	body: ReactNode;
	time: string;
}

const toneStyles: Record<ActivityTone, { bg: string; color: string }> = {
	ok: { bg: "rgba(52,211,153,0.15)", color: "#34D399" },
	violet: { bg: "rgba(139,92,246,0.18)", color: "#8B5CF6" },
	magenta: { bg: "rgba(236,72,153,0.18)", color: "#EC4899" },
};

export const ActivityRow = ({ tone, icon, body, time }: ActivityRowProps) => {
	const t = toneStyles[tone];
	return (
		<Grid
			gridTemplateColumns="38px 1fr auto"
			alignItems="center"
			gap="3.5"
			px="3.5"
			py="3.5"
			borderRadius="lg"
			_hover={{ bg: "rgba(255,255,255,0.04)" }}
		>
			<Center w="9.5" h="9.5" borderRadius="lg" borderWidth="1px" borderColor="border.strong" css={{ background: t.bg, color: t.color }}>
				{icon}
			</Center>
			<Box fontSize="sm" color="fg">
				{body}
			</Box>
			<Text fontSize="xs" color="fg.muted">
				{time}
			</Text>
		</Grid>
	);
};
```

- [ ] **Step 3:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 4:** No commit yet.

---

### Task 13: Mock data

**Files:**
- Create: `ui/src/consts/mock.tsx`

> **Note:** Use `.tsx` extension because mock body strings include inline JSX `<b>` for emphasis.

- [ ] **Step 1:** Create.

`ui/src/consts/mock.tsx`:
```tsx
import type { ReactNode } from "react";
import { LuMonitor, LuClock, LuShieldCheck, LuCheck, LuKey, LuUserPlus } from "react-icons/lu";
import type { StatTone } from "@/components/ui/stat-card";
import type { ActivityTone } from "@/components/ui/activity-row";

export interface StatEntry {
	tone: StatTone;
	icon: ReactNode;
	title: string;
	desc: string;
	stat: string;
	delta: string;
}

export interface ActivityEntry {
	tone: ActivityTone;
	icon: ReactNode;
	body: ReactNode;
	time: string;
}

export const MOCK_STATS: StatEntry[] = [
	{
		tone: "cyan",
		icon: <LuMonitor />,
		title: "Active sessions",
		desc: "Devices currently signed in.",
		stat: "3",
		delta: "▲ +1 since yesterday",
	},
	{
		tone: "violet",
		icon: <LuClock />,
		title: "Token rotations",
		desc: "Refreshes in last 24h.",
		stat: "128",
		delta: "▲ +12% w/w",
	},
	{
		tone: "magenta",
		icon: <LuShieldCheck />,
		title: "Security score",
		desc: "Account hardening level.",
		stat: "A+",
		delta: "● 2FA enabled",
	},
];

export const MOCK_ACTIVITY: ActivityEntry[] = [
	{
		tone: "ok",
		icon: <LuCheck />,
		body: (
			<>
				Sign-in from <b>MacBook · Safari</b> · Hồ Chí Minh
			</>
		),
		time: "just now",
	},
	{
		tone: "violet",
		icon: <LuKey />,
		body: (
			<>
				API key rotated for <b>billing-service</b>
			</>
		),
		time: "2h ago",
	},
	{
		tone: "magenta",
		icon: <LuUserPlus />,
		body: (
			<>
				Invited <b>thanhlp3@hasaki.vn</b> as Admin
			</>
		),
		time: "yesterday",
	},
];
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 14: AuthLayout

**Files:**
- Create: `ui/src/layouts/AuthLayout.tsx`

- [ ] **Step 1:** Create.

`ui/src/layouts/AuthLayout.tsx`:
```tsx
import { Box, Flex, Grid } from "@chakra-ui/react";
import type { ReactNode } from "react";

interface AuthLayoutProps {
	topRight?: ReactNode;
	brand: ReactNode;
	children: ReactNode;
}

export const AuthLayout = ({ topRight, brand, children }: AuthLayoutProps) => {
	return (
		<Flex w="full" minH="100vh" align="center" justify="center" px={{ base: "4", md: "6" }} py="12">
			<Box
				w="full"
				maxW="1280px"
				bg="bg.glass"
				borderWidth="1px"
				borderColor="border"
				borderRadius="3xl"
				boxShadow="glassSoft"
				overflow="hidden"
				position="relative"
			>
				<Grid
					templateColumns={{ base: "1fr", md: "1.05fr 1fr" }}
					minH={{ base: "auto", md: "720px" }}
				>
					{brand}
					<Flex
						direction="column"
						justify="center"
						p={{ base: "8", md: "14" }}
						bg="rgba(7,7,26,0.55)"
						css={{ backdropFilter: "blur(8px)", WebkitBackdropFilter: "blur(8px)" }}
					>
						{topRight && (
							<Flex justify="flex-end" mb="9">
								{topRight}
							</Flex>
						)}
						<Box w="full" maxW="440px" mx="auto">
							{children}
						</Box>
					</Flex>
				</Grid>
			</Box>
		</Flex>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 15: AppShell

**Files:**
- Create: `ui/src/layouts/AppShell.tsx`

- [ ] **Step 1:** Create.

`ui/src/layouts/AppShell.tsx`:
```tsx
import { Box, Stack } from "@chakra-ui/react";
import type { ReactNode } from "react";
import { Topbar, type TopbarTab } from "@/components/ui/topbar";

interface AppShellProps {
	active: TopbarTab;
	user: { name: string; email: string };
	children: ReactNode;
}

export const AppShell = ({ active, user, children }: AppShellProps) => {
	return (
		<Box minH="100vh" w="full">
			<Topbar active={active} user={user} />
			<Stack as="main" gap="6" p="7" maxW="1280px" mx="auto">
				{children}
			</Stack>
		</Box>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** Commit Phase 2 batch.

```bash
git add ui/src/components/ui/aurora-background.tsx \
        ui/src/components/ui/brand-mark.tsx \
        ui/src/components/ui/glass-card.tsx \
        ui/src/components/ui/feature-list.tsx \
        ui/src/components/ui/brand-panel.tsx \
        ui/src/components/ui/password-field.tsx \
        ui/src/components/ui/password-strength.tsx \
        ui/src/components/ui/topbar.tsx \
        ui/src/components/ui/user-chip.tsx \
        ui/src/components/ui/stat-card.tsx \
        ui/src/components/ui/activity-row.tsx \
        ui/src/consts/mock.tsx \
        ui/src/layouts/AuthLayout.tsx \
        ui/src/layouts/AppShell.tsx
git commit -m "feat(ui): aurora layouts and atoms

AuthLayout (split brand+form), AppShell (topbar+main),
plus AuroraBackground, BrandMark, BrandPanel, GlassCard,
FeatureList, PasswordField, PasswordStrength, Topbar,
UserChip, StatCard, ActivityRow. Mock data in consts/mock.tsx.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase 3 — Page Rewrites

### Task 16: Mount AuroraBackground in App

**Files:**
- Modify: `ui/src/App.tsx`

- [ ] **Step 1:** Replace App.tsx.

`ui/src/App.tsx`:
```tsx
import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { SSOLogin } from "./pages/SSOLogin";
import { Signup } from "./pages/Signup";
import { Welcome } from "./pages/Welcome";
import { NotFound } from "./pages/NotFound";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AuroraBackground } from "./components/ui/aurora-background";

function App() {
	return (
		<>
			<AuroraBackground />
			<BrowserRouter>
				<Routes>
					<Route path="/login" element={<Login />} />
					<Route path="/sso/login" element={<SSOLogin />} />
					<Route path="/signup" element={<Signup />} />
					<Route
						path="/welcome"
						element={
							<ProtectedRoute>
								<Welcome />
							</ProtectedRoute>
						}
					/>
					<Route path="/" element={<Navigate to="/welcome" replace />} />
					<Route path="/404" element={<NotFound />} />
				</Routes>
			</BrowserRouter>
		</>
	);
}

export default App;
```

> **Note:** New routes (`/forgot-password`, `/sessions`, `/team`, `/settings`) added in Task 21 once stub pages exist.

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet — bundled with Login/Signup/Welcome rewrites.

---

### Task 17: Rewrite Login

**Files:**
- Modify: `ui/src/pages/Login.tsx` (full rewrite)

- [ ] **Step 1:** Replace contents.

`ui/src/pages/Login.tsx`:
```tsx
"use client";

import { useState } from "react";
import { Button, Field, Flex, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { LuArrowRight, LuMail, LuShieldCheck, LuClock, LuCheckCheck } from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";
import { loginSchema, type LoginFormData } from "@/validators";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { toaster } from "@/components/ui/toaster";
import { PasswordField } from "@/components/ui/password-field";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";

const LOGIN_FEATURES = [
	{ icon: <LuShieldCheck />, title: "JWT + refresh rotation", desc: "HS256, rotation built-in." },
	{ icon: <LuClock />, title: "Per-device sessions", desc: "Revoke, list, audit instantly." },
	{ icon: <LuCheckCheck />, title: "Compliant by default", desc: "WCAG AA, audit-ready logs." },
];

export const Login = () => {
	const { login, loading, error } = useAuth();
	const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof LoginFormData, string>>>({});
	const [rememberMe, setRememberMe] = useState(false);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		const result = loginSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof LoginFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) errors[err.path[0] as keyof LoginFormData] = err.message;
			});
			setFormErrors(errors);
			return;
		}
		try {
			await login(formData);
			toaster.create({ title: "Login successful", description: "Welcome back!", type: "success" });
		} catch (err) {
			toaster.create({
				title: "Login failed",
				description: error?.message || "Please check your credentials",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof LoginFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) setFormErrors((prev) => ({ ...prev, [field]: undefined }));
	};

	return (
		<AuthLayout
			topRight={
				<Text fontSize="sm" color="fg.muted">
					New here?{" "}
					<RouterLink to="/signup" style={{ color: "var(--chakra-colors-fg)", fontWeight: 600, borderBottom: "1px solid var(--chakra-colors-aurora-violet)", paddingBottom: 1 }}>
						Create account
					</RouterLink>
				</Text>
			}
			brand={
				<BrandPanel
					pill="Welcome back"
					pillTone="violet"
					titleLead="Build, ship, repeat —"
					titleGrad="without the chaos."
					sub="One workspace for auth, sessions, teams. Aurora-grade security, quiet by default."
					features={LOGIN_FEATURES}
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Sign in
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Use your work email and password. SSO available.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				<Field.Root invalid={!!formErrors.email}>
					<Field.Label>Email</Field.Label>
					<Input
						type="email"
						autoComplete="email"
						placeholder="you@company.com"
						value={formData.email}
						onChange={handleChange("email")}
						startElement={<LuMail />}
					/>
					{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
				</Field.Root>

				<PasswordField
					label="Password"
					value={formData.password}
					onChange={handleChange("password")}
					error={formErrors.password}
					autoComplete="current-password"
					placeholder="Enter password"
				/>

				<Flex justify="space-between" align="center" mt="1" mb="2">
					<Checkbox checked={rememberMe} onCheckedChange={(e) => setRememberMe(!!e.checked)}>
						<Text fontSize="sm" color="fg.subtle">Remember me</Text>
					</Checkbox>
					<RouterLink
						to="/forgot-password"
						style={{ color: "var(--chakra-colors-fg-subtle)", fontSize: 14, fontWeight: 500, textDecoration: "none" }}
					>
						Forgot password?
					</RouterLink>
				</Flex>

				<Button
					type="submit"
					h="12"
					loading={loading}
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					_focusVisible={{ boxShadow: "focusRing" }}
					css={{
						background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
						backgroundSize: "200% 200%",
					}}
				>
					<HStack gap="2.5">
						<Text>Sign in</Text>
						<LuArrowRight />
					</HStack>
				</Button>
			</Stack>
		</AuthLayout>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 18: Rewrite Signup

**Files:**
- Modify: `ui/src/pages/Signup.tsx` (full rewrite)

- [ ] **Step 1:** Replace contents.

`ui/src/pages/Signup.tsx`:
```tsx
"use client";

import { useState } from "react";
import { Box, Button, Field, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import {
	LuArrowRight,
	LuCheckCheck,
	LuLock,
	LuMail,
	LuShieldCheck,
	LuUser,
	LuUsers,
} from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";
import { signupSchema, type SignupFormData } from "@/validators";
import { Input } from "@/components/ui/input";
import { toaster } from "@/components/ui/toaster";
import { PasswordField } from "@/components/ui/password-field";
import { PasswordStrength } from "@/components/ui/password-strength";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";

const SIGNUP_FEATURES = [
	{ icon: <LuUsers />, title: "One-click team invites", desc: "Email or magic link." },
	{ icon: <LuLock />, title: "Encryption at rest", desc: "RSA 2048 keys, never logged." },
	{ icon: <LuCheckCheck />, title: "Audit-ready", desc: "Every session, every token, logged." },
];

export const Signup = () => {
	const { signup, loading, error } = useAuth();
	const [formData, setFormData] = useState<SignupFormData>({ name: "", email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof SignupFormData, string>>>({});

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		const result = signupSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof SignupFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) errors[err.path[0] as keyof SignupFormData] = err.message;
			});
			setFormErrors(errors);
			return;
		}
		try {
			await signup(formData);
			toaster.create({
				title: "Signup successful",
				description: "Your account has been created. Please login.",
				type: "success",
			});
		} catch (err) {
			toaster.create({
				title: "Signup failed",
				description: error?.message || "Please check your information and try again",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof SignupFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) setFormErrors((prev) => ({ ...prev, [field]: undefined }));
	};

	return (
		<AuthLayout
			topRight={
				<Text fontSize="sm" color="fg.muted">
					Already have an account?{" "}
					<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg)", fontWeight: 600, borderBottom: "1px solid var(--chakra-colors-aurora-violet)", paddingBottom: 1 }}>
						Sign in
					</RouterLink>
				</Text>
			}
			brand={
				<BrandPanel
					pill="Free for solo devs"
					pillTone="cyan"
					titleLead="Start in"
					titleGrad="under 2 minutes."
					sub="No credit card. Cancel anytime. Your tokens, your rules."
					features={SIGNUP_FEATURES}
					footerRight="SOC 2 in progress"
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Create your account
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Spin up a workspace and bring your team in.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				<Field.Root invalid={!!formErrors.name}>
					<Field.Label>Full name</Field.Label>
					<Input
						type="text"
						autoComplete="name"
						placeholder="Test Me"
						value={formData.name}
						onChange={handleChange("name")}
						startElement={<LuUser />}
					/>
					{formErrors.name && <Field.ErrorText>{formErrors.name}</Field.ErrorText>}
				</Field.Root>

				<Field.Root invalid={!!formErrors.email}>
					<Field.Label>Work email</Field.Label>
					<Input
						type="email"
						autoComplete="email"
						placeholder="you@company.com"
						value={formData.email}
						onChange={handleChange("email")}
						startElement={<LuMail />}
					/>
					{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
					<Field.HelperText>We'll send a verification link.</Field.HelperText>
				</Field.Root>

				<Box>
					<PasswordField
						label="Password"
						value={formData.password}
						onChange={handleChange("password")}
						error={formErrors.password}
						autoComplete="new-password"
						placeholder="At least 8 characters"
					/>
					<PasswordStrength value={formData.password} />
					<Text fontSize="xs" color="fg.muted" mt="2">
						Use 8+ characters with letters, numbers, and a symbol.
					</Text>
				</Box>

				<Button
					type="submit"
					mt="3"
					h="12"
					loading={loading}
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					_focusVisible={{ boxShadow: "focusRing" }}
					css={{
						background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
						backgroundSize: "200% 200%",
					}}
				>
					<HStack gap="2.5">
						<Text>Create account</Text>
						<LuArrowRight />
					</HStack>
				</Button>
			</Stack>
		</AuthLayout>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 19: Rewrite Welcome

**Files:**
- Modify: `ui/src/pages/Welcome.tsx` (full rewrite)

- [ ] **Step 1:** Replace contents.

`ui/src/pages/Welcome.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { Box, Button, Center, Flex, Grid, HStack, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuArrowRight, LuFileText, LuPlus } from "react-icons/lu";
import { getCurrentUser } from "@/apis/auth";
import { GlassCard } from "@/components/ui/glass-card";
import { StatCard } from "@/components/ui/stat-card";
import { ActivityRow } from "@/components/ui/activity-row";
import { AppShell } from "@/layouts/AppShell";
import { MOCK_STATS, MOCK_ACTIVITY } from "@/consts/mock";
import type { GetMeResponse } from "@/types";

export const Welcome = () => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		const fetchUser = async () => {
			try {
				setLoading(true);
				setError(null);
				const response = await getCurrentUser();
				setUser((response as any).data as GetMeResponse);
			} catch (err: any) {
				setError(err?.response?.data?.message || "Failed to load user data");
			} finally {
				setLoading(false);
			}
		};
		fetchUser();
	}, []);

	if (loading) {
		return (
			<Center w="full" h="100vh">
				<Stack align="center" gap="4">
					<Spinner size="xl" color="accent" />
					<Text color="fg.muted">Loading...</Text>
				</Stack>
			</Center>
		);
	}

	if (error) {
		return (
			<Center w="full" h="100vh">
				<Text color="danger">{error}</Text>
			</Center>
		);
	}

	const name = user?.name || "User";
	const email = user?.email || "";

	return (
		<AppShell active="overview" user={{ name, email }}>
			{/* Hero */}
			<GlassCard p="9" position="relative" overflow="hidden">
				<Box position="absolute" w="300px" h="300px" right="-60px" top="-100px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(50px)", background: "radial-gradient(circle, rgba(99,102,241,0.55), transparent 70%)" }} />
				<Box position="absolute" w="240px" h="240px" right="30%" bottom="-120px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(50px)", background: "radial-gradient(circle, rgba(236,72,153,0.45), transparent 70%)" }} />
				<Box position="absolute" w="200px" h="200px" left="10%" top="-50px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(50px)", background: "radial-gradient(circle, rgba(34,211,238,0.40), transparent 70%)" }} />
				<Box position="relative" zIndex={1}>
					<HStack
						display="inline-flex"
						gap="2"
						px="3"
						py="1.5"
						borderRadius="full"
						bg="bg.muted"
						color="fg"
					>
						<Box w="2" h="2" borderRadius="full" bg="success" css={{ boxShadow: "0 0 0 4px rgba(22,163,74,0.18)" }} />
						<Text fontSize="xs" fontWeight="medium">You're signed in</Text>
					</HStack>
					<Heading
						as="h1"
						mt="3.5"
						mb="2"
						fontSize={{ base: "30px", md: "42px" }}
						lineHeight="1.1"
						letterSpacing="-0.025em"
					>
						Welcome back,{" "}
						<Text
							as="span"
							css={{
								backgroundImage: "linear-gradient(135deg, #22D3EE, #8B5CF6 60%, #EC4899)",
								WebkitBackgroundClip: "text",
								backgroundClip: "text",
								color: "transparent",
							}}
						>
							{name}
						</Text>
					</Heading>
					<Text color="fg.subtle" mb="6" maxW="560px">
						Last sign-in just now from this device. Pick up where you left off, or spin up something new.
					</Text>
					<HStack gap="2.5" flexWrap="wrap">
						<Button
							h="12"
							px="5.5"
							color="white"
							borderRadius="glassSm"
							boxShadow="ctaGlow"
							_hover={{ boxShadow: "ctaGlowHi" }}
							_focusVisible={{ boxShadow: "focusRing" }}
							css={{
								background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
								backgroundSize: "200% 200%",
							}}
						>
							<HStack gap="2.5"><LuPlus /><Text>Start a project</Text></HStack>
						</Button>
						<Button
							h="12"
							px="5"
							variant="outline"
							color="fg"
							borderColor="border.strong"
							bg="bg.glass"
							borderRadius="glassSm"
							css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
							_hover={{ bg: "bg.glassHi" }}
						>
							<HStack gap="2.5"><LuFileText /><Text>View docs</Text></HStack>
						</Button>
					</HStack>
				</Box>
			</GlassCard>

			{/* Stats */}
			<Grid templateColumns={{ base: "1fr", lg: "repeat(3, 1fr)" }} gap="4">
				{MOCK_STATS.map((s, i) => (
					<StatCard key={i} {...s} />
				))}
			</Grid>

			{/* Activity */}
			<Box>
				<Flex justify="space-between" align="baseline" mb="3">
					<Heading as="h2" fontSize="lg" letterSpacing="-0.01em" color="fg">
						Recent activity
					</Heading>
					<Text as="a" href="#" fontSize="sm" color="fg.subtle" _hover={{ color: "accentAlt" }}>
						View all →
					</Text>
				</Flex>
				<GlassCard p="2">
					<Stack
						gap="0"
						css={{
							"& > * + *": {
								borderTop: "1px solid var(--chakra-colors-border)",
							},
						}}
					>
						{MOCK_ACTIVITY.map((a, i) => (
							<ActivityRow key={i} {...a} />
						))}
					</Stack>
				</GlassCard>
			</Box>
		</AppShell>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** Manual smoke.

```bash
cd ui && npm run dev
```

Open `http://localhost:5173/login` (URL printed by Vite). Verify:
- Aurora flowing background visible.
- Split layout: brand panel left (logo + pill + iridescent title + 3 features), form right.
- Email + password fields render with icons; eye toggle on password works (click → input type changes to text).
- "Forgot password?" link present (broken target — fixed in Phase 4).
- Submit button has aurora gradient + glow.
- Resize to <900px width: panels stack.
- Visit `/signup`: strength meter updates as you type (`Aa1!1234` → 4 bars lit).
- Login with valid creds → `/welcome` shows topbar + 3 stat cards + activity list.
- Topbar UserChip click → menu with `Log out` works.

Stop dev server (`Ctrl+C`).

- [ ] **Step 4:** Commit Phase 3.

```bash
git add ui/src/App.tsx ui/src/pages/Login.tsx ui/src/pages/Signup.tsx ui/src/pages/Welcome.tsx
git commit -m "feat(ui): redesign Login + Signup + Welcome aurora

Login + Signup: AuthLayout + BrandPanel + PasswordField, drop
T&C and old gradient Box. Signup adds PasswordStrength meter.
Welcome wraps AppShell with hero + 3 StatCards + ActivityRow
list (mock data). Mount AuroraBackground at App root.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase 4 — Stub Pages + Routes

### Task 20: ForgotPassword

**Files:**
- Create: `ui/src/pages/ForgotPassword.tsx`

- [ ] **Step 1:** Create.

`ui/src/pages/ForgotPassword.tsx`:
```tsx
"use client";

import { useState } from "react";
import { Button, Field, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { LuArrowRight, LuMail, LuShieldCheck, LuClock, LuCheckCheck } from "react-icons/lu";
import { Input } from "@/components/ui/input";
import { toaster } from "@/components/ui/toaster";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";

const RECOVERY_FEATURES = [
	{ icon: <LuShieldCheck />, title: "Encrypted reset tokens", desc: "Single-use, expiring." },
	{ icon: <LuClock />, title: "5-minute window", desc: "Short-lived links." },
	{ icon: <LuCheckCheck />, title: "Audit trail", desc: "Reset attempts logged." },
];

export const ForgotPassword = () => {
	const [email, setEmail] = useState("");

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		toaster.create({
			title: "Coming soon",
			description: "Reset flow not implemented yet.",
			type: "info",
		});
	};

	return (
		<AuthLayout
			topRight={
				<Text fontSize="sm" color="fg.muted">
					Remember it?{" "}
					<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg)", fontWeight: 600, borderBottom: "1px solid var(--chakra-colors-aurora-violet)", paddingBottom: 1 }}>
						Sign in
					</RouterLink>
				</Text>
			}
			brand={
				<BrandPanel
					pill="Account recovery"
					pillTone="cyan"
					titleLead="Forgot your"
					titleGrad="password?"
					sub="Drop your email — we'll send a reset link in seconds."
					features={RECOVERY_FEATURES}
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Reset password
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Enter the email you used to register.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				<Field.Root>
					<Field.Label>Email</Field.Label>
					<Input
						type="email"
						autoComplete="email"
						placeholder="you@company.com"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						startElement={<LuMail />}
					/>
				</Field.Root>
				<Button
					type="submit"
					mt="2"
					h="12"
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					_focusVisible={{ boxShadow: "focusRing" }}
					css={{
						background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
						backgroundSize: "200% 200%",
					}}
				>
					<HStack gap="2.5">
						<Text>Send reset link</Text>
						<LuArrowRight />
					</HStack>
				</Button>
			</Stack>
		</AuthLayout>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 21: Sessions / Team / Settings stubs

**Files:**
- Create: `ui/src/pages/Sessions.tsx`
- Create: `ui/src/pages/Team.tsx`
- Create: `ui/src/pages/Settings.tsx`

> **Note:** Each stub is a thin wrapper around AppShell + a centered "Coming soon" GlassCard. Repeating per-file rather than abstracting because there are only three and they will likely diverge later.

> **Shared dependency note:** Stubs need access to `user` for Topbar (name+email). They call `getCurrentUser` like Welcome to keep behavior consistent under `<ProtectedRoute>`.

- [ ] **Step 1:** Create Sessions.

`ui/src/pages/Sessions.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuMonitor } from "react-icons/lu";
import { getCurrentUser } from "@/apis/auth";
import { GlassCard } from "@/components/ui/glass-card";
import { AppShell } from "@/layouts/AppShell";
import type { GetMeResponse } from "@/types";

export const Sessions = () => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		getCurrentUser()
			.then((r) => setUser((r as any).data as GetMeResponse))
			.finally(() => setLoading(false));
	}, []);

	if (loading) {
		return (
			<Center w="full" h="100vh">
				<Spinner size="xl" color="accent" />
			</Center>
		);
	}

	return (
		<AppShell active="sessions" user={{ name: user?.name || "User", email: user?.email || "" }}>
			<GlassCard p="14">
				<Stack align="center" gap="4" textAlign="center">
					<Box color="accentAlt" fontSize="3xl"><LuMonitor /></Box>
					<Heading as="h1" fontSize="2xl" color="fg">Sessions</Heading>
					<Text color="fg.muted">Coming soon.</Text>
				</Stack>
			</GlassCard>
		</AppShell>
	);
};
```

- [ ] **Step 2:** Create Team.

`ui/src/pages/Team.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuUsers } from "react-icons/lu";
import { getCurrentUser } from "@/apis/auth";
import { GlassCard } from "@/components/ui/glass-card";
import { AppShell } from "@/layouts/AppShell";
import type { GetMeResponse } from "@/types";

export const Team = () => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		getCurrentUser()
			.then((r) => setUser((r as any).data as GetMeResponse))
			.finally(() => setLoading(false));
	}, []);

	if (loading) {
		return (
			<Center w="full" h="100vh">
				<Spinner size="xl" color="accent" />
			</Center>
		);
	}

	return (
		<AppShell active="team" user={{ name: user?.name || "User", email: user?.email || "" }}>
			<GlassCard p="14">
				<Stack align="center" gap="4" textAlign="center">
					<Box color="accentAlt" fontSize="3xl"><LuUsers /></Box>
					<Heading as="h1" fontSize="2xl" color="fg">Team</Heading>
					<Text color="fg.muted">Coming soon.</Text>
				</Stack>
			</GlassCard>
		</AppShell>
	);
};
```

- [ ] **Step 3:** Create Settings.

`ui/src/pages/Settings.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuSettings } from "react-icons/lu";
import { getCurrentUser } from "@/apis/auth";
import { GlassCard } from "@/components/ui/glass-card";
import { AppShell } from "@/layouts/AppShell";
import type { GetMeResponse } from "@/types";

export const Settings = () => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		getCurrentUser()
			.then((r) => setUser((r as any).data as GetMeResponse))
			.finally(() => setLoading(false));
	}, []);

	if (loading) {
		return (
			<Center w="full" h="100vh">
				<Spinner size="xl" color="accent" />
			</Center>
		);
	}

	return (
		<AppShell active="settings" user={{ name: user?.name || "User", email: user?.email || "" }}>
			<GlassCard p="14">
				<Stack align="center" gap="4" textAlign="center">
					<Box color="accentAlt" fontSize="3xl"><LuSettings /></Box>
					<Heading as="h1" fontSize="2xl" color="fg">Settings</Heading>
					<Text color="fg.muted">Coming soon.</Text>
				</Stack>
			</GlassCard>
		</AppShell>
	);
};
```

- [ ] **Step 4:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 5:** No commit yet.

---

### Task 22: Wire new routes in App.tsx

**Files:**
- Modify: `ui/src/App.tsx`

- [ ] **Step 1:** Replace App.tsx with full route table.

`ui/src/App.tsx`:
```tsx
import "./App.css";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Login } from "./pages/Login";
import { SSOLogin } from "./pages/SSOLogin";
import { Signup } from "./pages/Signup";
import { Welcome } from "./pages/Welcome";
import { ForgotPassword } from "./pages/ForgotPassword";
import { Sessions } from "./pages/Sessions";
import { Team } from "./pages/Team";
import { Settings } from "./pages/Settings";
import { NotFound } from "./pages/NotFound";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AuroraBackground } from "./components/ui/aurora-background";

function App() {
	return (
		<>
			<AuroraBackground />
			<BrowserRouter>
				<Routes>
					<Route path="/login" element={<Login />} />
					<Route path="/sso/login" element={<SSOLogin />} />
					<Route path="/signup" element={<Signup />} />
					<Route path="/forgot-password" element={<ForgotPassword />} />
					<Route
						path="/welcome"
						element={
							<ProtectedRoute>
								<Welcome />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/sessions"
						element={
							<ProtectedRoute>
								<Sessions />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/team"
						element={
							<ProtectedRoute>
								<Team />
							</ProtectedRoute>
						}
					/>
					<Route
						path="/settings"
						element={
							<ProtectedRoute>
								<Settings />
							</ProtectedRoute>
						}
					/>
					<Route path="/" element={<Navigate to="/welcome" replace />} />
					<Route path="/404" element={<NotFound />} />
				</Routes>
			</BrowserRouter>
		</>
	);
}

export default App;
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** Manual smoke.

```bash
cd ui && npm run dev
```

Walk:
- `/forgot-password` — form renders, submit shows `Coming soon` toast.
- Login → `/welcome` → click `Sessions` nav → stub renders with topbar still active highlight on `Sessions`.
- Same for `Team`, `Settings`.

Stop dev (`Ctrl+C`).

- [ ] **Step 4:** Commit.

```bash
git add ui/src/App.tsx \
        ui/src/pages/ForgotPassword.tsx \
        ui/src/pages/Sessions.tsx \
        ui/src/pages/Team.tsx \
        ui/src/pages/Settings.tsx
git commit -m "feat(ui): forgot-password stub + sessions/team/settings pages

ForgotPassword: AuthLayout form posting Coming soon toast.
Sessions/Team/Settings: AppShell wrapping centered GlassCard
placeholder. Routes wired under ProtectedRoute (except
forgot-password which is public).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase 5 — Retheme Leftovers

### Task 23: SSOLogin token swap

**Files:**
- Modify: `ui/src/pages/SSOLogin.tsx`

- [ ] **Step 1:** Replace contents (keep single-card layout, swap colors).

`ui/src/pages/SSOLogin.tsx`:
```tsx
"use client";

import { useEffect, useState } from "react";
import { Box, HStack, Stack, Text, Field } from "@chakra-ui/react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { loginSchema, type LoginFormData } from "@/validators";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toaster } from "@/components/ui/toaster";
import { LuArrowRight, LuMail } from "react-icons/lu";
import { PasswordField } from "@/components/ui/password-field";

export const SSOLogin = () => {
	const { login, loading, error } = useAuth();
	const [searchParams] = useSearchParams();
	const navigate = useNavigate();
	const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof LoginFormData, string>>>({});

	useEffect(() => {
		const sessionId = searchParams.get("session_id");
		if (!sessionId || sessionId.trim() === "") {
			navigate("/404", { replace: true });
		}
	}, [searchParams, navigate]);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		const result = loginSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof LoginFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) errors[err.path[0] as keyof LoginFormData] = err.message;
			});
			setFormErrors(errors);
			return;
		}
		const sessionId = searchParams.get("session_id");
		if (!sessionId || sessionId.trim() === "") {
			toaster.create({ title: "Invalid session", description: "Session ID is required", type: "error" });
			return;
		}
		try {
			const response = await login({ ...formData, session_id: sessionId });
			if (response.data.redirect_url && response.data.authorization_code) {
				const redirectUrl = new URL(response.data.redirect_url);
				redirectUrl.searchParams.set("authorization_code", response.data.authorization_code);
				window.open(redirectUrl.toString(), "_blank");
				toaster.create({ title: "Login successful", description: "Redirecting to application...", type: "success" });
			} else {
				toaster.create({ title: "Login successful", description: "Welcome back!", type: "success" });
			}
		} catch (err) {
			toaster.create({
				title: "Login failed",
				description: error?.message || "Please check your credentials",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof LoginFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) setFormErrors((prev) => ({ ...prev, [field]: undefined }));
	};

	return (
		<Box w="full" minH="100vh" display="flex" alignItems="center" justifyContent="center" px="4">
			<Card w="full" maxW="md" p="8" bg="bg.glass" borderColor="border.strong" borderWidth="1px" borderRadius="2xl" boxShadow="glassSoft" position="relative" zIndex="1" css={{ backdropFilter: "blur(20px)", WebkitBackdropFilter: "blur(20px)" }}>
				<Stack gap="6" as="form" onSubmit={handleSubmit}>
					<Text fontSize="2xl" fontWeight="bold" textAlign="center" color="fg">
						Sign in
					</Text>
					<Stack gap="4">
						<Field.Root invalid={!!formErrors.email}>
							<Field.Label>Email</Field.Label>
							<Input
								type="email"
								placeholder="you@company.com"
								value={formData.email}
								onChange={handleChange("email")}
								startElement={<LuMail />}
							/>
							{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
						</Field.Root>
						<PasswordField
							label="Password"
							value={formData.password}
							onChange={handleChange("password")}
							error={formErrors.password}
							autoComplete="current-password"
							placeholder="Password"
						/>
					</Stack>
					<Button
						type="submit"
						h="12"
						loading={loading}
						color="white"
						borderRadius="glassSm"
						boxShadow="ctaGlow"
						_hover={{ boxShadow: "ctaGlowHi" }}
						_focusVisible={{ boxShadow: "focusRing" }}
						css={{
							background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
							backgroundSize: "200% 200%",
						}}
					>
						<HStack gap="2.5"><Text>Sign in</Text><LuArrowRight /></HStack>
					</Button>
				</Stack>
			</Card>
		</Box>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 24: NotFound retheme

**Files:**
- Modify: `ui/src/pages/NotFound.tsx`

- [ ] **Step 1:** Replace.

`ui/src/pages/NotFound.tsx`:
```tsx
"use client";

import { Box, Heading, HStack, Stack, Text } from "@chakra-ui/react";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";
import { LuArrowRight } from "react-icons/lu";

export const NotFound = () => {
	const navigate = useNavigate();

	return (
		<Box w="full" minH="100vh" display="flex" alignItems="center" justifyContent="center" px="4">
			<Stack align="center" gap="6" textAlign="center">
				<Heading
					as="h1"
					fontSize="6xl"
					letterSpacing="-0.04em"
					css={{
						backgroundImage: "linear-gradient(135deg, #22D3EE, #8B5CF6 60%, #EC4899)",
						WebkitBackgroundClip: "text",
						backgroundClip: "text",
						color: "transparent",
					}}
				>
					404
				</Heading>
				<Text fontSize="xl" color="fg">Page Not Found</Text>
				<Text color="fg.muted" fontSize="sm">The page you are looking for does not exist.</Text>
				<Button
					h="12"
					px="6"
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					css={{
						background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
						backgroundSize: "200% 200%",
					}}
					onClick={() => navigate("/")}
				>
					<HStack gap="2.5"><Text>Go Home</Text><LuArrowRight /></HStack>
				</Button>
			</Stack>
		</Box>
	);
};
```

- [ ] **Step 2:** Build.

```bash
cd ui && npm run build
```

Expected: pass.

- [ ] **Step 3:** No commit yet.

---

### Task 25: Home + ProtectedRoute spinner

**Files:**
- Modify: `ui/src/pages/Home.tsx`
- Modify: `ui/src/components/ProtectedRoute.tsx`

- [ ] **Step 1:** Replace Home.

`ui/src/pages/Home.tsx`:
```tsx
import { Box } from "@chakra-ui/react";

export const Home = () => {
	return <Box w="full" minH="100vh" bg="bg" />;
};
```

- [ ] **Step 2:** Patch ProtectedRoute spinner color.

In `ui/src/components/ProtectedRoute.tsx` find:

```tsx
<Spinner size="xl" color="brand.500" />
```

Replace with:

```tsx
<Spinner size="xl" color="accent" />
```

(Only one occurrence inside the loading branch.)

- [ ] **Step 3:** Build + lint.

```bash
cd ui && npm run build && npm run lint
```

Expected: both pass.

- [ ] **Step 4:** Manual smoke.

```bash
cd ui && npm run dev
```

Walk:
- `/sso/login?session_id=test-123` — single card renders with aurora colors, eye toggle on password works.
- `/404` — gradient 404 + Go Home button.
- `/` — should redirect to `/welcome` (or `/login` if not authed). ProtectedRoute spinner is now violet.

Stop dev (`Ctrl+C`).

- [ ] **Step 5:** Commit.

```bash
git add ui/src/pages/SSOLogin.tsx \
        ui/src/pages/NotFound.tsx \
        ui/src/pages/Home.tsx \
        ui/src/components/ProtectedRoute.tsx
git commit -m "chore(ui): retheme SSOLogin + NotFound + Home + ProtectedRoute

SSOLogin: glass card on aurora bg, PasswordField + aurora CTA,
single-card layout preserved. NotFound: iridescent 404 + aurora
button. Home: minimal bg=bg. ProtectedRoute spinner color
brand.500 -> accent.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase 6 — Final Verification

### Task 26: Full build + lint + smoke walk

**Files:** none (verification only)

- [ ] **Step 1:** Clean rebuild.

```bash
cd ui && rm -rf dist && npm run build
```

Expected: build completes, `dist/` populated.

- [ ] **Step 2:** Lint.

```bash
cd ui && npm run lint
```

Expected: no errors, no warnings.

- [ ] **Step 3:** TypeCheck via build (already covered by step 1, but explicit).

```bash
cd ui && npx tsc -b --pretty
```

Expected: pass.

- [ ] **Step 4:** Manual smoke walk — full route matrix.

Start dev:

```bash
cd ui && npm run dev
```

Verify each route:

| Route | Expected |
|-------|----------|
| `/login` | Split layout, aurora bg, password eye toggle, "Forgot password?" link |
| `/signup` | Split layout, password strength meter updates per keystroke (`A`→1, `Aa`→1, `Aaa11111`→3, `Aaa11111!`→4), no T&C checkbox |
| `/forgot-password` | Split layout, submit shows `Coming soon` toast |
| `/sso/login?session_id=test-123` | Single-card layout, aurora colors, password eye toggle |
| `/sso/login` (no session_id) | Redirects to `/404` |
| `/404` | Iridescent 404, Go Home button |
| Login w/ valid creds → `/welcome` | Topbar (Overview active), gradient name in hero, 3 stat cards, activity feed |
| `/welcome` → click `Sessions` | Stub page, topbar Sessions active |
| `/welcome` → click `Team` | Stub page, topbar Team active |
| `/welcome` → click `Settings` | Stub page, topbar Settings active |
| Topbar UserChip → `Log out` | Returns to `/login` |
| Resize <900px on `/login` | Brand panel + form panel stack vertically |
| `prefers-reduced-motion: reduce` (DevTools rendering tab) | Aurora background stops animating |

Stop dev (`Ctrl+C`).

- [ ] **Step 5:** Final commit (no-op acceptable).

```bash
git status
```

If clean, no commit needed. If any tweaks were made during smoke walk:

```bash
git add -p   # review changes
git commit -m "chore(ui): aurora redesign smoke fixes

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

- [ ] **Step 6:** Branch summary.

```bash
git log main..HEAD --oneline
```

Expected output (in order):

```
chore(ui): aurora redesign smoke fixes  # optional
chore(ui): retheme SSOLogin + NotFound + Home + ProtectedRoute
feat(ui): forgot-password stub + sessions/team/settings pages
feat(ui): redesign Login + Signup + Welcome aurora
feat(ui): aurora layouts and atoms
chore(ui): aurora base CSS + reduced-motion guard
chore(ui): simplify Button wrapper, drop brand palette mapping
chore(ui): force dark theme in Provider
chore(ui): rewrite theme to aurora dark tokens
feat(design): aurora UI redesign spec + mockup
```

---

## Spec coverage map

| Spec section | Plan task |
|--------------|-----------|
| §4.1 file tree | All tasks (file map at top of plan) |
| §4.2 theme rewrite | Task 1 |
| §4.2 forced dark provider | Task 2 |
| §4.3 AuthLayout | Task 14 |
| §4.3 AppShell | Task 15 |
| §4.4 AuroraBackground | Task 5 |
| §4.4 BrandMark | Task 6 |
| §4.4 BrandPanel | Task 9 |
| §4.4 GlassCard | Task 7 |
| §4.4 FeatureList | Task 8 |
| §4.4 PasswordField | Task 10 |
| §4.4 PasswordStrength | Task 10 |
| §4.4 StatCard | Task 12 |
| §4.4 ActivityRow | Task 12 |
| §4.4 Topbar | Task 11 |
| §4.4 UserChip | Task 11 |
| §4.5 mock data | Task 13 |
| §4.6 routing | Tasks 16, 22 |
| §4.7 Login rewrite | Task 17 |
| §4.7 Signup rewrite (drop T&C) | Task 18 |
| §4.7 Welcome rewrite | Task 19 |
| §4.7 ForgotPassword | Task 20 |
| §4.7 Sessions/Team/Settings stubs | Task 21 |
| §4.7 SSOLogin token swap | Task 23 |
| §4.7 NotFound retheme | Task 24 |
| §4.7 Home retheme | Task 25 |
| §4.8 index.css aurora keyframes + reduced-motion | Task 4 |
| §6 build/lint/smoke | Task 26 |
| §7 rollout commit per step | Tasks 1, 2, 3, 4, 15, 19, 22, 25, 26 |
| §8 risk: forcedTheme conflict with ColorModeProvider | Task 2 (resolved by `forcedTheme` prop) |
| §8 risk: Spinner color brand.500 | Task 25 |

## Out-of-scope confirmations (from spec §9)

- No light mode toggle.
- No `/sessions` `/activity` Go endpoints.
- No real forgot-password backend.
- No backend domain code changes.

---

## Implementation notes for the executor

1. **`bgGradient` token** — Earlier drafts of the spec used `bgGradient="auroraPrimary"`. Tasks 17–24 use raw `css={{ background: "linear-gradient(...)" }}` instead, because Chakra v3's `bgGradient` token resolution for `linear-gradient` strings is fragile across versions. Same for `conicLogo`, `auroraText`, `auroraStat`. Theme tokens are still defined (Task 1) for any future direct use or programmatic access.

2. **`<b>` in mock body** — `consts/mock.tsx` uses raw `<b>` for emphasis. If you prefer styled component, swap to `<Text as="span" fontWeight="semibold">` — semantically equivalent.

3. **`react-icons` already installed** (`react-icons@^5.5.0` in `package.json`). Lucide subset (`/lu`) imports work directly.

4. **No `tabIndex` issues** — IconButtons and links inside stacks default to keyboard-reachable. Keep `aria-label` on every IconButton (Tasks 10, 11, 23).

5. **CSS-in-JS via `css={{...}}` prop** — Used heavily for backdrop-filter, conic gradients, and complex mesh backgrounds. Chakra v3 passes this through Emotion. Webkit prefix (`WebkitBackdropFilter`) included where needed.

6. **`make run-ui`** is equivalent to `cd ui && npm run dev`. Either works.

7. **Branch already exists** (`feat/aurora-ui-redesign`). Plan tasks commit on this branch. No further branching.

8. **Step 1 → Step N** within each task — they are dependent. Don't reorder. Within Phase 2, atom tasks can be parallelized (subagent-driven), but BrandPanel (Task 9) needs BrandMark (Task 6) + FeatureList (Task 8); Topbar (Task 11) needs BrandMark + UserChip; AppShell (Task 15) needs Topbar.

9. **No tests written** — project has no test runner. Verification is build + lint + manual smoke. If you add Vitest later, the Login/Signup/Welcome forms are good first targets (formData state, validators integration).
