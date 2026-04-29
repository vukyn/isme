"use client";

import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuMonitor } from "react-icons/lu";
import { GlassCard } from "@/components/ui/glass-card";
import { AppShell } from "@/layouts/AppShell";
import { useCurrentUser } from "@/hooks/useCurrentUser";

export const Sessions = () => {
	const { user, loading } = useCurrentUser();
	const name = user?.name || "User";
	const email = user?.email || "";

	return (
		<AppShell active="sessions" user={{ name, email }}>
			{loading ? (
				<Center py="20">
					<Spinner size="xl" color="accent" />
				</Center>
			) : (
				<GlassCard p="14">
					<Stack align="center" gap="4" textAlign="center">
						<Box color="accentAlt" fontSize="3xl">
							<LuMonitor />
						</Box>
						<Heading as="h1" fontSize="2xl" color="fg">
							Sessions
						</Heading>
						<Text color="fg.muted">Coming soon.</Text>
					</Stack>
				</GlassCard>
			)}
		</AppShell>
	);
};
