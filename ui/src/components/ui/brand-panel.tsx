import { Box, Flex, HStack, Text, Heading } from "@chakra-ui/react";
import { BrandMark } from "./brand-mark";
import { FeatureList, type FeatureItem } from "./feature-list";

interface BrandPanelProps {
	pill: string;
	pillTone?: "violet" | "cyan" | "mint";
	titleLead: string;
	titleGrad: string;
	sub: string;
	features: FeatureItem[];
	footerLeft?: string;
	footerRight?: string;
}

const pillToneColor: Record<NonNullable<BrandPanelProps["pillTone"]>, string> = {
	violet: "#FBBF24",
	cyan: "#22D3EE",
	mint: "#34D399",
};

export const BrandPanel = ({
	pill,
	pillTone = "violet",
	titleLead,
	titleGrad,
	sub,
	features,
	footerLeft = "© isme 2026",
	footerRight = "v1.0 · main",
}: BrandPanelProps) => {
	const dot = pillToneColor[pillTone];

	return (
		<Flex
			position="relative"
			direction="column"
			justify="space-between"
			p="11"
			borderRightWidth={{ base: 0, md: "1px" }}
			borderBottomWidth={{ base: "1px", md: 0 }}
			borderColor="border"
			minH={{ base: "320px", md: "auto" }}
			overflow="hidden"
		>
			<Box position="absolute" w="320px" h="320px" right="-80px" top="-80px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(40px)", background: "radial-gradient(circle, rgba(99,102,241,0.6), transparent 70%)" }} />
			<Box position="absolute" w="260px" h="260px" left="-60px" bottom="40px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(40px)", background: "radial-gradient(circle, rgba(236,72,153,0.55), transparent 70%)" }} />
			<Box position="absolute" w="220px" h="220px" right="80px" bottom="-60px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(40px)", background: "radial-gradient(circle, rgba(34,211,238,0.45), transparent 70%)" }} />

			<HStack gap="3" position="relative" zIndex={2}>
				<BrandMark />
				<Text fontWeight="bold" fontSize="lg" letterSpacing="-0.01em" color="fg">
					isme
				</Text>
			</HStack>

			<Box position="relative" zIndex={2} maxW="480px">
				<HStack
					display="inline-flex"
					gap="2"
					px="3"
					py="1.5"
					borderRadius="full"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border.strong"
					css={{ backdropFilter: "blur(20px)", WebkitBackdropFilter: "blur(20px)" }}
				>
					<Box w="1.5" h="1.5" borderRadius="full" css={{ background: dot, boxShadow: `0 0 10px ${dot}` }} />
					<Text fontSize="xs" fontWeight="medium" color="fg.subtle">
						{pill}
					</Text>
				</HStack>
				<Heading
					as="h1"
					mt="4.5"
					mb="3.5"
					fontSize={{ base: "30px", md: "44px" }}
					lineHeight="1.05"
					letterSpacing="-0.025em"
					color="fg"
				>
					{titleLead}{" "}
					<Text
						as="span"
						css={{
							backgroundImage:
								"linear-gradient(135deg, #22D3EE 0%, #6366F1 40%, #8B5CF6 70%, #EC4899 100%)",
							backgroundSize: "200% 100%",
							WebkitBackgroundClip: "text",
							backgroundClip: "text",
							color: "transparent",
							animation: "auroraFlow 8s ease-in-out infinite alternate",
						}}
					>
						{titleGrad}
					</Text>
				</Heading>
				<Text color="fg.muted" fontSize="md" maxW="440px">
					{sub}
				</Text>
				<FeatureList items={features} />
			</Box>

			<HStack justify="space-between" position="relative" zIndex={2} color="fg.muted" fontSize="xs">
				<Text>{footerLeft}</Text>
				<Text>{footerRight}</Text>
			</HStack>
		</Flex>
	);
};
