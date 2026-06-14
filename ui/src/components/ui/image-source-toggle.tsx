"use client";

import { Box, Flex } from "@chakra-ui/react";
import { LuLink, LuUpload } from "react-icons/lu";

// Where the avatar image comes from. File upload is the default; paste mode
// accepts an external image URL stored directly as the avatar link.
export type ImageSource = "file" | "link";

interface ImageSourceToggleProps {
	value: ImageSource;
	onChange: (next: ImageSource) => void;
}

// Segmented "Upload file" / "Paste link" control — ports the mock's `.seg`
// (aurora glass active pill). Mirrors rainy's ImageSourceToggle pattern.
export const ImageSourceToggle = ({ value, onChange }: ImageSourceToggleProps) => {
	const options = [
		{ value: "file" as const, label: "Upload file", icon: <LuUpload size={14} /> },
		{ value: "link" as const, label: "Paste link", icon: <LuLink size={14} /> },
	];
	return (
		<Flex
			w="fit-content"
			gap="1"
			p="1"
			borderRadius="glassSm"
			bg="rgba(7,7,26,0.45)"
			borderWidth="1px"
			borderColor="border.strong"
		>
			{options.map((option) => {
				const active = value === option.value;
				return (
					<Box
						key={option.value}
						as="button"
						display="inline-flex"
						alignItems="center"
						gap="7px"
						h="34px"
						px="4"
						borderRadius="9px"
						fontSize="13px"
						fontWeight="semibold"
						cursor="pointer"
						color={active ? "fg" : "fg.muted"}
						bg={active ? "bg.glassHi" : "transparent"}
						borderWidth="1px"
						borderColor={active ? "border.strong" : "transparent"}
						boxShadow={active ? "0 0 16px rgba(99,102,241,0.20)" : "none"}
						_hover={active ? undefined : { color: "fg.subtle" }}
						aria-selected={active}
						role="tab"
						onClick={() => onChange(option.value)}
					>
						{option.icon}
						{option.label}
					</Box>
				);
			})}
		</Flex>
	);
};
