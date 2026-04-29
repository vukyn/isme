"use client";

import { useEffect, useState } from "react";
import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuUsers } from "react-icons/lu";
import { getCurrentUser } from "@/apis/auth";
import { GlassCard } from "@/components/ui/glass-card";
import { AppShell } from "@/layouts/AppShell";
import type { GetMeResponse } from "@/types";

export const Team = () => {
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		getCurrentUser()
			.then((r) => setUser((r as { data: GetMeResponse }).data))
			.finally(() => setLoading(false));
	}, []);

	if (loading) {
		return (
			<Center w="full" h="100vh">
				<Spinner size="xl" color="accent" />
			</Center>
		);
	}

	return (
		<AppShell active="team" user={{ name: user?.name || "User", email: user?.email || "" }}>
			<GlassCard p="14">
				<Stack align="center" gap="4" textAlign="center">
					<Box color="accentAlt" fontSize="3xl"><LuUsers /></Box>
					<Heading as="h1" fontSize="2xl" color="fg">Team</Heading>
					<Text color="fg.muted">Coming soon.</Text>
				</Stack>
			</GlassCard>
		</AppShell>
	);
};
