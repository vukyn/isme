"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, Field, HStack, Input, NativeSelect, Text } from "@chakra-ui/react";
import { LuKeyRound, LuPlus, LuType, LuX } from "react-icons/lu";
import { createRole } from "@/apis";
import { ColorSwatchPicker } from "@/components/ColorSwatchPicker";
import { IconPicker } from "@/components/IconPicker";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { RoleListItem } from "@/types";

/** Sensible appearance defaults for a brand-new role. */
const DEFAULT_ROLE_ICON = "key";
const DEFAULT_ROLE_COLOR = "violet";

interface CreateRoleDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Owning app for the new role (app-owned RBAC) — bound, not editable. */
	appId: string;
	appCode: string;
	/** Existing roles (same app) — clone-from options. */
	roles: RoleListItem[];
	/** Called with the new role id after a successful create. */
	onCreated: (roleId: string) => void;
}

const ROLE_CODE_PATTERN = /^[a-z0-9][a-z0-9-]*$/;

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

export const CreateRoleDialog = ({ open, onOpenChange, appId, appCode, roles, onCreated }: CreateRoleDialogProps) => {
	const [code, setCode] = useState("");
	const [name, setName] = useState("");
	const [description, setDescription] = useState("");
	const [cloneFromRoleId, setCloneFromRoleId] = useState("");
	const [icon, setIcon] = useState(DEFAULT_ROLE_ICON);
	const [color, setColor] = useState(DEFAULT_ROLE_COLOR);
	const [codeError, setCodeError] = useState<string | null>(null);
	const [nameError, setNameError] = useState<string | null>(null);
	const [submitting, setSubmitting] = useState(false);

	const resetForm = () => {
		setCode("");
		setName("");
		setDescription("");
		setCloneFromRoleId("");
		setIcon(DEFAULT_ROLE_ICON);
		setColor(DEFAULT_ROLE_COLOR);
		setCodeError(null);
		setNameError(null);
	};

	const handleSubmit = async () => {
		const trimmedCode = code.trim();
		const trimmedName = name.trim();
		let invalid = false;
		if (!trimmedCode) {
			setCodeError("Code is required");
			invalid = true;
		} else if (!ROLE_CODE_PATTERN.test(trimmedCode)) {
			setCodeError("Lowercase slug only (a-z, 0-9, hyphen)");
			invalid = true;
		} else {
			setCodeError(null);
		}
		if (!trimmedName) {
			setNameError("Name is required");
			invalid = true;
		} else {
			setNameError(null);
		}
		if (invalid) return;

		setSubmitting(true);
		try {
			const created = await createRole({
				app_id: appId,
				code: trimmedCode,
				name: trimmedName,
				description: description.trim(),
				clone_from_role_id: cloneFromRoleId || undefined,
				icon,
				color,
			});
			toaster.create({ title: `Role ${trimmedCode} created`, type: "success", meta: { closable: true } });
			resetForm();
			onOpenChange(false);
			onCreated(created.id);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } }; message?: string };
			toaster.create({
				title: err?.response?.data?.message || err?.message || "Failed to create role",
				type: "error",
				meta: { closable: true },
			});
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
								<LuKeyRound size={16} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								New role in{" "}
								<Text as="span" color="aurora.violet">
									{appCode || "—"}
								</Text>
							</Dialog.Title>
						</HStack>
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => onOpenChange(false)}>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>
								App{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· bound to the selected app, not editable
								</Text>
							</Field.Label>
							<Input {...INPUT_PROPS} value={appCode} readOnly opacity={0.7} aria-label="Owning app (locked)" />
						</Field.Root>
						<Field.Root invalid={!!codeError}>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Code{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· unique slug, immutable after create
								</Text>
							</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuType size={16} />
								</Box>
								<Input
									{...INPUT_PROPS}
									pl="10"
									letterSpacing="0.02em"
									placeholder="support-agent"
									value={code}
									onChange={(event) => setCode(event.target.value)}
								/>
							</Box>
							{codeError && <Field.ErrorText>{codeError}</Field.ErrorText>}
						</Field.Root>
						<Field.Root invalid={!!nameError}>
							<Field.Label {...FIELD_LABEL_PROPS}>Name</Field.Label>
							<Input {...INPUT_PROPS} placeholder="Support Agent" value={name} onChange={(event) => setName(event.target.value)} />
							{nameError && <Field.ErrorText>{nameError}</Field.ErrorText>}
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>Description</Field.Label>
							<Input
								{...INPUT_PROPS}
								placeholder="What this role is for…"
								value={description}
								onChange={(event) => setDescription(event.target.value)}
							/>
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Icon{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· role badge
								</Text>
							</Field.Label>
							<IconPicker value={icon} onChange={setIcon} ariaLabel="Role icon" />
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Color{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· aurora palette
								</Text>
							</Field.Label>
							<ColorSwatchPicker value={color} onChange={setColor} ariaLabel="Role color" />
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Clone permissions from{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· optional
								</Text>
							</Field.Label>
							<NativeSelect.Root size="sm" w="full">
								<NativeSelect.Field
									{...INPUT_PROPS}
									css={{ "& option": { background: "#12122E", color: "#F4F5FF" } }}
									value={cloneFromRoleId}
									onChange={(event) => setCloneFromRoleId(event.target.value)}
								>
									<option value="">— start empty —</option>
									{roles.map((role) => (
										<option key={role.id} value={role.id}>
											{role.code}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
							<Field.HelperText fontSize="12px" color="fg.muted">
								Copies the role_permission set; members and scope are not cloned.
							</Field.HelperText>
						</Field.Root>
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
							<LuPlus size={15} /> Create role
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
