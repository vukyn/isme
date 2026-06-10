"use client";

import { Box, HStack, Text } from "@chakra-ui/react";
import type { ReactNode } from "react";
import { LuLayers, LuCloud, LuMusic } from "react-icons/lu";
import { renderIcon, resolveAppColor } from "@/consts";

/**
 * Per-app role chip — `app:role` — reused across the user list, the invite
 * acceptance summary, and the invite assignment preview. RBAC is app-owned,
 * so every role belongs to one app_service; this chip surfaces that pairing.
 *
 * Colour is keyed off the OWNING APP (aurora theme): isme = violet/conic,
 * medioa2 = violet-tinted, rainy = cyan, anything else = neutral glass.
 */

interface AppMeta {
	tileBg: string;
	tileColor: string;
	border: string;
	bg: string;
	color: string;
	icon: ReactNode;
}

const FALLBACK: AppMeta = {
	tileBg: "rgba(255,255,255,0.08)",
	tileColor: "fg.subtle",
	border: "border.strong",
	bg: "bg.glass",
	color: "fg.subtle",
	icon: <LuLayers size={11} />,
};

/** Known platform apps get a fixed identity; unknown apps fall back to neutral. */
const APP_META: Record<string, AppMeta> = {
	isme: {
		tileBg: "conic-gradient(from 180deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
		tileColor: "white",
		border: "rgba(139,92,246,0.40)",
		bg: "rgba(139,92,246,0.12)",
		color: "aurora.violet",
		icon: <LuLayers size={11} />,
	},
	medioa2: {
		tileBg: "rgba(139,92,246,0.18)",
		tileColor: "aurora.violet",
		border: "rgba(34,211,238,0.40)",
		bg: "rgba(34,211,238,0.10)",
		color: "aurora.cyan",
		icon: <LuCloud size={11} />,
	},
	rainy: {
		tileBg: "rgba(34,211,238,0.14)",
		tileColor: "aurora.cyan",
		border: "rgba(52,211,153,0.35)",
		bg: "rgba(52,211,153,0.10)",
		color: "aurora.mint",
		icon: <LuMusic size={11} />,
	},
};

const appMeta = (appCode: string): AppMeta => APP_META[appCode] ?? FALLBACK;

interface AppRoleChipProps {
	appCode: string;
	role: string;
	/** Optional permission count suffix (e.g. "· 9 perms"). */
	permCount?: number;
	title?: string;
}

export const AppRoleChip = ({ appCode, role, permCount, title }: AppRoleChipProps) => {
	const meta = appMeta(appCode);
	return (
		<HStack
			as="span"
			display="inline-flex"
			gap="6px"
			px="9px"
			py="3px"
			borderRadius="full"
			fontSize="12px"
			fontWeight="semibold"
			borderWidth="1px"
			borderColor={meta.border}
			bg={meta.bg}
			color={meta.color}
			whiteSpace="nowrap"
			title={title ?? `${appCode} · ${role}`}
		>
			<Text as="span" color="fg.muted" fontWeight="medium">
				{appCode}:
			</Text>
			{role}
			{permCount !== undefined && (
				<Text as="span" color="fg.muted" fontWeight="medium">
					· {permCount} perm{permCount === 1 ? "" : "s"}
				</Text>
			)}
		</HStack>
	);
};

interface AppTileProps {
	appCode: string;
	size?: string;
	radius?: string;
	/** Stored appearance icon key — when set (with colorKey), overrides APP_META. */
	iconKey?: string;
	/** Stored appearance color palette key — when set, renders the tinted-glass tile. */
	colorKey?: string;
}

/** Square app identity tile (gradient/tinted) reused in chips, cards, rows. */
export const AppTile = ({ appCode, size = "22px", radius = "7px", iconKey, colorKey }: AppTileProps) => {
	// Stored-appearance path: tinted-glass tile from the color key (matches the
	// shared AppTile look). Falls back to APP_META when no color is stored, so
	// existing invite/user-list callers (which pass no keys) are unchanged.
	if (colorKey) {
		const color = resolveAppColor(colorKey);
		return (
			<Box
				as="span"
				display="grid"
				placeItems="center"
				w={size}
				h={size}
				flex="none"
				borderRadius={radius}
				color={color.hex}
				css={{
					background: `rgba(${color.rgb}, 0.16)`,
					border: `1px solid rgba(${color.rgb}, 0.42)`,
				}}
			>
				{renderIcon(iconKey, 12)}
			</Box>
		);
	}

	const meta = appMeta(appCode);
	return (
		<Box
			as="span"
			display="grid"
			placeItems="center"
			w={size}
			h={size}
			flex="none"
			borderRadius={radius}
			color={meta.tileColor}
			css={{ background: meta.tileBg }}
		>
			{meta.icon}
		</Box>
	);
};
