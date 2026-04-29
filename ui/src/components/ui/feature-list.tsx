import { HStack, Stack, Box, Text, Center } from "@chakra-ui/react";
import type { ReactNode } from "react";

export interface FeatureItem {
	icon: ReactNode;
	title: string;
	desc: string;
}

interface FeatureListProps {
	items: FeatureItem[];
}

export const FeatureList = ({ items }: FeatureListProps) => {
	return (
		<Stack gap="3" mt="7" position="relative" zIndex={2}>
			{items.map((item, i) => (
				<HStack
					key={i}
					gap="3.5"
					align="flex-start"
					p="3.5"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border"
					borderRadius="xl"
					css={{ backdropFilter: "blur(20px) saturate(1.2)", WebkitBackdropFilter: "blur(20px) saturate(1.2)" }}
				>
					<Center
						w="8"
						h="8"
						borderRadius="lg"
						borderWidth="1px"
						borderColor="border.strong"
						color="accentAlt"
						css={{
							background:
								"linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.25))",
						}}
						flex="0 0 auto"
					>
						{item.icon}
					</Center>
					<Box>
						<Text fontSize="sm" fontWeight="semibold" color="fg" mb="0.5">
							{item.title}
						</Text>
						<Text fontSize="xs" color="fg.muted">
							{item.desc}
						</Text>
					</Box>
				</HStack>
			))}
		</Stack>
	);
};
