"use client";

import { useEffect, useState } from "react";
import { Box, Button, Center, Flex, Grid, HStack, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuFileText, LuPlus } from "react-icons/lu";
import { getCurrentUser } from "@/apis/auth";
import { GlassCard } from "@/components/ui/glass-card";
import { StatCard } from "@/components/ui/stat-card";
import { ActivityRow } from "@/components/ui/activity-row";
import { AppShell } from "@/layouts/AppShell";
import { MOCK_STATS, MOCK_ACTIVITY } from "@/consts/mock";
import type { GetMeResponse } from "@/types";

export const Welcome = () => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		const fetchUser = async () => {
			try {
				setLoading(true);
				setError(null);
				const response = await getCurrentUser();
				setUser((response as { data: GetMeResponse }).data);
			} catch (err: unknown) {
				const msg = err instanceof Error ? err.message : "Failed to load user data";
				setError(msg);
			} finally {
				setLoading(false);
			}
		};
		fetchUser();
	}, []);

	if (loading) {
		return (
			<Center w="full" h="100vh">
				<Stack align="center" gap="4">
					<Spinner size="xl" color="accent" />
					<Text color="fg.muted">Loading...</Text>
				</Stack>
			</Center>
		);
	}

	if (error) {
		return (
			<Center w="full" h="100vh">
				<Text color="danger">{error}</Text>
			</Center>
		);
	}

	const name = user?.name || "User";
	const email = user?.email || "";

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
							css={{
								background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
								backgroundSize: "200% 200%",
							}}
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
					<a href="#" style={{ fontSize: "var(--chakra-font-sizes-sm)", color: "var(--chakra-colors-fg-subtle)", textDecoration: "none" }}>
						View all →
					</a>
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
