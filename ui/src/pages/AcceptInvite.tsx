"use client";

import { useEffect, useMemo, useState } from "react";
import { Box, Button, Center, Flex, Heading, HStack, Spinner, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink, useNavigate, useSearchParams } from "react-router-dom";
import {
	LuArrowRight,
	LuChevronDown,
	LuClock,
	LuKeyRound,
	LuLink2Off,
	LuLock,
	LuMail,
	LuShieldCheck,
	LuUser,
	LuUserCheck,
} from "react-icons/lu";
import { acceptInvite, getInvitationByToken } from "@/apis";
import { AppTile } from "@/components/AppRoleChip";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { PasswordField } from "@/components/ui/password-field";
import { PasswordStrength } from "@/components/ui/password-strength";
import { BrandPanel } from "@/components/ui/brand-panel";
import { toaster } from "@/components/ui/toaster";
import { AuthLayout } from "@/layouts/AuthLayout";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { InviteAssignmentDetail, InviteDetailResponse } from "@/types";
import { acceptInviteSchema, type AcceptInviteFormData } from "@/validators";

const INVITE_FEATURES = [
	{ icon: <LuUserCheck />, title: "Pre-verified account", desc: "Your inviter vouched — sign in right away." },
	{ icon: <LuKeyRound />, title: "You pick the password", desc: "Set it once, only you know it." },
	{ icon: <LuShieldCheck />, title: "Single-use link", desc: "This invitation is consumed on account creation." },
];

const PERMS_PREVIEW_COUNT = 4;

/**
 * Per-assignment access card (mock .acard). Renders one app's identity, the
 * granted role, and the resource:action permission list — expandable, with the
 * list collapsing to "+N more" past PERMS_PREVIEW_COUNT.
 */
const AccessCard = ({ assignment }: { assignment: InviteAssignmentDetail }) => {
	const [open, setOpen] = useState(true);
	const [permsOpen, setPermsOpen] = useState(false);
	const perms = assignment.permissions ?? [];
	const visible = permsOpen ? perms : perms.slice(0, PERMS_PREVIEW_COUNT);
	const hidden = perms.length - PERMS_PREVIEW_COUNT;

	return (
		<Box
			borderWidth="1px"
			borderColor={open ? "rgba(139,92,246,0.45)" : "border.strong"}
			borderRadius="16px"
			bg="rgba(7,7,26,0.35)"
			overflow="hidden"
			transition="border-color .2s ease-out"
		>
			<HStack
				as="button"
				gap="3"
				px="4"
				py="3.5"
				w="full"
				textAlign="left"
				cursor="pointer"
				_hover={{ bg: "rgba(255,255,255,0.03)" }}
				onClick={(event: React.MouseEvent) => {
					event.preventDefault();
					setOpen((value) => !value);
				}}
			>
				<AppTile appCode={assignment.app_code} size="38px" radius="11px" />
				<Box flex="1" minW="0" lineHeight="1.3">
					<Text fontSize="15px" fontWeight="semibold" color="fg" truncate>
						{assignment.app_name || assignment.app_code}
					</Text>
					<Text fontSize="12px" color="fg.muted" truncate>
						{assignment.app_code}
					</Text>
				</Box>
				<Box
					display="inline-flex"
					alignItems="center"
					gap="6px"
					px="11px"
					py="5px"
					borderRadius="full"
					fontSize="12px"
					fontWeight="semibold"
					color="aurora.cyan"
					borderWidth="1px"
					borderColor="rgba(34,211,238,0.40)"
					bg="rgba(34,211,238,0.10)"
					whiteSpace="nowrap"
				>
					{assignment.role_name || assignment.role_code || "member"}
				</Box>
				<Box color="fg.muted" transition="transform .2s ease-out" transform={open ? "rotate(180deg)" : undefined}>
					<LuChevronDown size={18} />
				</Box>
			</HStack>
			{open && (
				<Box px="4" pb="4" borderTopWidth="1px" borderColor="border">
					<Text fontSize="11px" fontWeight="semibold" letterSpacing="0.12em" textTransform="uppercase" color="fg.muted" mt="3.5" mb="2.5">
						This role can
					</Text>
					{perms.length === 0 ? (
						<Text fontSize="13px" color="fg.muted">
							This role grants no explicit permissions.
						</Text>
					) : (
						<Stack gap="2">
							{visible.map((perm) => {
								const [resource, action] = perm.split(":");
								return (
									<HStack
										key={perm}
										gap="2"
										px="2.5"
										py="1.5"
										borderRadius="9px"
										bg="bg.glass"
										borderWidth="1px"
										borderColor="border"
										fontSize="13px"
									>
										<Box color="aurora.mint" flex="none">
											<LuShieldCheck size={14} />
										</Box>
										<Text>
											<Text as="span" color="fg" fontWeight="medium">
												{resource}
											</Text>
											<Text as="span" color="fg.muted">
												:{action}
											</Text>
										</Text>
									</HStack>
								);
							})}
							{hidden > 0 && (
								<HStack
									as="button"
									justify="center"
									gap="2"
									px="2.5"
									py="2"
									borderRadius="9px"
									cursor="pointer"
									fontSize="12px"
									fontWeight="semibold"
									color="aurora.cyan"
									bg="rgba(34,211,238,0.07)"
									borderWidth="1px"
									borderStyle="dashed"
									borderColor="rgba(34,211,238,0.35)"
									_hover={{ bg: "rgba(34,211,238,0.12)" }}
									onClick={(event: React.MouseEvent) => {
										event.preventDefault();
										setPermsOpen((value) => !value);
									}}
								>
									{permsOpen ? "Show fewer" : `+${hidden} more permission${hidden === 1 ? "" : "s"}`}
								</HStack>
							)}
						</Stack>
					)}
				</Box>
			)}
		</Box>
	);
};

/** Centered status card (mock terminal states) — used for expired / revoked. */
const StatusCard = ({
	icon,
	tone,
	title,
	body,
}: {
	icon: React.ReactNode;
	tone: string;
	title: string;
	body: React.ReactNode;
}) => (
	<Flex minH="100vh" align="center" justify="center" px="6" py="12">
		<Box
			w="full"
			maxW="460px"
			borderRadius="20px"
			borderWidth="1px"
			borderColor={`${tone}30`}
			bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
			boxShadow="glassPop"
			px="8"
			py="9"
			textAlign="center"
		>
			<Center
				w="56px"
				h="56px"
				mx="auto"
				mb="4.5"
				borderRadius="18px"
				bg={`${tone}1F`}
				borderWidth="1px"
				borderColor={`${tone}66`}
				color={tone}
				boxShadow={`0 0 28px ${tone}38`}
			>
				{icon}
			</Center>
			<Heading as="h2" fontSize="22px" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				{title}
			</Heading>
			<Box fontSize="sm" color="fg.muted" mb="6">
				{body}
			</Box>
			<Text fontSize="12px" color="fg.muted">
				<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg-subtle)", fontWeight: 500 }}>
					Go to sign in
				</RouterLink>{" "}
				or contact your administrator.
			</Text>
		</Box>
	</Flex>
);

/**
 * Public /accept-invite?token=… page (wide aurora split). Resolves the token on
 * load, then drives the screen off display_status:
 *  - "valid"    → access summary (one card per assignment) + account setup form
 *  - "accepted" → already-accepted card pointing at sign in
 *  - "expired"  → expired card (no request-new action)
 *  - "revoked"  → revoked card (contact admin)
 * A missing token / network error falls back to the same expired-style card.
 */
export const AcceptInvite = () => {
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const token = searchParams.get("token") ?? "";

	const [loading, setLoading] = useState(true);
	const [loadFailed, setLoadFailed] = useState(false);
	const [detail, setDetail] = useState<InviteDetailResponse | null>(null);

	const [name, setName] = useState("");
	const [password, setPassword] = useState("");
	const [confirmPassword, setConfirmPassword] = useState("");
	const [acceptTerms, setAcceptTerms] = useState(false);
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof AcceptInviteFormData, string>>>({});
	const [submitting, setSubmitting] = useState(false);

	useEffect(() => {
		if (!token) {
			setLoadFailed(true);
			setLoading(false);
			return;
		}
		let active = true;
		getInvitationByToken(token)
			.then((response) => {
				if (active) setDetail(response);
			})
			.catch(() => {
				if (active) setLoadFailed(true);
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, [token]);

	const passwordsMatch = confirmPassword.length > 0 && password === confirmPassword;
	// CTA gate (decision): terms checked + password ≥ 8 + passwords match.
	const canSubmit = acceptTerms && password.length >= 8 && passwordsMatch;

	const accessFeatures = useMemo(() => INVITE_FEATURES, []);

	const handleSubmit = async (event: React.FormEvent) => {
		event.preventDefault();
		setFormErrors({});
		const result = acceptInviteSchema.safeParse({ name, password, confirmPassword, acceptTerms });
		if (!result.success) {
			const errors: Partial<Record<keyof AcceptInviteFormData, string>> = {};
			result.error.issues.forEach((issue) => {
				if (issue.path[0]) errors[issue.path[0] as keyof AcceptInviteFormData] = issue.message;
			});
			setFormErrors(errors);
			return;
		}
		setSubmitting(true);
		try {
			await acceptInvite({ token, name: result.data.name, password: result.data.password });
			toaster.create({ title: "Account created", description: "Sign in with your new password.", type: "success" });
			navigate("/login");
		} catch {
			// consumed / expired between load and submit — refetch to resolve the
			// real terminal state instead of guessing.
			setLoadFailed(true);
		} finally {
			setSubmitting(false);
		}
	};

	if (loading) {
		return (
			<Center minH="100vh">
				<Spinner size="lg" color="accent" />
			</Center>
		);
	}

	// Network / missing-token fallback (no display_status to switch on).
	if (loadFailed || !detail) {
		return (
			<StatusCard
				icon={<LuLink2Off size={24} />}
				tone="#F87171"
				title="This invitation is no longer valid"
				body={
					<>
						It may have <b>expired</b>, been <b>revoked</b>, or <b>already been used</b>. Invite links work once and expire after
						7 days.
					</>
				}
			/>
		);
	}

	if (detail.display_status === "accepted") {
		return (
			<StatusCard
				icon={<LuUserCheck size={24} />}
				tone="#34D399"
				title="This invitation was already accepted"
				body={
					<>
						The account for <b>{detail.email}</b> already exists. Sign in with the password you set.
					</>
				}
			/>
		);
	}

	if (detail.display_status === "expired") {
		return (
			<StatusCard
				icon={<LuClock size={24} />}
				tone="#F59E0B"
				title="This invitation has expired"
				body={<>Invite links expire after 7 days. Ask your administrator to send a new one.</>}
			/>
		);
	}

	if (detail.display_status === "revoked") {
		return (
			<StatusCard
				icon={<LuLink2Off size={24} />}
				tone="#EC4899"
				title="This invitation was revoked"
				body={<>This link has been revoked and can no longer be used. Contact your administrator for a new invite.</>}
			/>
		);
	}

	// display_status === "valid"
	const assignments = detail.assignments ?? [];

	return (
		<AuthLayout
			brand={
				<BrandPanel
					pill="You've been invited"
					pillTone="violet"
					titleLead="Welcome to"
					titleGrad="isme"
					sub={`Review the access you'll receive, then set a password to finish creating your account. This link works once and can't be reused.`}
					features={accessFeatures}
				/>
			}
		>
			{/* ===== Access summary — one card per assignment ===== */}
			<Box mb="7">
				<Flex align="center" justify="space-between" mb="3">
					<Text fontSize="11px" fontWeight="semibold" letterSpacing="0.14em" textTransform="uppercase" color="fg.muted">
						Your access
					</Text>
					<Text fontSize="12px" color="fg.muted">
						Tap a card for permissions
					</Text>
				</Flex>
				{assignments.length === 0 ? (
					<Box
						borderWidth="1px"
						borderColor="border.strong"
						borderRadius="16px"
						bg="rgba(7,7,26,0.35)"
						px="4"
						py="3.5"
						fontSize="13px"
						color="fg.muted"
					>
						No roles are pre-assigned — your administrator will grant access after you sign in.
					</Box>
				) : (
					<Stack gap="2.5">
						{assignments.map((assignment) => (
							<AccessCard key={`${assignment.app_code}-${assignment.role_code}`} assignment={assignment} />
						))}
					</Stack>
				)}
			</Box>

			{/* ===== Account setup ===== */}
			<Heading as="h2" fontSize="2xl" fontWeight="bold" letterSpacing="-0.02em" mb="1.5" color="fg">
				Set up your account
			</Heading>
			<Text color="fg.muted" mb="6" fontSize="sm">
				Your email is locked to the invite — choose a password to finish.
			</Text>

			<Stack as="form" onSubmit={handleSubmit} gap="4">
				{/* Email is bound to the invitation — locked, not editable */}
				<Box>
					<Text fontSize="13px" fontWeight="medium" color="fg.subtle" mb="2">
						Email
					</Text>
					<Input
						type="email"
						value={detail.email}
						disabled
						aria-label="Email (locked by invitation)"
						startElement={<LuMail />}
						endElement={
							<Box color="aurora.amber">
								<LuLock size={15} />
							</Box>
						}
					/>
					<Text fontSize="12px" color="fg.muted" mt="2">
						Locked to this invitation.
					</Text>
				</Box>

				<Box>
					<Text fontSize="13px" fontWeight="medium" color="fg.subtle" mb="2">
						Your name
					</Text>
					<Input
						type="text"
						autoComplete="name"
						placeholder="Full name"
						value={name}
						onChange={(event) => setName(event.target.value)}
						startElement={<LuUser />}
					/>
					{formErrors.name && (
						<Text fontSize="12px" color="aurora.magenta" mt="1.5">
							{formErrors.name}
						</Text>
					)}
				</Box>

				<Box>
					<PasswordField
						label="Password"
						value={password}
						onChange={(event) => setPassword(event.target.value)}
						error={formErrors.password}
						autoComplete="new-password"
						placeholder="At least 8 characters"
					/>
					<PasswordStrength value={password} />
				</Box>

				<PasswordField
					label="Confirm password"
					value={confirmPassword}
					onChange={(event) => setConfirmPassword(event.target.value)}
					error={formErrors.confirmPassword || (confirmPassword.length > 0 && !passwordsMatch ? "Passwords don't match" : undefined)}
					autoComplete="new-password"
					placeholder="Re-enter password"
				/>

				<Box>
					<Checkbox
						size="sm"
						colorPalette="purple"
						checked={acceptTerms}
						onCheckedChange={(details) => setAcceptTerms(!!details.checked)}
					>
						<Text fontSize="13px" color="fg.subtle">
							I agree to the{" "}
							<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg-subtle)", textDecoration: "underline" }}>
								Terms
							</RouterLink>{" "}
							and{" "}
							<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg-subtle)", textDecoration: "underline" }}>
								Privacy Policy
							</RouterLink>
							, and to receiving the access listed above.
						</Text>
					</Checkbox>
					{formErrors.acceptTerms && (
						<Text fontSize="12px" color="aurora.magenta" mt="1.5">
							{formErrors.acceptTerms}
						</Text>
					)}
				</Box>

				<Button
					type="submit"
					h="12"
					mt="1.5"
					loading={submitting}
					disabled={!canSubmit}
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					_focusVisible={{ boxShadow: "focusRing" }}
					css={AURORA_CTA_STYLE}
				>
					<HStack gap="2.5">
						<Text>Accept invite &amp; create account</Text>
						<LuArrowRight />
					</HStack>
				</Button>

				<HStack gap="2" justify="center" fontSize="12px" color="fg.muted">
					<LuClock size={13} /> This link expires in 7 days and works once.
				</HStack>

				<Text fontSize="12px" color="fg.muted" textAlign="center">
					Already have an account?{" "}
					<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg-subtle)", fontWeight: 500 }}>
						Sign in
					</RouterLink>
				</Text>
			</Stack>
		</AuthLayout>
	);
};
