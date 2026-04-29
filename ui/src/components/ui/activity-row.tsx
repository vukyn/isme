import { Box, Center, Grid, Text } from "@chakra-ui/react";
import type { ReactNode } from "react";

export type ActivityTone = "ok" | "violet" | "magenta";

interface ActivityRowProps {
	tone: ActivityTone;
	icon: ReactNode;
	body: ReactNode;
	time: string;
}

const toneStyles: Record<ActivityTone, { bg: string; color: string }> = {
	ok: { bg: "rgba(52,211,153,0.15)", color: "#34D399" },
	violet: { bg: "rgba(139,92,246,0.18)", color: "#8B5CF6" },
	magenta: { bg: "rgba(236,72,153,0.18)", color: "#EC4899" },
};

export const ActivityRow = ({ tone, icon, body, time }: ActivityRowProps) => {
	const t = toneStyles[tone];
	return (
		<Grid
			gridTemplateColumns="36px 1fr auto"
			alignItems="center"
			gap="3.5"
			px="3.5"
			py="3"
			borderRadius="lg"
			_hover={{ bg: "rgba(255,255,255,0.04)" }}
		>
			<Center w="9" h="9" borderRadius="lg" css={{ background: t.bg, color: t.color }}>
				{icon}
			</Center>
			<Box fontSize="sm" color="fg">
				{body}
			</Box>
			<Text fontSize="xs" color="fg.muted">
				{time}
			</Text>
		</Grid>
	);
};
