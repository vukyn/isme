"use client";

import { Box, Button, Center, Flex, Grid, HStack, Heading, Link, Spinner, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { LuFileText, LuPlus } from "react-icons/lu";
import { GlassCard } from "@/components/ui/glass-card";
import { StatCard } from "@/components/ui/stat-card";
import { ActivityRow } from "@/components/ui/activity-row";
import { AppShell } from "@/layouts/AppShell";
import { MOCK_STATS, MOCK_ACTIVITY } from "@/consts/mock";
import { useCurrentUser } from "@/hooks/useCurrentUser";
import { AURORA_CTA_STYLE } from "@/consts/styles";

export const Welcome = () => {
	const { user, loading, error } = useCurrentUser();
	const name = user?.name || "User";
	const email = user?.email || "";

	if (error) {
		return (
			<AppShell active="overview" user={{ name, email }}>
				<Center py="20">
					<Text color="danger">{error}</Text>
				</Center>
			</AppShell>
		);
	}

	if (loading) {
		return (
			<AppShell active="overview" user={{ name, email }}>
				<Center py="20">
					<Stack align="center" gap="4">
						<Spinner size="xl" color="accent" />
						<Text color="fg.muted">Loading...</Text>
					</Stack>
				</Center>
			</AppShell>
		);
	}

	return (
		<AppShell active="overview" user={{ name, email }}>
			<GlassCard p="8" position="relative" overflow="hidden">
				<Box position="absolute" w="300px" h="300px" right="-60px" top="-100px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(50px)", background: "radial-gradient(circle, rgba(99,102,241,0.55), transparent 70%)" }} />
				<Box position="absolute" w="240px" h="240px" right="30%" bottom="-120px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(50px)", background: "radial-gradient(circle, rgba(236,72,153,0.45), transparent 70%)" }} />
				<Box position="absolute" w="200px" h="200px" left="10%" top="-50px" borderRadius="full" pointerEvents="none" css={{ filter: "blur(50px)", background: "radial-gradient(circle, rgba(34,211,238,0.40), transparent 70%)" }} />
				<Box position="relative" zIndex={1}>
					<HStack
						display="inline-flex"
						gap="2"
						px="3"
						py="1.5"
						borderRadius="full"
						bg="bg.muted"
						color="fg"
					>
						<Box w="2" h="2" borderRadius="full" bg="success" css={{ boxShadow: "0 0 0 4px rgba(22,163,74,0.18)" }} />
						<Text fontSize="xs" fontWeight="medium">You're signed in</Text>
					</HStack>
					<Heading
						as="h1"
						mt="3.5"
						mb="2"
						fontSize={{ base: "28px", md: "40px" }}
						lineHeight="1.1"
						letterSpacing="-0.025em"
					>
						Welcome back,{" "}
						<Text
							as="span"
							css={{
								backgroundImage: "linear-gradient(135deg, #22D3EE, #8B5CF6 60%, #EC4899)",
								WebkitBackgroundClip: "text",
								backgroundClip: "text",
								color: "transparent",
							}}
						>
							{name}
						</Text>
					</Heading>
					<Text color="fg.subtle" mb="6" maxW="560px">
						Last sign-in just now from this device. Pick up where you left off, or spin up something new.
					</Text>
					<HStack gap="2.5" flexWrap="wrap">
						<Button
							h="12"
							px="5.5"
							color="white"
							borderRadius="glassSm"
							boxShadow="ctaGlow"
							_hover={{ boxShadow: "ctaGlowHi" }}
							_focusVisible={{ boxShadow: "focusRing" }}
							css={AURORA_CTA_STYLE}
						>
							<HStack gap="2.5"><LuPlus /><Text>Start a project</Text></HStack>
						</Button>
						<Button
							h="12"
							px="5"
							variant="outline"
							color="fg"
							borderColor="border.strong"
							bg="bg.glass"
							borderRadius="glassSm"
							css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
							_hover={{ bg: "bg.glassHi" }}
						>
							<HStack gap="2.5"><LuFileText /><Text>View docs</Text></HStack>
						</Button>
					</HStack>
				</Box>
			</GlassCard>

			<Grid templateColumns={{ base: "1fr", lg: "repeat(3, 1fr)" }} gap="4">
				{MOCK_STATS.map((s, i) => (
					<StatCard key={i} {...s} />
				))}
			</Grid>

			<Box>
				<Flex justify="space-between" align="baseline" mb="3">
					<Heading as="h2" fontSize="lg" letterSpacing="-0.01em" color="fg">
						Recent activity
					</Heading>
					<Link asChild fontSize="sm" color="fg.subtle" _hover={{ color: "accentAlt" }}>
						<RouterLink to="/sessions">View all →</RouterLink>
					</Link>
				</Flex>
				<GlassCard p="2">
					<Stack
						gap="0"
						css={{
							"& > * + *": {
								borderTop: "1px solid var(--chakra-colors-border)",
							},
						}}
					>
						{MOCK_ACTIVITY.map((a, i) => (
							<ActivityRow key={i} {...a} />
						))}
					</Stack>
				</GlassCard>
			</Box>
		</AppShell>
	);
};
