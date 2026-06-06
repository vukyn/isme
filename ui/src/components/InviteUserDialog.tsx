"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, Field, HStack, Input, NativeSelect, Text } from "@chakra-ui/react";
import { LuInfo, LuMail, LuSend, LuUser, LuUserPlus, LuX } from "react-icons/lu";
import { inviteUser } from "@/apis";
import { Checkbox } from "@/components/ui/checkbox";
import { toaster } from "@/components/ui/toaster";
import { USER_ROLE_OPTIONS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { UserListItem } from "@/types";
import { inviteUserSchema } from "@/validators";

interface InviteUserDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onInvited: (user: UserListItem) => void;
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

export const InviteUserDialog = ({ open, onOpenChange, onInvited }: InviteUserDialogProps) => {
	const [email, setEmail] = useState("");
	const [name, setName] = useState("");
	const [role, setRole] = useState("member");
	const [isAdmin, setIsAdmin] = useState(false);
	const [emailError, setEmailError] = useState<string | null>(null);
	const [submitting, setSubmitting] = useState(false);

	const resetForm = () => {
		setEmail("");
		setName("");
		setRole("member");
		setIsAdmin(false);
		setEmailError(null);
	};

	const handleSubmit = async () => {
		const parsed = inviteUserSchema.safeParse({ email, name: name || undefined, role, is_admin: isAdmin });
		if (!parsed.success) {
			setEmailError(parsed.error.issues[0]?.message ?? "Invalid input");
			return;
		}
		setEmailError(null);
		setSubmitting(true);
		try {
			const invited = await inviteUser(parsed.data);
			onInvited(invited);
			toaster.create({ title: `Invite sent to ${parsed.data.email}`, type: "success", meta: { closable: true } });
			resetForm();
			onOpenChange(false);
		} catch {
			toaster.create({ title: "Failed to send invite", type: "error", meta: { closable: true } });
		} finally {
			setSubmitting(false);
		}
	};

	return (
		<Dialog.Root open={open} onOpenChange={(details) => onOpenChange(details.open)} placement="center">
			<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
			<Dialog.Positioner>
				<Dialog.Content
					w="480px"
					maxW="92vw"
					borderRadius="20px"
					borderWidth="1px"
					borderColor="border.strong"
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					color="fg"
					boxShadow="glassPop"
					overflow="hidden"
				>
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
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => onOpenChange(false)}>
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
								<Input {...INPUT_PROPS} pl="10" placeholder="name@hasaki.vn" value={email} onChange={(event) => setEmail(event.target.value)} />
							</Box>
							{emailError && <Field.ErrorText>{emailError}</Field.ErrorText>}
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Name{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· optional, user can edit on first login
								</Text>
							</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuUser size={16} />
								</Box>
								<Input {...INPUT_PROPS} pl="10" placeholder="Full name" value={name} onChange={(event) => setName(event.target.value)} />
							</Box>
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>Role</Field.Label>
							<NativeSelect.Root size="sm" w="full">
								<NativeSelect.Field
									{...INPUT_PROPS}
									css={{ "& option": { background: "#12122E", color: "#F4F5FF" } }}
									value={role}
									onChange={(event) => setRole(event.target.value)}
								>
									{USER_ROLE_OPTIONS.map((option) => (
										<option key={option} value={option}>
											{option}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Field.Root>
						<Checkbox
							size="sm"
							colorPalette="purple"
							css={CHECKBOX_CSS}
							checked={isAdmin}
							onCheckedChange={(details) => setIsAdmin(!!details.checked)}
						>
							<HStack gap="2" fontSize="sm" color="fg.subtle">
								<Text>Grant administrator</Text>
								<Text fontSize="12px" color="fg.muted">
									· sets is_admin = true
								</Text>
							</HStack>
						</Checkbox>
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
								Invite link expires after 72h · account stays{" "}
								<Text as="span" color="aurora.amber" fontWeight="semibold">
									pending
								</Text>{" "}
								until password is set
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
							onClick={() => onOpenChange(false)}
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
							<LuSend size={15} /> Send invite
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
