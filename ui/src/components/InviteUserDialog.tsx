"use client";

import { useEffect, useState } from "react";
import { Box, Button, Center, Dialog, Field, Flex, HStack, Input, Text } from "@chakra-ui/react";
import { LuCheck, LuClock, LuCopy, LuInfo, LuLink, LuMail, LuTriangleAlert, LuUserPlus, LuX } from "react-icons/lu";
import { createInvitation, listAppServices } from "@/apis";
import { AppScopedRoleAssignments } from "@/components/AppScopedRoleAssignments";
import { Checkbox } from "@/components/ui/checkbox";
import { toaster } from "@/components/ui/toaster";
import { APP_SERVICE_STATUS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { AppService, InvitationRoleAssignment } from "@/types";
import { copyToClipboard } from "@/utils";
import { inviteUserSchema } from "@/validators";

interface InviteUserDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Fired when the admin closes the success state — refetch the invitations list. */
	onInvited: () => void;
	/** Prefill for the "New invite" re-issue action on revoked/expired rows. */
	prefill?: { email: string; assignments: InvitationRoleAssignment[] } | null;
}

const FIELD_LABEL_PROPS = {
	fontSize: "13px",
	fontWeight: "medium",
	color: "fg.subtle",
} as const;

const INPUT_PROPS = {
	h: "11",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "sm",
	color: "fg",
	_placeholder: { color: "fg.muted" },
	_hover: { borderColor: "rgba(255,255,255,0.28)" },
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" },
} as const;

const CHECKBOX_CSS = {
	"& .chakra-checkbox__control": {
		borderRadius: "5px",
		borderColor: "rgba(255,255,255,0.18)",
		background: "rgba(255,255,255,0.06)",
	},
	"& .chakra-checkbox__control[data-state=checked]": {
		background: "linear-gradient(135deg, #6366F1, #8B5CF6)",
		borderColor: "#8B5CF6",
		boxShadow: "0 0 12px rgba(139,92,246,0.45)",
	},
} as const;

interface CreatedInvite {
	email: string;
	assignmentCount: number;
	link: string;
	expiresAt: Date;
}

/**
 * One dialog, two states: the form (state A) swaps to the one-time link
 * success view (state B) after POST succeeds. State B mirrors the
 * app-service secret dialog: no ✕ and no backdrop dismiss — the only exit
 * is the explicit Done button, gated by the ack checkbox, so the link
 * can't be lost by accident.
 */
export const InviteUserDialog = ({ open, onOpenChange, onInvited, prefill }: InviteUserDialogProps) => {
	const [email, setEmail] = useState("");
	const [apps, setApps] = useState<AppService[]>([]);
	const [assignments, setAssignments] = useState<InvitationRoleAssignment[]>([]);
	const [emailError, setEmailError] = useState<string | null>(null);
	const [assignmentError, setAssignmentError] = useState<string | null>(null);
	const [submitting, setSubmitting] = useState(false);

	const [created, setCreated] = useState<CreatedInvite | null>(null);
	const [acknowledged, setAcknowledged] = useState(false);
	const [copied, setCopied] = useState(false);

	// Load active apps when the dialog opens; apply the re-issue prefill or seed
	// one empty assignment row.
	useEffect(() => {
		if (!open) return;
		setEmail(prefill?.email ?? "");
		setEmailError(null);
		setAssignmentError(null);
		let active = true;
		listAppServices({ page: 1, page_size: 100, status: APP_SERVICE_STATUS.ACTIVE })
			.then((response) => {
				if (!active) return;
				const items = response.items;
				setApps(items);
				if (prefill?.assignments.length) {
					setAssignments(prefill.assignments);
				} else {
					const first = items[0];
					setAssignments(first ? [{ app_service_id: first.id, role_id: "" }] : []);
				}
			})
			.catch(() => {
				if (active) toaster.create({ title: "Failed to load apps", type: "error", meta: { closable: true } });
			});
		return () => {
			active = false;
		};
	}, [open, prefill]);

	const resetForm = () => {
		setEmail("");
		setAssignments([]);
		setEmailError(null);
		setAssignmentError(null);
		setCreated(null);
		setAcknowledged(false);
		setCopied(false);
	};

	const handleSubmit = async () => {
		const completeAssignments = assignments.filter(
			(assignment) => assignment.app_service_id && assignment.role_id
		);
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
			setCreated({
				email: parsed.data.email,
				assignmentCount: parsed.data.assignments.length,
				link: `${window.location.origin}${response.invite_link}`,
				expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
			});
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			const message = err?.response?.data?.message;
			if (message) setEmailError(message);
			else toaster.create({ title: "Failed to create invite", type: "error", meta: { closable: true } });
		} finally {
			setSubmitting(false);
		}
	};

	const handleCopy = async () => {
		if (!created) return;
		try {
			await copyToClipboard(created.link);
			setCopied(true);
			toaster.create({ title: "Invite link copied", type: "success", meta: { closable: true } });
			setTimeout(() => setCopied(false), 2000);
		} catch {
			toaster.create({ title: "Failed to copy — select the link and copy manually", type: "error", meta: { closable: true } });
		}
	};

	const handleDone = () => {
		resetForm();
		onInvited();
		onOpenChange(false);
	};

	const handleCancel = () => {
		resetForm();
		onOpenChange(false);
	};

	const success = created !== null;

	return (
		<Dialog.Root
			open={open}
			onOpenChange={(details) => {
				// state B is exit-by-Done only
				if (success) return;
				if (!details.open) resetForm();
				onOpenChange(details.open);
			}}
			placement="center"
			closeOnInteractOutside={!success}
			closeOnEscape={!success}
		>
			<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
			<Dialog.Positioner>
				<Dialog.Content
					w={success ? "480px" : "560px"}
					maxW="92vw"
					borderRadius="20px"
					borderWidth="1px"
					borderColor={success ? "rgba(34,211,238,0.35)" : "border.strong"}
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					color="fg"
					boxShadow={success ? "0 20px 60px rgba(34,211,238,0.20), 0 4px 12px rgba(0,0,0,0.35)" : "glassPop"}
					overflow="hidden"
				>
					{success ? (
						<>
							{/* Intentionally no ✕ — the link is shown only once. */}
							<Dialog.Header px="5" py="4" display="flex" alignItems="center" borderBottomWidth="1px" borderColor="border">
								<HStack gap="3">
									<Center
										w="9"
										h="9"
										borderRadius="11px"
										bg="rgba(52,211,153,0.14)"
										borderWidth="1px"
										borderColor="rgba(52,211,153,0.35)"
										color="aurora.mint"
									>
										<LuCheck size={16} />
									</Center>
									<Dialog.Title fontSize="15px" fontWeight="semibold">
										Invite created
									</Dialog.Title>
									<Box
										display="inline-flex"
										alignItems="center"
										px="11px"
										py="1"
										borderRadius="full"
										fontSize="12px"
										fontWeight="medium"
										color="aurora.cyan"
										borderWidth="1px"
										borderColor="rgba(34,211,238,0.40)"
										bg="rgba(34,211,238,0.10)"
										whiteSpace="nowrap"
									>
										{created.assignmentCount} role{created.assignmentCount === 1 ? "" : "s"}
									</Box>
								</HStack>
							</Dialog.Header>
							<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
								<Box>
									<Text fontSize="13px" fontWeight="medium" color="fg.subtle" mb="2">
										Invite link for{" "}
										<Text as="span" color="fg" fontWeight="semibold">
											{created.email}
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
											lineHeight="52px"
											title={created.link}
											boxShadow="0 0 24px rgba(34,211,238,0.18), 0 0 0 4px rgba(34,211,238,0.06)"
											css={{ userSelect: "all" }}
										>
											{created.link}
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
											_hover={{ bg: "bg.glassHi", color: copied ? "aurora.mint" : "fg", borderColor: copied ? "rgba(52,211,153,0.40)" : "rgba(255,255,255,0.30)" }}
											title="Copy invite link"
											aria-label="Copy invite link"
											onClick={handleCopy}
										>
											{copied ? <LuCheck size={18} /> : <LuCopy size={18} />}
										</Center>
									</Flex>
								</Box>
								<HStack
									gap="2"
									px="3.5"
									py="2.5"
									borderRadius="glassSm"
									borderWidth="1px"
									borderColor="rgba(245,158,11,0.30)"
									bg="rgba(245,158,11,0.07)"
									fontSize="12px"
									color="aurora.amber"
									alignItems="flex-start"
								>
									<Box mt="0.5" flex="none">
										<LuTriangleAlert size={14} />
									</Box>
									<Box>
										This link is shown <b>only once</b> — the token is hashed at rest and cannot be displayed again. Share it with the invitee now. If lost, revoke and create a new invite.
									</Box>
								</HStack>
								<HStack
									gap="2"
									px="3.5"
									py="2.5"
									borderRadius="glassSm"
									borderWidth="1px"
									borderColor="border"
									bg="bg.glass"
									fontSize="12px"
									color="fg.muted"
									alignItems="flex-start"
								>
									<Box mt="0.5" flex="none">
										<LuClock size={14} />
									</Box>
									<Box>
										Expires{" "}
										<Text as="span" color="aurora.amber" fontWeight="semibold">
											{created.expiresAt.toLocaleDateString()} {created.expiresAt.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
										</Text>{" "}
										(7 days) · single use — consumed when the account is created
									</Box>
								</HStack>
								<Checkbox
									size="sm"
									colorPalette="purple"
									css={CHECKBOX_CSS}
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
									onClick={handleDone}
								>
									Done
								</Button>
							</HStack>
						</>
					) : (
						<>
							<Dialog.Header px="5" py="4" display="flex" alignItems="center" justifyContent="space-between" borderBottomWidth="1px" borderColor="border">
								<HStack gap="3">
									<Center
										w="9"
										h="9"
										borderRadius="11px"
										bg="linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))"
										borderWidth="1px"
										borderColor="border.strong"
										color="aurora.cyan"
									>
										<LuUserPlus size={16} />
									</Center>
									<Dialog.Title fontSize="15px" fontWeight="semibold">
										Invite user
									</Dialog.Title>
								</HStack>
								<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={handleCancel}>
									<LuX size={16} />
								</Button>
							</Dialog.Header>
							<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
								<Field.Root invalid={!!emailError}>
									<Field.Label {...FIELD_LABEL_PROPS}>Email</Field.Label>
									<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
										<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
											<LuMail size={16} />
										</Box>
										<Input {...INPUT_PROPS} pl="10" placeholder="name@example.com" value={email} onChange={(event) => setEmail(event.target.value)} />
									</Box>
									{emailError && <Field.ErrorText>{emailError}</Field.ErrorText>}
								</Field.Root>
								<Field.Root invalid={!!assignmentError}>
									<Field.Label {...FIELD_LABEL_PROPS}>Roles to grant</Field.Label>
									<AppScopedRoleAssignments
										apps={apps}
										value={assignments}
										onChange={(next) => {
											setAssignments(next);
											if (assignmentError) setAssignmentError(null);
										}}
									/>
									{assignmentError && <Field.ErrorText>{assignmentError}</Field.ErrorText>}
								</Field.Root>
								<HStack
									gap="2"
									px="3.5"
									py="2.5"
									borderRadius="glassSm"
									borderWidth="1px"
									borderColor="border"
									bg="bg.glass"
									fontSize="12px"
									color="fg.muted"
								>
									<LuInfo size={14} style={{ flex: "0 0 auto" }} />
									<Box>
										No email is sent. You get a{" "}
										<Text as="span" color="aurora.amber" fontWeight="semibold">
											one-time link
										</Text>{" "}
										to share yourself · expires after 7 days
									</Box>
								</HStack>
							</Dialog.Body>
							<HStack justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
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
									onClick={handleCancel}
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
									loading={submitting}
									onClick={handleSubmit}
								>
									<LuLink size={15} /> Create invite
								</Button>
							</HStack>
						</>
					)}
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
