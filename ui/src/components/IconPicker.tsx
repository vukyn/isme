"use client";

import { Center, Flex } from "@chakra-ui/react";
import { ICON_KEYS, renderIcon } from "@/consts";

interface IconPickerProps {
	/** The currently selected icon key (empty = none selected). */
	value: string;
	/** Called with the icon key when a swatch is clicked. */
	onChange: (key: string) => void;
	/** When true the grid is non-interactive (e.g. icon locked to a resource). */
	disabled?: boolean;
	/** Accessible label for the radiogroup. */
	ariaLabel?: string;
}

/**
 * Reusable icon-grid picker over the shared ICON_KEYS allowlist. Renders each
 * key with renderIcon and the aurora violet selected-ring. Extracted from the
 * app_service appearance editor so role + permission appearance reuse the exact
 * same look. Presentational only — the parent owns the selected value.
 */
export const IconPicker = ({ value, onChange, disabled = false, ariaLabel = "Icon" }: IconPickerProps) => (
	<Flex
		wrap="wrap"
		gap="8px"
		p="12px"
		borderRadius="glassSm"
		borderWidth="1px"
		borderColor="border"
		css={{ background: "rgba(7,7,26,0.35)", opacity: disabled ? 0.55 : 1, pointerEvents: disabled ? "none" : "auto" }}
		role="radiogroup"
		aria-label={ariaLabel}
		aria-disabled={disabled}
	>
		{ICON_KEYS.map((key) => {
			const selected = value === key;
			return (
				<Center
					key={key}
					as="button"
					role="radio"
					aria-checked={selected}
					aria-label={key}
					title={key}
					onClick={() => onChange(key)}
					w="38px"
					h="38px"
					borderRadius="9px"
					cursor="pointer"
					borderWidth="1px"
					borderColor={selected ? "aurora.violet" : "border.strong"}
					color={selected ? "aurora.violet" : "fg.subtle"}
					css={{
						background: selected
							? "linear-gradient(135deg, rgba(139,92,246,0.22), rgba(99,102,241,0.14))"
							: "rgba(255,255,255,0.06)",
						boxShadow: selected ? "0 0 0 3px rgba(139,92,246,0.25), 0 0 16px rgba(139,92,246,0.20) inset" : "none",
					}}
					_hover={{ color: "fg", borderColor: "rgba(255,255,255,0.30)" }}
				>
					{renderIcon(key, 17)}
				</Center>
			);
		})}
	</Flex>
);
