import { Center } from "@chakra-ui/react";

interface BrandMarkProps {
	// "lg" matches the 52px SSO handshake app-tile (.app-tile.isme in the mock)
	size?: "sm" | "md" | "lg";
}

// Explicit px — Chakra's size scale has no "13" token (and skips other odd
// values), so token strings like "13" are silently dropped, collapsing the
// tile to its content size. Pixel values render deterministically.
//
// Geometry mirrors demo/aurora-favicon.svg (the canonical isme logo): a single
// linear aurora-gradient rounded tile carrying the "i / person" monogram (white
// dot + rounded stem = identity / "is me"). The favicon mark lives on a 32-unit
// viewBox, so the svg renders at the full tile size to keep its proportions.
const SIZES = {
	sm: { outer: "32px", radii: "9px" },
	md: { outer: "40px", radii: "10px" },
	lg: { outer: "52px", radii: "15px" },
} as const;

export const BrandMark = ({ size = "md" }: BrandMarkProps) => {
	const { outer, radii } = SIZES[size];

	return (
		<Center
			w={outer}
			h={outer}
			borderRadius={radii}
			boxShadow="0 8px 24px rgba(99,102,241,0.45)"
			css={{
				background: "linear-gradient(135deg, #22D3EE 0%, #8B5CF6 60%, #EC4899 100%)",
			}}
		>
			{/* "i / person" monogram — favicon viewBox (0 0 32 32), white fill */}
			<svg width={outer} height={outer} viewBox="0 0 32 32" fill="none" role="img" aria-label="isme">
				<circle cx="16" cy="10" r="3.6" fill="#FFFFFF" />
				<rect x="12.4" y="16.4" width="7.2" height="9.6" rx="3.6" fill="#FFFFFF" />
			</svg>
		</Center>
	);
};
