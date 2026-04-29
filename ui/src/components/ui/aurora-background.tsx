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
