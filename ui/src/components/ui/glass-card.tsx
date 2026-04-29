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
