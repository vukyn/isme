import { Center } from "@chakra-ui/react";

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
