"use client";

import { Box, Center, Flex, HStack, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuActivity, LuInfo } from "react-icons/lu";
import { GlassCard } from "@/components/ui/glass-card";
import { ActivityRow } from "@/components/ui/activity-row";
import { AppShell } from "@/layouts/AppShell";
import { useUser } from "@/hooks/useUser";
import { useActivity } from "@/hooks/useActivity";
import { activityToRow } from "@/utils/activity";

// Server clamps the feed to MaxActivityLimit (constants.MaxActivityLimit = 50).
// The full-history page requests the cap; there is no offset pagination yet.
const ACTIVITY_PAGE_LIMIT = 50;

export const Activity = () => {
	const { user } = useUser();
	const name = user?.name || "User";
	const email = user?.email || "";

	const { activity, loading, error } = useActivity(ACTIVITY_PAGE_LIMIT);

	const renderBody = () => {
		if (loading) {
			return (
				<Center py="20">
					<Stack align="center" gap="4">
						<Spinner size="xl" color="accent" />
						<Text color="fg.muted">Loading...</Text>
					</Stack>
				</Center>
			);
		}

		if (error) {
			return (
				<Center py="20">
					<Text color="danger">{error.message}</Text>
				</Center>
			);
		}

		if (activity.length === 0) {
			return (
				<GlassCard>
					<Stack align="center" gap="3.5" px="6" py="14" textAlign="center">
						<Center w="64px" h="64px" borderRadius="20px" borderWidth="1px" borderColor="border.strong" bg="bg.glass" color="fg.muted">
							<LuActivity size={28} />
						</Center>
						<Heading as="h3" fontSize="17px" fontWeight="semibold" color="fg">
							No activity yet
						</Heading>
						<Text fontSize="13px" color="fg.muted" maxW="360px">
							Sign-ins, invitations, and password changes on your account will show up here.
						</Text>
					</Stack>
				</GlassCard>
			);
		}

		return (
			<Stack gap="5">
				<GlassCard p="2">
					<Stack
						gap="0"
						css={{
							"& > * + *": {
								borderTop: "1px solid var(--chakra-colors-border)",
							},
						}}
					>
						{activity.map((item) => (
							<ActivityRow key={item.id} {...activityToRow(item)} />
						))}
					</Stack>
				</GlassCard>

				<HStack
					gap="2.5"
					align="flex-start"
					px="4"
					py="3.5"
					borderRadius="glassSm"
					borderWidth="1px"
					borderColor="border"
					bg="bg.glass"
					fontSize="12px"
					color="fg.muted"
				>
					<Box color="aurora.cyan" flex="none" mt="0.5">
						<LuInfo size={15} />
					</Box>
					<Text>
						Showing your {ACTIVITY_PAGE_LIMIT} most recent events, newest first. Older activity is pruned automatically by the
						account retention policy.
					</Text>
				</HStack>
			</Stack>
		);
	};

	return (
		<AppShell active="activity" user={{ name, email }}>
			<Flex justify="space-between" align="flex-end" gap="4" wrap="wrap">
				<Box>
					<Heading as="h1" fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1">
						Activity
					</Heading>
					<Text mt="1.5" color="fg.muted" fontSize="sm">
						Recent security and account events on your account
					</Text>
				</Box>
			</Flex>

			{renderBody()}
		</AppShell>
	);
};
