import { Box, Center, Heading, HStack, Text } from "@chakra-ui/react";
import type { ReactNode } from "react";

export type StatTone = "cyan" | "violet" | "magenta";

interface StatCardProps {
	icon: ReactNode;
	title: string;
	desc: string;
	stat: string;
	delta: string;
	tone?: StatTone;
}

const toneColor: Record<StatTone, string> = {
	cyan: "#22D3EE",
	violet: "#8B5CF6",
	magenta: "#EC4899",
};

export const StatCard = ({ icon, title, desc, stat, delta, tone = "cyan" }: StatCardProps) => {
	return (
		<Box
			p="5"
			bg="bg.subtle"
			borderWidth="1px"
			borderColor="border"
			borderRadius="2xl"
			boxShadow="glassSoft"
			position="relative"
		>
			<Center
				w="9"
				h="9"
				borderRadius="lg"
				borderWidth="1px"
				borderColor="border.strong"
				color={toneColor[tone]}
				mb="3"
				css={{
					background:
						"linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))",
				}}
			>
				{icon}
			</Center>
			<Text fontSize="md" fontWeight="semibold" color="fg" mb="0.5">
				{title}
			</Text>
			<Text fontSize="sm" color="fg.muted" mb="3.5">
				{desc}
			</Text>
			<Heading
				as="div"
				fontSize="28px"
				fontWeight="bold"
				letterSpacing="-0.02em"
				lineHeight="1.1"
				css={{
					backgroundImage: "linear-gradient(135deg, #22D3EE, #8B5CF6)",
					WebkitBackgroundClip: "text",
					backgroundClip: "text",
					color: "transparent",
				}}
			>
				{stat}
			</Heading>
			<HStack gap="1" mt="1" fontSize="xs" color="success" fontWeight="medium">
				<Text as="span">{delta}</Text>
			</HStack>
		</Box>
	);
};
