"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, HStack, Text } from "@chakra-ui/react";
import { LuBadgeCheck, LuCircleCheck, LuShieldQuestion, LuTriangleAlert, LuX } from "react-icons/lu";
import { verifyUser } from "@/apis";
import { toaster } from "@/components/ui/toaster";
import type { UserListItem } from "@/types";

interface VerifyUserDialogProps {
	open: boolean;
	/** Unverified accounts the confirmation targets (1 = row action, >1 = bulk). */
	targets: UserListItem[];
	/** Already-verified rows skipped at selection time (bulk toast reports them). */
	skippedCount: number;
	onOpenChange: (open: boolean) => void;
	/** Called with the verified ids after every POST succeeded. */
	onVerified: (ids: string[]) => void;
}

const AVATAR_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
	"linear-gradient(135deg, #F59E0B, #EC4899)",
	"linear-gradient(135deg, #34D399, #22D3EE)",
];

const initials = (name: string) =>
	name
		.split(" ")
		.filter(Boolean)
		.map((part) => part[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

const avatarGradient = (id: string) => {
	let hash = 0;
	for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) % 997;
	return AVATAR_GRADIENTS[hash % AVATAR_GRADIENTS.length];
};

const UnverifiedPill = () => (
	<Box
		display="inline-flex"
		alignItems="center"
		gap="7px"
		px="11px"
		py="1"
		borderRadius="full"
		fontSize="12px"
		fontWeight="medium"
		borderWidth="1px"
		whiteSpace="nowrap"
		color="aurora.amber"
		borderColor="rgba(245,158,11,0.35)"
		bg="rgba(245,158,11,0.10)"
	>
		<LuShieldQuestion size={12} /> unverified
	</Box>
);

/**
 * One-way verify confirmation (mock #verify-modal) — flips is_verified to
 * true and unblocks login. The bulk variant swaps the identity block for
 * "N unverified accounts selected".
 */
export const VerifyUserDialog = ({ open, targets, skippedCount, onOpenChange, onVerified }: VerifyUserDialogProps) => {
	const [submitting, setSubmitting] = useState(false);

	const single = targets.length === 1 ? targets[0] : null;

	const handleSubmit = async () => {
		if (targets.length === 0) return;
		setSubmitting(true);
		try {
			await Promise.all(targets.map((target) => verifyUser(target.id)));
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({
				title: err?.response?.data?.message || "Failed to verify account",
				type: "error",
				meta: { closable: true },
			});
			setSubmitting(false);
			return;
		}
		setSubmitting(false);
		onOpenChange(false);
		toaster.create({
			title: single
				? `${single.name || single.email} verified`
				: skippedCount > 0
					? `${targets.length} verified · ${skippedCount} skipped`
					: `${targets.length} account${targets.length > 1 ? "s" : ""} verified`,
			type: "success",
			meta: { closable: true },
		});
		onVerified(targets.map((target) => target.id));
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
					borderColor="rgba(52,211,153,0.40)"
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					color="fg"
					boxShadow="0 20px 60px rgba(52,211,153,0.18), 0 4px 12px rgba(0,0,0,0.35)"
					overflow="hidden"
				>
					<Dialog.Header px="5" py="4" display="flex" alignItems="center" justifyContent="space-between" borderBottomWidth="1px" borderColor="border">
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
								<LuBadgeCheck size={16} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Verify account{targets.length > 1 ? "s" : ""}
							</Dialog.Title>
						</HStack>
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => onOpenChange(false)}>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						{single ? (
							<HStack gap="3">
								<Center
									w="8"
									h="8"
									flex="none"
									borderRadius="full"
									fontSize="12px"
									fontWeight="bold"
									color="white"
									css={{ background: avatarGradient(single.id), boxShadow: "0 0 14px rgba(99,102,241,0.30)" }}
								>
									{initials(single.name)}
								</Center>
								<Box minW="0" lineHeight="1.25">
									<Text fontSize="sm" fontWeight="semibold" color="fg" truncate>
										{single.name || "—"}
									</Text>
									<Text fontSize="12px" color="fg.muted" truncate>
										{single.email}
									</Text>
								</Box>
								<Box ml="auto">
									<UnverifiedPill />
								</Box>
							</HStack>
						) : (
							<HStack gap="3">
								<Text fontSize="sm" fontWeight="semibold" color="fg" css={{ fontVariantNumeric: "tabular-nums" }}>
									{targets.length} unverified account{targets.length > 1 ? "s" : ""} selected
								</Text>
								<Box ml="auto">
									<UnverifiedPill />
								</Box>
							</HStack>
						)}
						<HStack
							gap="2.5"
							align="flex-start"
							px="3.5"
							py="3"
							borderRadius="12px"
							borderWidth="1px"
							borderColor="rgba(52,211,153,0.30)"
							bg="bg.glass"
							fontSize="13px"
							color="fg.subtle"
						>
							<Box color="aurora.mint" flex="none" mt="0.5">
								<LuCircleCheck size={14} />
							</Box>
							<Text>Marks the account verified and unblocks login.</Text>
						</HStack>
						<HStack
							gap="2.5"
							align="flex-start"
							px="3.5"
							py="3"
							borderRadius="12px"
							borderWidth="1px"
							borderColor="border.strong"
							bg="bg.glass"
							fontSize="13px"
							color="fg.subtle"
						>
							<Box color="fg.muted" flex="none" mt="0.5">
								<LuTriangleAlert size={14} />
							</Box>
							<Text>
								<Text as="span" color="aurora.amber" fontWeight="semibold">
									One-way:
								</Text>{" "}
								cannot be undone. To block a verified account, deactivate it.
							</Text>
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
							fontWeight="bold"
							color="#052E22"
							bg="linear-gradient(135deg, #34D399, #22D3EE)"
							boxShadow="0 10px 30px rgba(52,211,153,0.35), inset 0 0 0 1px rgba(255,255,255,0.10)"
							_hover={{ boxShadow: "0 14px 40px rgba(52,211,153,0.50)" }}
							_focusVisible={{ boxShadow: "0 0 0 4px rgba(52,211,153,0.35), 0 14px 40px rgba(52,211,153,0.50)", outline: "none" }}
							loading={submitting}
							onClick={handleSubmit}
						>
							<LuBadgeCheck size={15} /> Verify account{targets.length > 1 ? "s" : ""}
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
