"use client";

import { Center, Flex } from "@chakra-ui/react";
import { LuCheck } from "react-icons/lu";
import { APP_COLOR_KEYS, APP_COLORS } from "@/consts";

interface ColorSwatchPickerProps {
	/** The currently selected color palette key (empty = none selected). */
	value: string;
	/** Called with the palette key when a swatch is clicked. */
	onChange: (key: string) => void;
	/** When true the swatches are non-interactive (e.g. color locked to a resource). */
	disabled?: boolean;
	/** Accessible label for the radiogroup. */
	ariaLabel?: string;
}

/**
 * Reusable color-swatch picker over the shared APP_COLOR_KEYS palette. Renders
 * the ringed selected swatch with a check, matching the app_service appearance
 * editor so role + permission appearance reuse the exact same look.
 * Presentational only — the parent owns the selected value.
 */
export const ColorSwatchPicker = ({ value, onChange, disabled = false, ariaLabel = "Color" }: ColorSwatchPickerProps) => (
	<Flex
		wrap="wrap"
		gap="12px"
		p="14px"
		borderRadius="glassSm"
		borderWidth="1px"
		borderColor="border"
		css={{ background: "rgba(7,7,26,0.35)", opacity: disabled ? 0.55 : 1, pointerEvents: disabled ? "none" : "auto" }}
		role="radiogroup"
		aria-label={ariaLabel}
		aria-disabled={disabled}
	>
		{APP_COLOR_KEYS.map((key) => {
			const selected = value === key;
			const hex = APP_COLORS[key].hex;
			return (
				<Center
					key={key}
					as="button"
					role="radio"
					aria-checked={selected}
					aria-label={`${key} (${hex})`}
					title={key}
					onClick={() => onChange(key)}
					position="relative"
					w="36px"
					h="36px"
					borderRadius="full"
					cursor="pointer"
					borderWidth="2px"
					borderColor="transparent"
					css={{
						background: hex,
						boxShadow: selected
							? `0 0 0 2px #0B0B23, 0 0 0 4px ${hex}, 0 0 16px ${hex}`
							: "0 0 0 1px rgba(255,255,255,0.12) inset",
						transition: "transform .15s ease, box-shadow .15s ease",
					}}
					_hover={{ transform: "scale(1.08)" }}
				>
					{selected && <LuCheck size={16} color="white" />}
				</Center>
			);
		})}
	</Flex>
);
