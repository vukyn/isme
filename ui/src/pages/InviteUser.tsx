"use client";

import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Box, Button, Center, Dialog, Flex, Heading, HStack, Input, Spinner, Text } from "@chakra-ui/react";
import { LuCheck, LuChevronRight, LuClock, LuCopy, LuInfo, LuLock, LuMail, LuSend, LuTriangleAlert, LuUserPlus } from "react-icons/lu";
import { createInvitation, listAppServices } from "@/apis";
import { AppScopedRoleAssignments } from "@/components/AppScopedRoleAssignments";
import { Checkbox } from "@/components/ui/checkbox";
import { toaster } from "@/components/ui/toaster";
import { APP_SERVICE_STATUS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { useUser } from "@/hooks/useUser";
import { AppShell } from "@/layouts/AppShell";
import type { AppService, InvitationRoleAssignment } from "@/types";
import { copyToClipboard } from "@/utils";
import { inviteUserSchema } from "@/validators";

/**
 * Full-page invite (mock aurora-user-invite.html). Invitee email + one or more
 * App → Role assignments scoped per app.
 *
 * Backend (user_invitation models.CreateRequest = { email, assignments[] }):
 * each assignment becomes one app-scoped user_role { role_id, app_service_id }
 * on acceptance. The success view surfaces the one-time invite link (no email is
 * sent — the link is shared manually). CreateRequest takes only email +
 * assignments, so no name / message fields are collected.
 */

const FIELD_LABEL_PROPS = {
	display: "block",
	fontSize: "13px",
	fontWeight: "medium",
	color: "fg.subtle",
	mb: "2",
} as const;

const INPUT_PROPS = {
	h: "11",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "sm",
	color: "fg",
	css: { backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" },
	_placeholder: { color: "fg.muted" },
	_hover: { borderColor: "rgba(255,255,255,0.28)" },
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" },
} as const;

const PANEL_PROPS = {
	bg: "bg.glass",
	borderWidth: "1px",
	borderColor: "border",
	borderRadius: "20px",
	boxShadow: "glassSoft",
	overflow: "hidden",
	css: { backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" },
} as const;

const errorMessage = (error: unknown, fallback: string): string => {
	const err = error as { response?: { data?: { message?: string } }; message?: string };
	return err?.response?.data?.message || err?.message || fallback;
};

const PanelHead = ({ icon, title, sub }: { icon: React.ReactNode; title: string; sub: string }) => (
	<HStack gap="3" px="5" py="4" borderBottomWidth="1px" borderColor="border">
		<Center
			w="9"
			h="9"
			flex="none"
			borderRadius="11px"
			bg="linear-gradient(135deg, rgba(99,102,241,0.22), rgba(236,72,153,0.16))"
			borderWidth="1px"
			borderColor="border.strong"
			color="aurora.cyan"
		>
			{icon}
		</Center>
		<Box lineHeight="1.3">
			<Text fontSize="15px" fontWeight="semibold" color="fg">
				{title}
			</Text>
			<Text fontSize="12px" color="fg.muted">
				{sub}
			</Text>
		</Box>
	</HStack>
);

export const InviteUser = () => {
	const { user: currentUser } = useUser();
	const navigate = useNavigate();

	const [email, setEmail] = useState("");
	const [emailError, setEmailError] = useState<string | null>(null);
	const [assignmentError, setAssignmentError] = useState<string | null>(null);

	const [apps, setApps] = useState<AppService[]>([]);
	const [assignments, setAssignments] = useState<InvitationRoleAssignment[]>([]);
	const [loading, setLoading] = useState(true);

	const [submitting, setSubmitting] = useState(false);
	// One-time invite link hand-off after a successful POST (the raw token is
	// only ever shown here — never re-displayable). Exit is Done-gated.
	const [inviteLink, setInviteLink] = useState<string | null>(null);
	const [acknowledged, setAcknowledged] = useState(false);
	const [copied, setCopied] = useState(false);

	// Load active apps for the assignment App selects; seed one empty row.
	useEffect(() => {
		let active = true;
		listAppServices({ page: 1, page_size: 100, status: APP_SERVICE_STATUS.ACTIVE })
			.then((response) => {
				if (!active) return;
				const items = response.items;
				setApps(items);
				const first = items[0];
				if (first) setAssignments([{ app_service_id: first.id, role_id: "" }]);
			})
			.catch(() => {
				if (active) toaster.create({ title: "Failed to load apps", type: "error", meta: { closable: true } });
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, []);

	// At least one fully-specified assignment is required.
	const completeAssignments = assignments.filter(
		(assignment) => assignment.app_service_id && assignment.role_id
	);
	const canSubmit = email.length > 0 && completeAssignments.length > 0;

	const handleSubmit = async () => {
		const parsed = inviteUserSchema.safeParse({ email, assignments: completeAssignments });
		if (!parsed.success) {
			const issue = parsed.error.issues[0];
			if (issue?.path[0] === "assignments") setAssignmentError(issue.message);
			else setEmailError(issue?.message ?? "Invalid input");
			return;
		}
		setEmailError(null);
		setAssignmentError(null);
		setSubmitting(true);
		try {
			const response = await createInvitation(parsed.data);
			setInviteLink(`${window.location.origin}${response.invite_link}`);
		} catch (error: unknown) {
			const message = errorMessage(error, "");
			if (message) setEmailError(message);
			else toaster.create({ title: "Failed to create invite", type: "error", meta: { closable: true } });
		} finally {
			setSubmitting(false);
		}
	};

	const handleInviteLinkCopy = async () => {
		if (!inviteLink) return;
		try {
			await copyToClipboard(inviteLink);
			setCopied(true);
			toaster.create({ title: "Invite link copied", type: "success", meta: { closable: true } });
			setTimeout(() => setCopied(false), 1500);
		} catch {
			toaster.create({ title: "Failed to copy — select the link manually", type: "error", meta: { closable: true } });
		}
	};

	return (
		<AppShell active="users" user={{ name: currentUser?.name || "User", email: currentUser?.email || "" }}>
			{/* Breadcrumb */}
			<HStack gap="2" fontSize="13px" color="fg.muted">
				<Box as="button" cursor="pointer" _hover={{ color: "fg" }} onClick={() => navigate("/users")}>
					Users
				</Box>
				<Text opacity={0.5}>/</Text>
				<Text color="fg.subtle">Invite user</Text>
			</HStack>

			<Box>
				<Heading fontSize="28px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1" color="fg">
					Invite a user
				</Heading>
				<Text fontSize="sm" color="fg.muted" mt="1.5" maxW="640px">
					Send an invitation with roles pre-assigned per app. The invitee accepts via a one-time link and sets their own
					password — each assignment becomes one{" "}
					<Text as="span" color="fg.subtle">
						user_role {"{ role_id, app_service_id }"}
					</Text>{" "}
					on acceptance.
				</Text>
			</Box>

			{loading ? (
				<Center py="16">
					<Spinner size="lg" color="accent" />
				</Center>
			) : (
				<>
					{/* Invitee details */}
					<Box {...PANEL_PROPS}>
						<PanelHead
							icon={<LuUserPlus size={17} />}
							title="Invitee"
							sub="where to send the invite — no account exists yet, they create one on acceptance"
						/>
						<Box p="5" display="grid" gap="4.5">
							<HStack
								gap="3"
								align="flex-start"
								px="4"
								py="3.5"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="rgba(34,211,238,0.30)"
								bg="linear-gradient(135deg, rgba(34,211,238,0.10), rgba(99,102,241,0.08))"
							>
								<Center
									w="30px"
									h="30px"
									flex="none"
									borderRadius="9px"
									bg="rgba(34,211,238,0.14)"
									borderWidth="1px"
									borderColor="rgba(34,211,238,0.40)"
									color="aurora.cyan"
								>
									<LuSend size={15} />
								</Center>
								<Box lineHeight="1.5">
									<Text fontSize="sm" fontWeight="semibold" color="fg">
										No email is sent — you get a one-time link to share.
									</Text>
									<Text fontSize="12.5px" color="fg.subtle" mt="0.5">
										The invitee lands on a public acceptance page showing exactly the access below, sets a password, and the
										account is created with these roles applied. The link expires in 7 days and can be revoked from the Users
										list.
									</Text>
								</Box>
							</HStack>

							<Box maxW="420px">
								<Text {...FIELD_LABEL_PROPS}>
									Invitee email{" "}
									<Text as="span" color="aurora.magenta">
										*
									</Text>
								</Text>
								<Box position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
									<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
										<LuMail size={16} />
									</Box>
									<Input
										{...INPUT_PROPS}
										pl="10"
										type="email"
										autoComplete="email"
										placeholder="name@example.com"
										borderColor={emailError ? "rgba(236,72,153,0.55)" : "border.strong"}
										value={email}
										onChange={(event) => {
											setEmail(event.target.value);
											if (emailError) setEmailError(null);
										}}
									/>
								</Box>
								{emailError ? (
									<HStack gap="1.5" mt="1.5" fontSize="12px" color="aurora.magenta">
										<LuTriangleAlert size={12} /> {emailError}
									</HStack>
								) : (
									<Text fontSize="12px" color="fg.muted" mt="2">
										The invite is addressed to this email; only this address can accept.
									</Text>
								)}
							</Box>
						</Box>
					</Box>

					{/* App-scoped role assignments */}
					<Box {...PANEL_PROPS}>
						<PanelHead
							icon={<LuUserPlus size={17} />}
							title="Roles to grant on acceptance"
							sub="pre-assigned now — the invitee sees these on the acceptance page; add as many apps as needed"
						/>
						<Box p="5" display="grid" gap="4">
							<AppScopedRoleAssignments
								apps={apps}
								value={assignments}
								onChange={(next) => {
									setAssignments(next);
									if (assignmentError) setAssignmentError(null);
								}}
							/>

							{assignmentError && (
								<HStack gap="1.5" fontSize="12px" color="aurora.magenta">
									<LuTriangleAlert size={12} /> {assignmentError}
								</HStack>
							)}

							<HStack gap="2" fontSize="12px" color="fg.muted" align="flex-start">
								<Box mt="0.5" flex="none">
									<LuLock size={13} />
								</Box>
								<Text>
									Inviting with the{" "}
									<Text as="span" color="aurora.amber" fontWeight="semibold">
										admin
									</Text>{" "}
									role on the{" "}
									<Text as="span" color="fg.subtle">
										isme
									</Text>{" "}
									app makes this user a platform administrator on acceptance — the replacement for the removed is_admin flag.
								</Text>
							</HStack>
						</Box>

						<Flex align="center" justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
							<HStack mr="auto" gap="2" fontSize="12px" color="fg.muted">
								<LuInfo size={13} /> {completeAssignments.length} assignment{completeAssignments.length === 1 ? "" : "s"} · link expires in 7 days
							</HStack>
							<Button
								variant="outline"
								h="11"
								px="4.5"
								borderRadius="glassSm"
								borderColor="border.strong"
								bg="bg.glass"
								fontSize="sm"
								fontWeight="semibold"
								color="fg"
								_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
								onClick={() => navigate("/users")}
							>
								Cancel
							</Button>
							<Button
								h="11"
								px="4.5"
								borderRadius="glassSm"
								fontSize="sm"
								fontWeight="semibold"
								color="white"
								css={AURORA_CTA_STYLE}
								boxShadow="ctaGlow"
								_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
								_focusVisible={{ boxShadow: "focusRing" }}
								disabled={!canSubmit}
								loading={submitting}
								onClick={handleSubmit}
							>
								<LuSend size={15} /> Send invite
								<LuChevronRight size={15} />
							</Button>
						</Flex>
					</Box>
				</>
			)}

			{/* One-time invite-link success — no ✕ / backdrop dismiss; Done-gated by
			    the ack so the link can't be lost by accident. */}
			<Dialog.Root open={inviteLink !== null} onOpenChange={() => {}} placement="center" closeOnInteractOutside={false} closeOnEscape={false}>
				<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
				<Dialog.Positioner>
					<Dialog.Content
						w="480px"
						maxW="92vw"
						borderRadius="20px"
						borderWidth="1px"
						borderColor="rgba(34,211,238,0.35)"
						bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
						color="fg"
						boxShadow="0 20px 60px rgba(34,211,238,0.20), 0 4px 12px rgba(0,0,0,0.35)"
						overflow="hidden"
					>
						<Dialog.Header px="5" py="4" display="flex" alignItems="center" borderBottomWidth="1px" borderColor="border">
							<HStack gap="3">
								<Center w="9" h="9" borderRadius="11px" bg="rgba(52,211,153,0.14)" borderWidth="1px" borderColor="rgba(52,211,153,0.35)" color="aurora.mint">
									<LuCheck size={16} />
								</Center>
								<Dialog.Title fontSize="15px" fontWeight="semibold">
									Invite created
								</Dialog.Title>
							</HStack>
						</Dialog.Header>
						<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
							<Box>
								<Text fontSize="13px" fontWeight="medium" color="fg.subtle" mb="2">
									Invite link for{" "}
									<Text as="span" color="fg" fontWeight="semibold">
										{email}
									</Text>{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										·{" "}
										<Text as="span" color="aurora.amber" fontWeight="semibold">
											shown only once
										</Text>
									</Text>
								</Text>
								<Flex align="center" gap="2.5">
									<Box
										flex="1"
										minW="0"
										h="52px"
										px="4"
										lineHeight="52px"
										borderRadius="glassSm"
										borderWidth="1px"
										borderColor="rgba(34,211,238,0.40)"
										bg="rgba(7,7,26,0.65)"
										color="aurora.cyan"
										fontSize="13px"
										fontWeight="medium"
										whiteSpace="nowrap"
										overflow="hidden"
										textOverflow="ellipsis"
										title={inviteLink ?? undefined}
										boxShadow="0 0 24px rgba(34,211,238,0.18), 0 0 0 4px rgba(34,211,238,0.06)"
										css={{ userSelect: "all" }}
									>
										{inviteLink}
									</Box>
									<Center
										as="button"
										w="52px"
										h="52px"
										flex="none"
										borderRadius="glassSm"
										cursor="pointer"
										bg="bg.glass"
										borderWidth="1px"
										borderColor={copied ? "rgba(52,211,153,0.40)" : "border.strong"}
										color={copied ? "aurora.mint" : "fg.subtle"}
										css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
										_hover={{ bg: "bg.glassHi", color: copied ? "aurora.mint" : "fg" }}
										title="Copy invite link"
										aria-label="Copy invite link"
										onClick={handleInviteLinkCopy}
									>
										{copied ? <LuCheck size={18} /> : <LuCopy size={18} />}
									</Center>
								</Flex>
							</Box>
							<HStack gap="2" px="3.5" py="2.5" borderRadius="glassSm" borderWidth="1px" borderColor="rgba(245,158,11,0.30)" bg="rgba(245,158,11,0.07)" fontSize="12px" color="aurora.amber" align="flex-start">
								<Box mt="0.5" flex="none">
									<LuTriangleAlert size={14} />
								</Box>
								<Box>
									This link is shown <b>only once</b> — the token is hashed at rest and cannot be displayed again. Share it with the invitee now.
								</Box>
							</HStack>
							<HStack gap="2" px="3.5" py="2.5" borderRadius="glassSm" borderWidth="1px" borderColor="border" bg="bg.glass" fontSize="12px" color="fg.muted" align="flex-start">
								<Box mt="0.5" flex="none">
									<LuClock size={14} />
								</Box>
								<Box>Expires in 7 days · single use — consumed when the account is created.</Box>
							</HStack>
							<Checkbox
								size="sm"
								colorPalette="purple"
								checked={acknowledged}
								onCheckedChange={(details) => setAcknowledged(!!details.checked)}
							>
								<Text fontSize="sm" color="fg.subtle">
									I have copied the link
								</Text>
							</Checkbox>
						</Dialog.Body>
						<HStack justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
							<Button
								h="11"
								px="4.5"
								borderRadius="glassSm"
								fontSize="sm"
								fontWeight="semibold"
								color="white"
								css={AURORA_CTA_STYLE}
								boxShadow="ctaGlow"
								_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
								disabled={!acknowledged}
								onClick={() => {
									setInviteLink(null);
									setAcknowledged(false);
									setCopied(false);
									navigate("/users");
								}}
							>
								Done
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>
		</AppShell>
	);
};
