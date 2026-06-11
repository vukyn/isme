"use client";

import { Box, Center, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { RotationCleanupCard } from "@/components/RotationCleanupCard";
import { SessionAutoRevokeCard } from "@/components/SessionAutoRevokeCard";
import { AppShell } from "@/layouts/AppShell";
import { useUser } from "@/hooks/useUser";

export const Settings = () => {
	const { user, loading } = useUser();
	const name = user?.name || "User";
	const email = user?.email || "";

	return (
		<AppShell active="settings" user={{ name, email }}>
			{loading ? (
				<Center py="20">
					<Spinner size="xl" color="accent" />
				</Center>
			) : (
				<Stack gap="5" maxW="880px" w="full" mx="auto">
					<Box>
						<Heading as="h1" fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1" color="fg">
							Settings
						</Heading>
						<Text mt="6px" color="fg.muted" fontSize="14px">
							Platform-wide configuration for the isme SSO service.
						</Text>
					</Box>

					<Stack gap="5">
						<SessionAutoRevokeCard />
						<RotationCleanupCard />
					</Stack>
				</Stack>
			)}
		</AppShell>
	);
};
