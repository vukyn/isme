"use client";

import { useEffect, useState, type ReactNode } from "react";
import { Box, Center, Flex, Grid, HStack, Heading, Link, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuCheck, LuKey, LuLogOut, LuUserPlus } from "react-icons/lu";
import { Link as RouterLink } from "react-router-dom";
import { GlassCard } from "@/components/ui/glass-card";
import { StatCard } from "@/components/ui/stat-card";
import { ActivityRow, type ActivityTone } from "@/components/ui/activity-row";
import { AppShell } from "@/layouts/AppShell";
import { MOCK_STATS, type StatEntry } from "@/consts/mock";
import type { ActivityItem } from "@/types";
import { useUser } from "@/hooks/useUser";
import { useActivity } from "@/hooks/useActivity";
import { getMySessionsCount } from "@/apis";
import { formatRelative } from "@/utils/time";

interface ActivityRowData {
	tone: ActivityTone;
	icon: ReactNode;
	body: ReactNode;
	time: string;
}

/** Maps a structured activity record to its display row. The server sends only
 *  {type, meta}; the copy/icon/tone are composed here from the type. */
const activityToRow = (item: ActivityItem): ActivityRowData => {
	const time = item.created_at ? formatRelative(item.created_at) : "";
	switch (item.type) {
		case "sign_in": {
			const device = typeof item.meta.device === "string" ? item.meta.device : "this device";
			const clientIp = typeof item.meta.client_ip === "string" ? item.meta.client_ip : "";
			return {
				tone: "ok",
				icon: <LuCheck />,
				body: (
					<>
						Sign-in from <b>{device}</b>
						{clientIp ? ` · ${clientIp}` : ""}
					</>
				),
				time,
			};
		}
		case "invitation_sent": {
			const email = typeof item.meta.email === "string" ? item.meta.email : "someone";
			const roles = Array.isArray(item.meta.roles) ? (item.meta.roles as unknown[]).filter((r): r is string => typeof r === "string") : [];
			return {
				tone: "magenta",
				icon: <LuUserPlus />,
				body: (
					<>
						Invited <b>{email}</b>
						{roles.length > 0 ? ` as ${roles.join(", ")}` : ""}
					</>
				),
				time,
			};
		}
		case "password_changed":
			return { tone: "violet", icon: <LuKey />, body: <>Password changed</>, time };
		case "sign_out":
			return { tone: "violet", icon: <LuLogOut />, body: <>Signed out</>, time };
		default:
			return { tone: "violet", icon: <LuCheck />, body: <>{item.type}</>, time };
	}
};

export const Welcome = () => {
	const { user, loading, error } = useUser();
	const { activity, loading: activityLoading, error: activityError } = useActivity();
	const name = user?.name || "User";
	const email = user?.email || "";

	// Live stats from a single /sessions/count response: card #0 "Active sessions"
	// (count + 24h delta) and card #1 "Token rotations" (sliding-24h rotation count
	// + relative last-refreshed delta). Card #2 "Member since" is derived from the
	// signed-in user's account-creation date (below).
	const [liveStats, setLiveStats] = useState<Record<number, { stat: string; delta: string }> | null>(null);
	useEffect(() => {
		getMySessionsCount()
			.then((count) =>
				setLiveStats({
					0: {
						stat: String(count.count),
						delta: `▲ +${count.new_in_24h} new in 24h`,
					},
					1: {
						stat: String(count.rotations_24h),
						delta: count.last_refreshed_at
							? `↻ last refreshed ${formatRelative(count.last_refreshed_at)}`
							: "↻ no refreshes yet",
					},
				})
			)
			.catch(() => setLiveStats(null));
	}, []);

	// "now" captured once at mount (lazy init) so the account-age math stays pure
	// across re-renders — no impure Date read during render.
	const [mountedAt] = useState(() => Date.now());
	// Card #2 "Member since" — month/year joined + account age in days.
	const memberSince = user?.created_at
		? (() => {
				const created = new Date(user.created_at);
				const monthYear = created.toLocaleDateString("en-US", { month: "short", year: "numeric" });
				const days = Math.max(0, Math.floor((mountedAt - created.getTime()) / 86_400_000));
				return { stat: monthYear, delta: `● ${days} day${days === 1 ? "" : "s"}` };
			})()
		: null;

	const stats: StatEntry[] = MOCK_STATS.map((entry, index) => {
		if (index === 2 && memberSince) return { ...entry, stat: memberSince.stat, delta: memberSince.delta };
		return liveStats?.[index] ? { ...entry, stat: liveStats[index].stat, delta: liveStats[index].delta } : entry;
	});

	if (error) {
		return (
			<AppShell active="overview" user={{ name, email }}>
				<Center py="20">
					<Text color="danger">{error.message}</Text>
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
					<Text color="fg.subtle" maxW="560px">
						Last sign-in just now from this device. Pick up where you left off, or spin up something new.
					</Text>
				</Box>
			</GlassCard>

			<Grid templateColumns={{ base: "1fr", lg: "repeat(3, 1fr)" }} gap="4">
				{stats.map((s, i) => (
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
					{activityLoading ? (
						<Center py="10">
							<Spinner color="accent" />
						</Center>
					) : activityError ? (
						<Center py="10">
							<Text fontSize="sm" color="danger">
								{activityError.message}
							</Text>
						</Center>
					) : activity.length === 0 ? (
						<Center py="10">
							<Text fontSize="sm" color="fg.muted">
								No recent activity
							</Text>
						</Center>
					) : (
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
					)}
				</GlassCard>
			</Box>
		</AppShell>
	);
};
