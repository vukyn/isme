"use client";

import { Box } from "@chakra-ui/react";
import type { ReactNode } from "react";
import { LuAppWindow, LuCloud, LuLayers, LuMusic } from "react-icons/lu";
import { renderIcon, resolveAppColor } from "@/consts";

/**
 * Shared, presentational app identity tile. Renders the aurora tinted-glass tile
 * (chosen color tint + bordered glass + colored glyph) from a stored icon/color
 * key pair — the look ported from aurora-app-service-edit-design.html.
 *
 * FALLBACK (so rows that have never been styled render unchanged): when colorKey
 * is empty, the tile falls back to a hashed gradient (list/detail surfaces) or
 * the APP_META identity (chip surface), matching the pre-feature behaviour.
 * When iconKey is empty, a neutral default glyph is shown.
 *
 * Terminated/inactive visual states are preserved via the `status` prop, mirroring
 * the old local AppTile in AppServices.tsx.
 */

/** Maps to constants.AppServiceStatus* (1=active, 2=inactive, 3=terminated). */
type AppTileStatus = 1 | 2 | 3;

const STATUS_TERMINATED = 3;
const STATUS_INACTIVE = 2;

/** Pixel sizes for the three surfaces the tile appears on. */
const SIZE_PX: Record<NonNullable<AppTileProps["size"]>, { box: string; radius: string; icon: number }> = {
	chip: { box: "22px", radius: "7px", icon: 12 },
	list: { box: "34px", radius: "10px", icon: 16 },
	detail: { box: "64px", radius: "18px", icon: 30 },
};

/** Hashed-gradient fallback — identical list + hash to the old AppServices.tsx tileGradient. */
const TILE_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
];

const hashGradient = (seed: string): string => {
	let hash = 0;
	for (const char of seed) hash = (hash * 31 + char.charCodeAt(0)) % 997;
	return TILE_GRADIENTS[hash % TILE_GRADIENTS.length];
};

/** Known-app chip fallback icon (matches AppRoleChip APP_META) by app_code. */
const APP_META_ICON: Record<string, (size: number) => ReactNode> = {
	isme: (size) => <LuLayers size={size} />,
	medioa2: (size) => <LuCloud size={size} />,
	rainy: (size) => <LuMusic size={size} />,
};

export interface AppTileProps {
	/** Stored appearance icon key; empty falls back to the neutral default. */
	iconKey?: string;
	/** Stored appearance color palette key; empty triggers the fallback look. */
	colorKey?: string;
	/** Which surface to size for. */
	size: "chip" | "list" | "detail";
	/** Override the tile border radius. */
	radius?: string;
	/** App service status — drives terminated/inactive visuals. */
	status?: AppTileStatus;
	/** Seed for the hashed-gradient fallback (typically the app id). */
	fallbackSeed?: string;
	/** App code, used to pick the chip identity fallback icon. */
	appCode?: string;
}

export const AppTile = ({ iconKey, colorKey, size, radius, status, fallbackSeed, appCode }: AppTileProps) => {
	const dim = SIZE_PX[size];
	const borderRadius = radius ?? dim.radius;

	// Terminated: dashed neutral tile (no icon/color expressed).
	if (status === STATUS_TERMINATED) {
		return (
			<Box
				as="span"
				display="grid"
				placeItems="center"
				w={dim.box}
				h={dim.box}
				flex="none"
				borderRadius={borderRadius}
				bg="bg.glass"
				borderWidth="1.5px"
				borderStyle="dashed"
				borderColor="border.strong"
				color="fg.muted"
			>
				<LuAppWindow size={dim.icon} />
			</Box>
		);
	}

	const inactive = status === STATUS_INACTIVE;

	// Stored-appearance path: tinted-glass tile from the color key.
	if (colorKey) {
		const color = resolveAppColor(colorKey);
		return (
			<Box
				as="span"
				display="grid"
				placeItems="center"
				w={dim.box}
				h={dim.box}
				flex="none"
				borderRadius={borderRadius}
				color={inactive ? "fg.subtle" : color.hex}
				css={{
					background: inactive ? "linear-gradient(135deg, #3A3F6E, #23264A)" : `rgba(${color.rgb}, 0.16)`,
					border: inactive ? "none" : `1px solid rgba(${color.rgb}, 0.42)`,
					boxShadow: inactive ? "none" : `0 0 18px rgba(${color.rgb}, 0.20)`,
				}}
			>
				{renderIcon(iconKey, dim.icon)}
			</Box>
		);
	}

	// Fallback: chip surface uses the app_code identity icon; list/detail use a
	// hashed gradient seeded by the id — both preserve the pre-feature look.
	const chipIcon = appCode && APP_META_ICON[appCode];
	const glyph = chipIcon ? chipIcon(dim.icon) : renderIcon(iconKey, dim.icon);
	return (
		<Box
			as="span"
			display="grid"
			placeItems="center"
			w={dim.box}
			h={dim.box}
			flex="none"
			borderRadius={borderRadius}
			color={inactive ? "fg.subtle" : "white"}
			css={{
				background: inactive ? "linear-gradient(135deg, #3A3F6E, #23264A)" : hashGradient(fallbackSeed ?? appCode ?? ""),
				boxShadow: inactive ? "none" : "0 0 14px rgba(99,102,241,0.30)",
			}}
		>
			{glyph}
		</Box>
	);
};
