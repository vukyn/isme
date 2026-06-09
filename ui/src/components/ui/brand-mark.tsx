import { Center } from "@chakra-ui/react";

interface BrandMarkProps {
	// "lg" matches the 52px SSO handshake app-tile (.app-tile.isme in the mock)
	size?: "sm" | "md" | "lg";
}

// Explicit px — Chakra's size scale has no "13" token (and skips other odd
// values), so token strings like "13" are silently dropped, collapsing the
// tile to its content size. Pixel values render deterministically.
const SIZES = {
	sm: { outer: "32px", inner: "24px", radii: "9px", innerRadii: "7px", icon: 14 },
	md: { outer: "40px", inner: "32px", radii: "10px", innerRadii: "9px", icon: 16 },
	lg: { outer: "52px", inner: "40px", radii: "15px", innerRadii: "11px", icon: 22 },
} as const;

export const BrandMark = ({ size = "md" }: BrandMarkProps) => {
	const { outer, inner, radii, innerRadii, icon: iconSize } = SIZES[size];

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
					strokeWidth={2.2}
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
