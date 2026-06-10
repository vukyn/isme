"use client";

import { Box, Center, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { LuArrowRight, LuLock, LuUser, LuShieldCheck, LuFileCheck, LuCheck } from "react-icons/lu";
import type { IconType } from "react-icons";
import { GlassCard } from "@/components/ui/glass-card";
import { AppTile } from "@/components/AppTile";
import { BrandMark } from "@/components/ui/brand-mark";
import { Button } from "@/components/ui/button";
import { Link } from "@/components/ui/link";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { SSOScope } from "@/types";

interface SSOConsentProps {
	appName: string;
	appCode: string;
	appIcon: string;
	appColor: string;
	userName: string;
	userEmail: string;
	scopes: SSOScope[];
	loading: boolean;
	onAllow: () => void;
	onSwitchAccount: () => void;
	onDeny: () => void;
}

/** Avatar initials derived from the user's name (no avatar field on the entity). */
const initials = (name: string) =>
	name
		.split(" ")
		.filter(Boolean)
		.map((s) => s[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

// Per-scope icon + accent, cycled by index so the list reads like the mock
// (cyan / violet / magenta). Scope copy itself comes from the backend.
const SCOPE_ACCENTS: { icon: IconType; color: string }[] = [
	{ icon: LuUser, color: "accentAlt" },
	{ icon: LuShieldCheck, color: "accent" },
	{ icon: LuFileCheck, color: "#EC4899" },
];

export const SSOConsent = ({
	appName,
	appCode,
	appIcon,
	appColor,
	userName,
	userEmail,
	scopes,
	loading,
	onAllow,
	onSwitchAccount,
	onDeny,
}: SSOConsentProps) => {
	const firstName = userName.split(" ")[0] || userName;

	return (
		<Box minH="100dvh" display="grid" placeItems="center" px="5" pt="10" pb="18">
			<Stack w="full" maxW="460px" gap="18px">
				{/* isme brand — the identity provider doing the auth */}
				<HStack justify="center" gap="2.5">
					<BrandMark />
					<Text fontWeight="bold" fontSize="lg" letterSpacing="-0.01em" color="fg">
						isme
					</Text>
				</HStack>

				<GlassCard tone="strong" p="7" pb="6" borderRadius="20px" position="relative" overflow="hidden">
					{/* ambient orbs */}
					<Box
						position="absolute"
						w="220px"
						h="220px"
						right="-70px"
						top="-90px"
						borderRadius="full"
						pointerEvents="none"
						css={{ filter: "blur(46px)", background: "radial-gradient(circle, rgba(99,102,241,0.50), transparent 70%)" }}
					/>
					<Box
						position="absolute"
						w="200px"
						h="200px"
						left="-70px"
						bottom="-90px"
						borderRadius="full"
						pointerEvents="none"
						css={{ filter: "blur(46px)", background: "radial-gradient(circle, rgba(236,72,153,0.42), transparent 70%)" }}
					/>

					<Box position="relative" zIndex={1}>
						{/* ===== consent header: which app wants access ===== */}
						<Stack align="center" textAlign="center" gap="3.5" mb="5">
							<HStack gap="3.5" aria-hidden="true">
								{/* requesting app — real stored icon/color tile */}
								<AppTile iconKey={appIcon} colorKey={appColor} size="lg" appCode={appCode} fallbackSeed={appCode} />
								<Center color="fg.muted">
									<svg width="26" height="16" viewBox="0 0 26 16" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
										<path d="M2 8h22" />
										<path d="m7 3-5 5 5 5" />
										<path d="m19 3 5 5-5 5" />
									</svg>
								</Center>
								{/* isme (IdP) — lg = 52px to match the medioa tile */}
								<BrandMark size="lg" />
							</HStack>
							<Heading as="h1" fontSize="21px" fontWeight="bold" letterSpacing="-0.02em" lineHeight="1.3" color="fg">
								<Text as="span" color="fg">
									{appName}
								</Text>{" "}
								wants to access
								<br />
								your isme account
							</Heading>
							<Text fontSize="sm" color="fg.muted" maxW="360px">
								You're already signed in to isme. Review what {appName} will be able to do, then continue.
							</Text>
						</Stack>

						{/* ===== logged-in identity chip ===== */}
						<HStack
							gap="3.5"
							px="3.5"
							py="3.5"
							mb="18px"
							borderRadius="glass"
							borderWidth="1px"
							borderColor="border.strong"
							bg="bg.glass"
							css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
						>
							<Center
								w="11"
								h="11"
								flex="0 0 auto"
								borderRadius="full"
								color="white"
								fontWeight="bold"
								fontSize="sm"
								boxShadow="0 0 16px rgba(139,92,246,0.45)"
								css={{ background: "conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)" }}
							>
								{initials(userName)}
							</Center>
							<Box flex="1" minW="0" lineHeight="1.3">
								<Text fontSize="sm" fontWeight="semibold" color="fg">
									{userName}
								</Text>
								<Text fontSize="xs" color="fg.muted" truncate>
									{userEmail}
								</Text>
							</Box>
							{/* "Not you? Switch account" → drops to the password login form */}
							<Box
								asChild
								flex="0 0 auto"
								fontSize="12.5px"
								fontWeight="semibold"
								color="fg.subtle"
								textAlign="right"
								lineHeight="1.25"
								whiteSpace="nowrap"
								cursor="pointer"
								_hover={{ color: "accentAlt" }}
								_disabled={{ opacity: 0.55, cursor: "not-allowed" }}
							>
								<button type="button" onClick={onSwitchAccount} disabled={loading}>
									Not you?
									<br />
									Switch account
								</button>
							</Box>
						</HStack>

						{/* ===== scopes / permissions list ===== */}
						<Text
							fontSize="11px"
							fontWeight="semibold"
							letterSpacing="0.14em"
							textTransform="uppercase"
							color="fg.muted"
							mb="3"
						>
							{appName} will be able to
						</Text>
						<Stack gap="2.5" mb="22px">
							{scopes.map((scope, i) => {
								const accent = SCOPE_ACCENTS[i % SCOPE_ACCENTS.length];
								const ScopeIcon = accent.icon;
								return (
									<HStack
										key={scope.title}
										align="flex-start"
										gap="3"
										px="3.5"
										py="3"
										borderRadius="glassSm"
										bg="bg.glass"
										borderWidth="1px"
										borderColor="border"
									>
										<Center
											w="34px"
											h="34px"
											flex="0 0 auto"
											borderRadius="10px"
											color={accent.color}
											borderWidth="1px"
											borderColor="border.strong"
											css={{ background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.18))" }}
										>
											<ScopeIcon size={17} />
										</Center>
										<Box flex="1">
											<Text fontSize="sm" fontWeight="semibold" color="fg" mb="0.5">
												{scope.title}
											</Text>
											<Text fontSize="12.5px" color="fg.muted">
												{scope.description}
											</Text>
										</Box>
										<Box color="success" flex="0 0 auto" mt="1">
											<LuCheck size={16} />
										</Box>
									</HStack>
								);
							})}
						</Stack>

						{/* ===== actions ===== */}
						<Stack gap="2.5">
							<Button
								type="button"
								onClick={onAllow}
								w="full"
								h="50px"
								borderWidth="1px"
								borderColor="transparent"
								loading={loading}
								loadingText="Authorizing…"
								color="white"
								borderRadius="glassSm"
								boxShadow="ctaGlow"
								_hover={{ boxShadow: "ctaGlowHi" }}
								_focusVisible={{ boxShadow: "focusRing" }}
								css={AURORA_CTA_STYLE}
							>
								<HStack gap="2.5">
									<Text>Allow &amp; continue as {firstName}</Text>
									<LuArrowRight />
								</HStack>
							</Button>
							<Button
								type="button"
								onClick={onDeny}
								disabled={loading}
								variant="outline"
								w="full"
								h="50px"
								borderRadius="glassSm"
								bg="bg.glass"
								borderColor="border.strong"
								color="fg"
								_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
								css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
							>
								Cancel
							</Button>
						</Stack>
					</Box>
				</GlassCard>

				{/* trust footer: secured-by pill + revoke line */}
				<HStack
					alignSelf="center"
					gap="2"
					px="3"
					py="1.5"
					borderRadius="full"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border.strong"
					color="fg.subtle"
					fontSize="xs"
					fontWeight="medium"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
				>
					<Box color="success" display="grid" placeItems="center">
						<LuLock size={13} />
					</Box>
					Protected by isme SSO
				</HStack>
				<Text textAlign="center" fontSize="xs" color="fg.muted">
					You can revoke {appName}'s access anytime in{" "}
					<Link to="/settings" fontSize="xs" color="fg.subtle" _hover={{ color: "accentAlt" }}>
						isme settings
					</Link>
					.
				</Text>
			</Stack>
		</Box>
	);
};
