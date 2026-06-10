"use client";

import { useMemo, useState } from "react";
import { Box, Center, Dialog, Flex, HStack, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import {
	LuClock,
	LuInfo,
	LuLaptop,
	LuLogOut,
	LuMapPin,
	LuMonitor,
	LuMonitorOff,
	LuRefreshCw,
	LuSmartphone,
	LuTrash2,
	LuTriangleAlert,
} from "react-icons/lu";
import { GlassCard } from "@/components/ui/glass-card";
import { Button } from "@/components/ui/button";
import { AppTile } from "@/components/AppTile";
import { AppShell } from "@/layouts/AppShell";
import { toaster } from "@/components/ui/toaster";
import { useUser } from "@/hooks/useUser";
import { useSessions } from "@/hooks/useSessions";
import { revokeMySession, revokeMyOtherSessions } from "@/apis";
import type { MySessionItem } from "@/types";
import { parseUserAgent } from "@/utils/agent";
import { formatRelative, formatRelativeTime, isZeroTime } from "@/utils/time";

type ExpiryTone = "ok" | "soon" | "neutral";

/** Picks a device glyph from the parsed user-agent label. */
const deviceIcon = (userAgent: string) => {
	const label = parseUserAgent(userAgent).toLowerCase();
	if (/iphone|android|mobile|ios/.test(label)) return <LuSmartphone size={22} />;
	if (/mac|laptop|macbook/.test(label)) return <LuLaptop size={22} />;
	return <LuMonitor size={20} />;
};

/** "expires in 6 days" + tone (mint plenty / amber soon). Empty → neutral "no expiry". */
const expiry = (expiresAt: string): { label: string; tone: ExpiryTone } => {
	if (isZeroTime(expiresAt)) return { label: "no expiry", tone: "neutral" };
	const ms = new Date(expiresAt).getTime() - Date.now();
	if (Number.isNaN(ms)) return { label: "no expiry", tone: "neutral" };
	if (ms <= 0) return { label: "expired", tone: "soon" };

	const hours = Math.floor(ms / 3_600_000);
	const days = Math.floor(hours / 24);
	const label =
		days >= 1
			? `expires in ${days} day${days > 1 ? "s" : ""}`
			: `expires in ${Math.max(hours, 1)} hour${hours > 1 ? "s" : ""}`;
	// under a day left = "soon" (amber); otherwise plenty (mint).
	return { label, tone: days < 1 ? "soon" : "ok" };
};

const EXPIRY_STYLE: Record<ExpiryTone, { color: string; borderColor: string; bg: string }> = {
	ok: { color: "aurora.mint", borderColor: "rgba(52,211,153,0.30)", bg: "rgba(52,211,153,0.08)" },
	soon: { color: "aurora.amber", borderColor: "rgba(245,158,11,0.35)", bg: "rgba(245,158,11,0.10)" },
	neutral: { color: "fg.subtle", borderColor: "border.strong", bg: "bg.glass" },
};

interface AppChipModel {
	label: string;
	/** Drives the isme identity-icon fallback; only set for first-party sessions. */
	appCode?: string;
	color: string;
	/** Stored app icon key for the tinted-glass tile; empty = fallback. */
	iconKey?: string;
	/** Stored app color key for the tinted-glass tile; empty = fallback. */
	colorKey?: string;
	/** Hashed-gradient fallback seed (the app id) when no stored appearance. */
	seed?: string;
}

/**
 * Resolves the owning-app chip from the enriched session:
 *   - empty app_service_id → first-party isme session (conic isme identity tile).
 *   - non-empty + enriched  → the real app name + its stored icon/color tile.
 *   - non-empty but unenriched (deleted app_service) → hashed-gradient fallback.
 */
const appChip = (session: MySessionItem): AppChipModel => {
	if (!session.app_service_id) {
		return { label: "isme", appCode: "isme", color: "fg" };
	}
	return {
		label: session.app_name || "App",
		color: session.app_name ? "fg" : "fg.subtle",
		iconKey: session.app_icon,
		colorKey: session.app_color,
		seed: session.app_service_id,
	};
};

const ExpiryPill = ({ expiresAt }: { expiresAt: string }) => {
	const { label, tone } = expiry(expiresAt);
	const style = EXPIRY_STYLE[tone];
	return (
		<HStack
			as="span"
			gap="1.5"
			px="2.5"
			py="1"
			borderRadius="full"
			borderWidth="1px"
			fontSize="11px"
			fontWeight="medium"
			whiteSpace="nowrap"
			color={style.color}
			borderColor={style.borderColor}
			bg={style.bg}
			css={{ fontVariantNumeric: "tabular-nums" }}
		>
			<LuClock size={12} />
			<Text as="span">{label}</Text>
		</HStack>
	);
};

const AppChip = ({ session }: { session: MySessionItem }) => {
	const chip = appChip(session);
	return (
		<HStack
			as="span"
			gap="1.5"
			pl="1"
			pr="2.5"
			py="1"
			borderRadius="full"
			borderWidth="1px"
			borderColor="border"
			bg="rgba(7,7,26,0.45)"
			fontSize="12px"
			fontWeight="semibold"
			whiteSpace="nowrap"
		>
			<AppTile
				size="chip"
				appCode={chip.appCode}
				iconKey={chip.iconKey}
				colorKey={chip.colorKey}
				fallbackSeed={chip.seed}
			/>
			<Text as="span" color={chip.color}>
				{chip.label}
			</Text>
		</HStack>
	);
};

/**
 * Per-session token-rotation metadata — a faint inline line under the IP/active
 * meta (no boxed pill), matching the aurora design. A never-refreshed session
 * (refresh_count 0 / empty last_refreshed_at) collapses to a single dim
 * "not yet refreshed" phrase instead of the noisier "0× · —".
 */
const RotationPill = ({ session }: { session: MySessionItem }) => {
	const never = session.refresh_count === 0 || isZeroTime(session.last_refreshed_at);
	const lastRefreshed = formatRelative(session.last_refreshed_at);
	return (
		<HStack
			as="span"
			gap="1.5"
			mt="1.5"
			fontSize="12px"
			whiteSpace="nowrap"
			maxW="100%"
			color="fg.muted"
			opacity={never ? 0.7 : 1}
			title={
				never
					? "No token refreshes yet on this session"
					: `This session's access token has been refreshed ${session.refresh_count} time${session.refresh_count > 1 ? "s" : ""}; most recent refresh ${lastRefreshed}`
			}
		>
			<Box as="span" color={never ? "fg.muted" : "aurora.cyan"} opacity={never ? 0.5 : 0.75} display="inline-flex" flex="none">
				<LuRefreshCw size={12} />
			</Box>
			{never ? (
				<Text as="span">not yet refreshed</Text>
			) : (
				<Text as="span">
					refreshed{" "}
					<Text as="span" color="fg.subtle" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
						{session.refresh_count}×
					</Text>
					<Text as="span" color="fg.muted" opacity={0.45} px="1">
						·
					</Text>
					<Text as="span" css={{ fontVariantNumeric: "tabular-nums" }}>
						{lastRefreshed}
					</Text>
				</Text>
			)}
		</HStack>
	);
};

interface SessionRowProps {
	session: MySessionItem;
	onRevoke: (session: MySessionItem) => void;
}

const SessionRow = ({ session, onRevoke }: SessionRowProps) => {
	const isCurrent = session.current;
	return (
		<Flex
			align="center"
			gap="4"
			px="4.5"
			py="4"
			borderRadius="2xl"
			borderWidth="1px"
			borderColor={isCurrent ? "rgba(139,92,246,0.45)" : "border"}
			bg={isCurrent ? "rgba(139,92,246,0.07)" : "rgba(7,7,26,0.35)"}
			boxShadow={isCurrent ? "0 0 0 1px rgba(139,92,246,0.20)" : "none"}
			wrap={{ base: "wrap", md: "nowrap" }}
			_hover={{ borderColor: "border.strong", bg: isCurrent ? "rgba(139,92,246,0.07)" : "rgba(255,255,255,0.03)" }}
		>
			<Center
				w="46px"
				h="46px"
				flex="none"
				borderRadius="13px"
				borderWidth="1px"
				borderColor={isCurrent ? "rgba(139,92,246,0.45)" : "border.strong"}
				bg={isCurrent ? "rgba(139,92,246,0.12)" : "bg.glass"}
				color={isCurrent ? "aurora.violet" : "fg.subtle"}
				css={isCurrent ? { boxShadow: "0 0 16px rgba(139,92,246,0.25)" } : undefined}
			>
				{deviceIcon(session.user_agent)}
			</Center>

			<Box flex="1" minW="0" lineHeight="1.35">
				<HStack gap="2.5" wrap="wrap">
					<Text fontSize="15px" fontWeight="semibold" color="fg" truncate>
						{parseUserAgent(session.user_agent)}
					</Text>
					{isCurrent && (
						<HStack
							as="span"
							gap="1.5"
							px="2.5"
							py="0.5"
							borderRadius="full"
							fontSize="11px"
							fontWeight="semibold"
							color="aurora.violet"
							bg="rgba(139,92,246,0.14)"
							borderWidth="1px"
							borderColor="rgba(139,92,246,0.40)"
							whiteSpace="nowrap"
						>
							<Box as="span" w="6px" h="6px" borderRadius="full" bg="aurora.violet" css={{ boxShadow: "0 0 10px #8B5CF6" }} />
							This device
						</HStack>
					)}
				</HStack>
				<HStack gap="6px 14px" wrap="wrap" mt="1.5" fontSize="12.5px" color="fg.muted">
					<HStack as="span" gap="1.5" whiteSpace="nowrap">
						<LuMapPin size={13} />
						<Text as="span" css={{ fontVariantNumeric: "tabular-nums" }}>
							{session.client_ip || "Unknown IP"}
						</Text>
					</HStack>
					<HStack as="span" gap="1.5" whiteSpace="nowrap">
						<LuClock size={13} />
						<Text as="span">
							Active{" "}
							<Text as="span" color="fg.subtle" fontWeight="semibold">
								{isZeroTime(session.last_login_at) ? "unknown" : formatRelativeTime(session.last_login_at)}
							</Text>
						</Text>
					</HStack>
				</HStack>
				<RotationPill session={session} />
			</Box>

			<Stack align={{ base: "flex-start", md: "flex-end" }} gap="2.5" flex="none">
				<AppChip session={session} />
				<HStack gap="2">
					<ExpiryPill expiresAt={session.expires_at} />
					{isCurrent ? (
						<HStack
							as="span"
							gap="1.5"
							h="34px"
							px="3"
							borderRadius="10px"
							borderWidth="1px"
							borderColor="border.strong"
							bg="rgba(7,7,26,0.55)"
							fontSize="13px"
							fontWeight="semibold"
							color="fg.muted"
							cursor="default"
							title="Sign out of this device from the account menu"
						>
							<LuMonitorOff size={14} />
							This device
						</HStack>
					) : (
						<Button
							variant="outline"
							h="34px"
							px="3"
							borderRadius="10px"
							borderColor="border.strong"
							bg="rgba(7,7,26,0.55)"
							fontSize="13px"
							fontWeight="semibold"
							color="fg.subtle"
							_hover={{ color: "aurora.magenta", borderColor: "rgba(236,72,153,0.45)", bg: "rgba(236,72,153,0.12)" }}
							onClick={() => onRevoke(session)}
							aria-label={`Revoke ${parseUserAgent(session.user_agent)} session`}
						>
							<LuLogOut size={14} />
							Revoke
						</Button>
					)}
				</HStack>
			</Stack>
		</Flex>
	);
};

export const Sessions = () => {
	const { user } = useUser();
	const name = user?.name || "User";
	const email = user?.email || "";

	const { sessions, loading, error, refetch } = useSessions();

	// per-session revoke confirm target (null = closed)
	const [revokeTarget, setRevokeTarget] = useState<MySessionItem | null>(null);
	const [revoking, setRevoking] = useState(false);
	// revoke-all-others confirm
	const [revokeAllOpen, setRevokeAllOpen] = useState(false);
	const [revokingAll, setRevokingAll] = useState(false);

	const otherCount = useMemo(() => sessions.filter((s) => !s.current).length, [sessions]);
	const appCount = useMemo(() => new Set(sessions.map((s) => s.app_service_id || "isme")).size, [sessions]);

	const handleRevoke = async () => {
		if (!revokeTarget) return;
		setRevoking(true);
		try {
			await revokeMySession(revokeTarget.id);
		} catch (err: unknown) {
			const message = (err as { response?: { data?: { message?: string } } })?.response?.data?.message;
			toaster.create({ title: message || "Failed to revoke session", type: "error", meta: { closable: true } });
			setRevoking(false);
			return;
		}
		setRevoking(false);
		setRevokeTarget(null);
		toaster.create({ title: "Session revoked", type: "success", meta: { closable: true } });
		await refetch();
	};

	const handleRevokeAll = async () => {
		setRevokingAll(true);
		try {
			await revokeMyOtherSessions();
		} catch (err: unknown) {
			const message = (err as { response?: { data?: { message?: string } } })?.response?.data?.message;
			toaster.create({ title: message || "Failed to revoke sessions", type: "error", meta: { closable: true } });
			setRevokingAll(false);
			return;
		}
		setRevokingAll(false);
		setRevokeAllOpen(false);
		toaster.create({ title: "Other sessions revoked", type: "success", meta: { closable: true } });
		await refetch();
	};

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

		if (sessions.length === 0) {
			return (
				<GlassCard>
					<Stack align="center" gap="3.5" px="6" py="14" textAlign="center">
						<Center w="64px" h="64px" borderRadius="20px" borderWidth="1px" borderColor="border.strong" bg="bg.glass" color="fg.muted">
							<LuMonitorOff size={28} />
						</Center>
						<Heading as="h3" fontSize="17px" fontWeight="semibold" color="fg">
							No active sessions
						</Heading>
						<Text fontSize="13px" color="fg.muted" maxW="360px">
							New sign-ins from other browsers or devices will appear here.
						</Text>
					</Stack>
				</GlassCard>
			);
		}

		return (
			<Stack gap="5">
				<GlassCard>
					<HStack gap="4" px="5" py="4">
						<Center
							w="42px"
							h="42px"
							flex="none"
							borderRadius="13px"
							borderWidth="1px"
							borderColor="border.strong"
							color="aurora.cyan"
							css={{
								background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(34,211,238,0.18))",
								boxShadow: "0 0 18px rgba(99,102,241,0.18)",
							}}
						>
							<LuMonitor size={20} />
						</Center>
						<Box minW="0" lineHeight="1.3">
							<Text fontSize="17px" fontWeight="bold" letterSpacing="-0.01em" css={{ fontVariantNumeric: "tabular-nums" }}>
								<Text as="span" color="aurora.cyan">
									{sessions.length}
								</Text>{" "}
								active session{sessions.length > 1 ? "s" : ""}
							</Text>
							<Text fontSize="12px" color="fg.muted" mt="0.5">
								Across {appCount} app{appCount > 1 ? "s" : ""} · {sessions.length - otherCount} on this device. Revoke any session you don't recognise.
							</Text>
						</Box>
					</HStack>
				</GlassCard>

				<Stack gap="3.5">
					{sessions.map((session) => (
						<SessionRow key={session.id} session={session} onRevoke={setRevokeTarget} />
					))}
				</Stack>

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
						<Text as="span" color="fg.subtle" fontWeight="semibold">
							Revoke
						</Text>{" "}
						ends a session immediately on every app — the next request from that device is signed out. You can't revoke the
						session you're using now; sign out from the account menu instead.
					</Text>
				</HStack>
			</Stack>
		);
	};

	return (
		<AppShell active="sessions" user={{ name, email }}>
			<Flex justify="space-between" align="flex-end" gap="4" wrap="wrap">
				<Box>
					<Heading as="h1" fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1">
						Active sessions
					</Heading>
					<Text mt="1.5" color="fg.muted" fontSize="sm">
						Devices currently signed in to your account
					</Text>
				</Box>
				{otherCount > 0 && (
					<Button
						variant="outline"
						h="9"
						px="3.5"
						borderRadius="10px"
						fontSize="13px"
						fontWeight="semibold"
						color="aurora.magenta"
						borderColor="rgba(236,72,153,0.35)"
						bg="rgba(236,72,153,0.08)"
						_hover={{ bg: "rgba(236,72,153,0.16)", borderColor: "rgba(236,72,153,0.55)" }}
						onClick={() => setRevokeAllOpen(true)}
						aria-label="Revoke all other sessions"
					>
						<LuTrash2 size={15} />
						Revoke all other sessions
					</Button>
				)}
			</Flex>

			{renderBody()}

			{/* per-session revoke confirm */}
			<Dialog.Root open={revokeTarget !== null} onOpenChange={(details) => !details.open && setRevokeTarget(null)} placement="center">
				<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
				<Dialog.Positioner>
					<Dialog.Content
						w="440px"
						maxW="92vw"
						borderRadius="20px"
						borderWidth="1px"
						borderColor="rgba(236,72,153,0.40)"
						bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
						color="fg"
						boxShadow="0 20px 60px rgba(236,72,153,0.18), 0 4px 12px rgba(0,0,0,0.35)"
					>
						<Dialog.Header px="5" py="4" borderBottomWidth="1px" borderColor="border">
							<HStack gap="3">
								<Center w="9" h="9" borderRadius="11px" bg="rgba(236,72,153,0.12)" borderWidth="1px" borderColor="rgba(236,72,153,0.35)" color="aurora.magenta">
									<LuLogOut size={16} />
								</Center>
								<Dialog.Title fontSize="15px" fontWeight="semibold">
									Revoke session
								</Dialog.Title>
							</HStack>
						</Dialog.Header>
						<Dialog.Body p="5">
							<Text fontSize="sm" color="fg.subtle">
								Sign out <Text as="span" color="fg" fontWeight="semibold">{revokeTarget ? parseUserAgent(revokeTarget.user_agent) : ""}</Text>? The next
								request from that device will be signed out.
							</Text>
						</Dialog.Body>
						<HStack justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
							<Button variant="outline" h="11" px="4.5" borderRadius="glassSm" borderColor="border.strong" bg="bg.glass" fontSize="sm" fontWeight="semibold" color="fg" _hover={{ bg: "bg.glassHi" }} onClick={() => setRevokeTarget(null)}>
								Cancel
							</Button>
							<Button
								h="11"
								px="4.5"
								borderRadius="glassSm"
								fontSize="sm"
								fontWeight="bold"
								color="white"
								bg="linear-gradient(135deg, #EC4899, #8B5CF6)"
								boxShadow="0 10px 30px rgba(236,72,153,0.35)"
								_hover={{ boxShadow: "0 14px 40px rgba(236,72,153,0.50)" }}
								loading={revoking}
								onClick={handleRevoke}
							>
								<LuLogOut size={15} /> Revoke session
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>

			{/* revoke all other sessions confirm */}
			<Dialog.Root open={revokeAllOpen} onOpenChange={(details) => !details.open && setRevokeAllOpen(false)} placement="center">
				<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
				<Dialog.Positioner>
					<Dialog.Content
						w="440px"
						maxW="92vw"
						borderRadius="20px"
						borderWidth="1px"
						borderColor="rgba(236,72,153,0.40)"
						bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
						color="fg"
						boxShadow="0 20px 60px rgba(236,72,153,0.18), 0 4px 12px rgba(0,0,0,0.35)"
					>
						<Dialog.Header px="5" py="4" borderBottomWidth="1px" borderColor="border">
							<HStack gap="3">
								<Center w="9" h="9" borderRadius="11px" bg="rgba(236,72,153,0.12)" borderWidth="1px" borderColor="rgba(236,72,153,0.35)" color="aurora.magenta">
									<LuTriangleAlert size={16} />
								</Center>
								<Dialog.Title fontSize="15px" fontWeight="semibold">
									Revoke all other sessions
								</Dialog.Title>
							</HStack>
						</Dialog.Header>
						<Dialog.Body p="5">
							<Text fontSize="sm" color="fg.subtle">
								Sign out of{" "}
								<Text as="span" color="fg" fontWeight="semibold">
									{otherCount} other session{otherCount > 1 ? "s" : ""}
								</Text>
								? Your current device stays signed in.
							</Text>
						</Dialog.Body>
						<HStack justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
							<Button variant="outline" h="11" px="4.5" borderRadius="glassSm" borderColor="border.strong" bg="bg.glass" fontSize="sm" fontWeight="semibold" color="fg" _hover={{ bg: "bg.glassHi" }} onClick={() => setRevokeAllOpen(false)}>
								Cancel
							</Button>
							<Button
								h="11"
								px="4.5"
								borderRadius="glassSm"
								fontSize="sm"
								fontWeight="bold"
								color="white"
								bg="linear-gradient(135deg, #EC4899, #8B5CF6)"
								boxShadow="0 10px 30px rgba(236,72,153,0.35)"
								_hover={{ boxShadow: "0 14px 40px rgba(236,72,153,0.50)" }}
								loading={revokingAll}
								onClick={handleRevokeAll}
							>
								<LuTrash2 size={15} /> Revoke all others
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>
		</AppShell>
	);
};
