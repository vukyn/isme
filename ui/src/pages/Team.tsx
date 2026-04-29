"use client";

import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuUsers } from "react-icons/lu";
import { GlassCard } from "@/components/ui/glass-card";
import { AppShell } from "@/layouts/AppShell";
import { useCurrentUser } from "@/hooks/useCurrentUser";

export const Team = () => {
	const { user, loading } = useCurrentUser();
	const name = user?.name || "User";
	const email = user?.email || "";

	return (
		<AppShell active="team" user={{ name, email }}>
			{loading ? (
				<Center py="20">
					<Spinner size="xl" color="accent" />
				</Center>
			) : (
				<GlassCard p="14">
					<Stack align="center" gap="4" textAlign="center">
						<Box color="accentAlt" fontSize="3xl">
							<LuUsers />
						</Box>
						<Heading as="h1" fontSize="2xl" color="fg">
							Team
						</Heading>
						<Text color="fg.muted">Coming soon.</Text>
					</Stack>
				</GlassCard>
			)}
		</AppShell>
	);
};
