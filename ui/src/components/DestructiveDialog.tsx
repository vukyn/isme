"use client";

import type { ReactNode, RefObject } from "react";
import { Box, Button, Center, Dialog, HStack, Stack, Text } from "@chakra-ui/react";
import { LuX } from "react-icons/lu";

/**
 * Shared shell for the destructive confirm dialogs (aurora mock
 * demo/aurora-user-delete-design.html): magenta frame for a pending
 * destructive action, neutral frame for an outcome that needs no action.
 */

export type DialogTone = "danger" | "neutral";
export type IconTone = "magenta" | "amber" | "mint";

const ICON_TONES: Record<IconTone, { bg: string; border: string; color: string }> = {
	magenta: { bg: "rgba(236,72,153,0.12)", border: "rgba(236,72,153,0.40)", color: "aurora.magenta" },
	amber: { bg: "rgba(245,158,11,0.12)", border: "rgba(245,158,11,0.40)", color: "aurora.amber" },
	mint: { bg: "rgba(52,211,153,0.12)", border: "rgba(52,211,153,0.40)", color: "aurora.mint" },
};

interface DestructiveDialogProps {
	open: boolean;
	onClose: () => void;
	/** While busy the dialog cannot be dismissed (requests are in flight). */
	busy: boolean;
	tone: DialogTone;
	wide?: boolean;
	icon: ReactNode;
	iconTone: IconTone;
	title: ReactNode;
	/** Element that receives focus when the dialog opens (Cancel by default). */
	initialFocusRef: RefObject<HTMLElement | null>;
	children: ReactNode;
	footer: ReactNode;
}

export const DestructiveDialog = ({
	open,
	onClose,
	busy,
	tone,
	wide,
	icon,
	iconTone,
	title,
	initialFocusRef,
	children,
	footer,
}: DestructiveDialogProps) => {
	const iconStyle = ICON_TONES[iconTone];
	return (
		<Dialog.Root
			role="alertdialog"
			open={open}
			placement="center"
			closeOnEscape={!busy}
			closeOnInteractOutside={!busy}
			initialFocusEl={() => initialFocusRef.current}
			onOpenChange={(details) => {
				if (!details.open && !busy) onClose();
			}}
		>
			<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
			<Dialog.Positioner px="4">
				<Dialog.Content
					w={wide ? "540px" : "480px"}
					maxW="92vw"
					maxH="calc(100vh - 32px)"
					borderRadius="20px"
					borderWidth="1px"
					borderColor={tone === "danger" ? "rgba(236,72,153,0.40)" : "border.strong"}
					bg="linear-gradient(180deg, rgba(18,18,46,0.96), rgba(11,11,35,0.97))"
					color="fg"
					boxShadow={
						tone === "danger"
							? "0 20px 60px rgba(236,72,153,0.20), 0 4px 12px rgba(0,0,0,0.35)"
							: "0 20px 60px rgba(99,102,241,0.25), 0 4px 12px rgba(0,0,0,0.35)"
					}
					overflow="hidden"
					display="flex"
					flexDirection="column"
					aria-busy={busy || undefined}
				>
					<Dialog.Header px="5" py="4" display="flex" alignItems="center" justifyContent="space-between" gap="3" borderBottomWidth="1px" borderColor="border">
						<HStack gap="3" minW="0">
							<Center
								w="9"
								h="9"
								flex="none"
								borderRadius="11px"
								bg={iconStyle.bg}
								borderWidth="1px"
								borderColor={iconStyle.border}
								color={iconStyle.color}
							>
								{icon}
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold" lineHeight="1.3" css={{ fontVariantNumeric: "tabular-nums" }}>
								{title}
							</Dialog.Title>
						</HStack>
						<Button
							variant="ghost"
							size="xs"
							w="8"
							h="8"
							p="0"
							minW="auto"
							flex="none"
							borderRadius="9px"
							color="fg.muted"
							_hover={{ bg: "bg.glass", color: "fg" }}
							disabled={busy}
							aria-label="Close"
							onClick={onClose}
						>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4" overflowY="auto" minH="0">
						{children}
					</Dialog.Body>
					<HStack justify="flex-end" align="center" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
						{footer}
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};

/** Mock .facts — what the action does, one consequence per line. */
export const FactList = ({ children }: { children: ReactNode }) => (
	<Stack as="ul" gap="10px" m="0" p="0" listStyleType="none">
		{children}
	</Stack>
);

export const Fact = ({ icon, children }: { icon: ReactNode; children: ReactNode }) => (
	<Box
		as="li"
		display="grid"
		gridTemplateColumns="16px 1fr"
		gap="10px"
		fontSize="13px"
		color="fg.subtle"
		lineHeight="1.55"
		css={{ "& b": { color: "var(--chakra-colors-fg)", fontWeight: 600, fontVariantNumeric: "tabular-nums" } }}
	>
		<Box mt="3px" color="fg.muted">
			{icon}
		</Box>
		<Text as="span">{children}</Text>
	</Box>
);

const NOTE_TONES = {
	amber: { border: "rgba(245,158,11,0.30)", bg: "rgba(245,158,11,0.07)", color: "#FCD9A0", icon: "aurora.amber" },
	red: { border: "rgba(236,72,153,0.35)", bg: "rgba(236,72,153,0.08)", color: "#FBC6DE", icon: "aurora.magenta" },
	violet: { border: "rgba(139,92,246,0.35)", bg: "rgba(139,92,246,0.08)", color: "#D9CCFF", icon: "aurora.violet" },
} as const;

/** Mock .note — tinted inline callout. `title` renders the bold first line. */
export const Note = ({
	tone,
	icon,
	title,
	children,
	alertRef,
}: {
	tone: keyof typeof NOTE_TONES;
	icon: ReactNode;
	title?: ReactNode;
	children: ReactNode;
	/** Set on error notes: the note becomes role="alert" and can take focus. */
	alertRef?: (node: HTMLDivElement | null) => void;
}) => {
	const style = NOTE_TONES[tone];
	return (
		<HStack
			ref={alertRef}
			role={alertRef ? "alert" : undefined}
			tabIndex={alertRef ? -1 : undefined}
			align="flex-start"
			gap="10px"
			px="14px"
			py="12px"
			borderRadius="glassSm"
			borderWidth="1px"
			borderColor={style.border}
			bg={style.bg}
			color={style.color}
			fontSize="12px"
			lineHeight="1.55"
			outline="none"
			css={{ "& b": { color: "var(--chakra-colors-fg)", fontWeight: 600, fontVariantNumeric: "tabular-nums" } }}
		>
			<Box flex="none" mt="2px" color={style.icon}>
				{icon}
			</Box>
			<Box minW="0">
				{title && (
					<Text fontSize="13px" fontWeight="semibold" color="fg" mb="2px">
						{title}
					</Text>
				)}
				<Text as="span">{children}</Text>
			</Box>
		</HStack>
	);
};

/** Mock .lead — the body of an outcome view. */
export const Lead = ({ children }: { children: ReactNode }) => (
	<Text m="0" fontSize="14px" color="fg.subtle" lineHeight="1.6" css={{ "& b": { color: "var(--chakra-colors-fg)", fontWeight: 600 } }}>
		{children}
	</Text>
);
