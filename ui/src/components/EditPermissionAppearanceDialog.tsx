"use client";

import { useEffect, useMemo, useState } from "react";
import { Box, Button, Center, Dialog, HStack, Text } from "@chakra-ui/react";
import { LuCheck, LuInfo, LuPencil } from "react-icons/lu";
import { updatePermissionAppearance } from "@/apis";
import { ColorSwatchPicker } from "@/components/ColorSwatchPicker";
import { IconPicker } from "@/components/IconPicker";
import { toaster } from "@/components/ui/toaster";
import { APP_COLORS, renderPermissionIcon, resolveResourceIconKey } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";

const FIELD_LABEL_PROPS = {
	display: "block",
	fontSize: "13px",
	fontWeight: "medium",
	color: "fg.subtle",
	mb: "8px",
} as const;

interface EditPermissionAppearanceDialogProps {
	/** Controls visibility. */
	open: boolean;
	/** Fired with the next open state (false = dialog requests close). */
	onOpenChange: (open: boolean) => void;
	/** Owning app id (the appearance is scoped to (app_id, resource)). */
	appId: string;
	/** The resource whose icon + color are edited (shared by all its actions). */
	resource: string;
	/** Human label for the resource (Title Case). */
	label: string;
	/** The RAW stored icon key (empty when unset) — seeds the picker on open. */
	savedIcon: string;
	/** The RAW stored color key (empty when unset) — seeds the picker on open. */
	savedColor: string;
	/** Whether the signed-in user may save (role:update). */
	canEdit: boolean;
	/** Called after a successful save (parent refetches the catalog). */
	onSaved: () => void | Promise<void>;
}

/**
 * Edits a permission RESOURCE's appearance (icon + color) — shared by every
 * resource:action row of an (app_id, resource). Mirrors EditAppService's
 * icon/color form body (inline IconPicker + ColorSwatchPicker + live preview)
 * inside the catalog's glass dialog shell. The pickers are allowlist-only with
 * NO clear/allowClear: a resource is neutral only if it was created without an
 * icon/color. Seeds from the RAW stored keys so an unset resource opens neutral.
 */
export const EditPermissionAppearanceDialog = ({
	open,
	onOpenChange,
	appId,
	resource,
	label,
	savedIcon,
	savedColor,
	canEdit,
	onSaved,
}: EditPermissionAppearanceDialogProps) => {
	const [icon, setIcon] = useState(savedIcon);
	const [color, setColor] = useState(savedColor);
	const [saving, setSaving] = useState(false);

	// Re-seed from the raw stored keys each time the dialog opens (or targets a
	// different resource) so a previous edit can't leak into the next one.
	useEffect(() => {
		if (open) {
			setIcon(savedIcon);
			setColor(savedColor);
		}
	}, [open, resource, savedIcon, savedColor]);

	const dirty = icon !== savedIcon || color !== savedColor;
	const colorHex = useMemo(
		() => (color && color in APP_COLORS ? APP_COLORS[color as keyof typeof APP_COLORS].hex : ""),
		[color]
	);

	// Preview tile mirrors the catalog row's resource cell: a stored color wins,
	// otherwise a neutral surface (no cycling palette fallback in the dialog).
	const previewIconKey = resolveResourceIconKey(resource, icon);
	const previewAccent =
		color && color in APP_COLORS
			? {
					color: APP_COLORS[color as keyof typeof APP_COLORS].hex,
					bg: `rgba(${APP_COLORS[color as keyof typeof APP_COLORS].rgb},0.14)`,
				}
			: { color: "fg.subtle", bg: "rgba(255,255,255,0.06)" };

	const handleSave = async () => {
		if (!dirty || !canEdit) return;
		setSaving(true);
		try {
			await updatePermissionAppearance({ app_id: appId, resource, icon, color });
			toaster.create({ title: "Appearance saved", type: "success", meta: { closable: true } });
			await onSaved();
			onOpenChange(false);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({
				title: err?.response?.data?.message || "Failed to save appearance",
				type: "error",
				meta: { closable: true },
			});
		} finally {
			setSaving(false);
		}
	};

	return (
		<Dialog.Root open={open} onOpenChange={(details) => onOpenChange(details.open)} placement="center">
			<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
			<Dialog.Positioner>
				<Dialog.Content
					w="560px"
					maxW="92vw"
					borderRadius="20px"
					borderWidth="1px"
					borderColor="border.strong"
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					color="fg"
					boxShadow="glassPop"
					overflow="hidden"
				>
					<Dialog.Header
						px="5"
						py="4"
						display="flex"
						alignItems="center"
						gap="3"
						borderBottomWidth="1px"
						borderColor="border"
					>
						<Center
							w="32px"
							h="32px"
							flex="none"
							borderRadius="10px"
							color="aurora.cyan"
							borderWidth="1px"
							borderColor="border.strong"
							css={{ background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))" }}
						>
							<LuPencil size={15} />
						</Center>
						<Dialog.Title fontSize="15px" fontWeight="semibold">
							Edit appearance ·{" "}
							<Text as="span" color="fg.subtle">
								{resource}
							</Text>
						</Dialog.Title>
					</Dialog.Header>

					<Dialog.Body p="5" display="flex" flexDirection="column" gap="20px" maxH="70vh" overflowY="auto">
						{/* context line: scope of the change */}
						<HStack
							gap="9px"
							px="14px"
							py="11px"
							borderRadius="glassSm"
							borderWidth="1px"
							borderColor="border"
							bg="bg.glass"
							fontSize="12px"
							color="fg.muted"
							alignItems="flex-start"
						>
							<Box color="aurora.cyan" mt="1px" flex="none">
								<LuInfo size={15} />
							</Box>
							<Box>
								Appearance is stored per resource — this icon + color apply to{" "}
								<Text as="b" color="fg.subtle" fontWeight="semibold">
									all actions
								</Text>{" "}
								of{" "}
								<Text as="code" color="aurora.cyan" fontFamily="inherit">
									{resource}
								</Text>
								.
							</Box>
						</HStack>

						{/* ICON PICKER */}
						<Box>
							<Text {...FIELD_LABEL_PROPS}>
								Icon{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· pick from the shared icon set
								</Text>
							</Text>
							<IconPicker value={icon} onChange={setIcon} ariaLabel="Resource icon" />
							<Text mt="8px" fontSize="12px" color="fg.muted">
								Stored as a string key (
								<Text as="b" color="fg.subtle" fontWeight="semibold">
									{icon || "neutral"}
								</Text>
								). Same allowlist used by the permission catalog.
							</Text>
						</Box>

						{/* COLOR PICKER */}
						<Box>
							<Text {...FIELD_LABEL_PROPS}>
								Color{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· aurora palette
								</Text>
							</Text>
							<ColorSwatchPicker value={color} onChange={setColor} ariaLabel="Resource color" />
							<Text mt="10px" fontSize="12px" color="fg.muted">
								Stored as palette key{" "}
								<Text as="b" color="fg.subtle" fontWeight="semibold">
									{color || "neutral"}
								</Text>
								{colorHex && (
									<>
										{" "}
										·{" "}
										<Text as="span" css={{ fontVariantNumeric: "tabular-nums" }}>
											{colorHex}
										</Text>
									</>
								)}
							</Text>
						</Box>

						{/* LIVE PREVIEW — the catalog row's resource cell with the chosen pair */}
						<Box>
							<Text
								fontSize="11px"
								fontWeight="semibold"
								letterSpacing="0.12em"
								textTransform="uppercase"
								color="fg.muted"
								mb="10px"
							>
								Live preview
							</Text>
							<HStack
								gap="3"
								px="12px"
								py="10px"
								borderRadius="12px"
								borderWidth="1px"
								borderColor="border"
								bg="rgba(255,255,255,0.03)"
							>
								<Center
									w="8"
									h="8"
									flex="none"
									borderRadius="10px"
									bg={previewAccent.bg}
									borderWidth="1px"
									borderColor="border.strong"
									color={previewAccent.color}
								>
									{renderPermissionIcon(previewIconKey, 14)}
								</Center>
								<Box lineHeight="1.25" minW="0">
									<Text fontSize="sm" fontWeight="medium" color="fg" truncate>
										{label}
									</Text>
									<Text fontSize="11px" color="fg.muted">
										{resource}
									</Text>
								</Box>
							</HStack>
						</Box>
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
							_focusVisible={{ boxShadow: "focusRing" }}
							disabled={!dirty || !canEdit}
							loading={saving}
							onClick={handleSave}
						>
							<LuCheck size={15} /> Save appearance
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
