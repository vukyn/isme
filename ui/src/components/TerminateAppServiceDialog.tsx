"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, HStack, Input, Text } from "@chakra-ui/react";
import { LuCircleX, LuCode, LuX } from "react-icons/lu";
import { updateAppServiceStatus } from "@/apis";
import { toaster } from "@/components/ui/toaster";
import { APP_SERVICE_STATUS } from "@/consts";
import type { AppService } from "@/types";

interface TerminateAppServiceDialogProps {
	open: boolean;
	/** Row the termination targets — supplies id + app_code for type-to-confirm. */
	appService: AppService | null;
	onOpenChange: (open: boolean) => void;
	/** Called after a successful PATCH status=3 so the page can refetch. */
	onTerminated: (appService: AppService) => void;
}

const INPUT_PROPS = {
	h: "11",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "sm",
	color: "fg",
	_placeholder: { color: "fg.muted" },
	_hover: { borderColor: "rgba(255,255,255,0.28)" },
	_focus: { borderColor: "aurora.magenta", boxShadow: "0 0 0 4px rgba(236,72,153,0.14), 0 0 20px rgba(236,72,153,0.18)", outline: "none", bg: "rgba(255,255,255,0.08)" },
} as const;

/**
 * Terminate confirm modal (mock #terminate-modal) — destructive + terminal,
 * so the action is gated behind typing the exact app_code.
 */
export const TerminateAppServiceDialog = ({ open, appService, onOpenChange, onTerminated }: TerminateAppServiceDialogProps) => {
	const [confirmCode, setConfirmCode] = useState("");
	const [submitting, setSubmitting] = useState(false);

	const codeMatches = appService !== null && confirmCode.trim() === appService.app_code;

	const handleClose = (next: boolean) => {
		if (!next) setConfirmCode("");
		onOpenChange(next);
	};

	const handleSubmit = async () => {
		if (!appService || !codeMatches) return;
		setSubmitting(true);
		try {
			await updateAppServiceStatus(appService.id, APP_SERVICE_STATUS.TERMINATED);
			setConfirmCode("");
			onTerminated(appService);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({
				title: err?.response?.data?.message || "Failed to terminate app service",
				type: "error",
				meta: { closable: true },
			});
		} finally {
			setSubmitting(false);
		}
	};

	return (
		<Dialog.Root open={open} onOpenChange={(details) => handleClose(details.open)} placement="center">
			<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
			<Dialog.Positioner>
				<Dialog.Content
					w="480px"
					maxW="92vw"
					borderRadius="20px"
					borderWidth="1px"
					borderColor="rgba(236,72,153,0.40)"
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					color="fg"
					boxShadow="0 20px 60px rgba(236,72,153,0.20), 0 4px 12px rgba(0,0,0,0.35)"
					overflow="hidden"
				>
					<Dialog.Header px="5" py="4" display="flex" alignItems="center" justifyContent="space-between" borderBottomWidth="1px" borderColor="border">
						<HStack gap="3">
							<Center
								w="9"
								h="9"
								borderRadius="11px"
								bg="rgba(236,72,153,0.12)"
								borderWidth="1px"
								borderColor="rgba(236,72,153,0.40)"
								color="aurora.magenta"
							>
								<LuCircleX size={16} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Terminate app service
							</Dialog.Title>
						</HStack>
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => handleClose(false)}>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						<Text fontSize="13px" color="fg.subtle" lineHeight="1.6">
							Terminating{" "}
							<Text as="b" color="fg">
								{appService?.app_name ?? "this app"}
							</Text>{" "}
							permanently revokes its credentials. All verify calls from this app will fail and{" "}
							<Text as="span" color="aurora.magenta">
								this cannot be undone
							</Text>{" "}
							— re-onboarding requires a new registration.
						</Text>
						<Box>
							<Text fontSize="13px" fontWeight="medium" color="fg.subtle" mb="2">
								Type{" "}
								<Text as="b" color="aurora.magenta">
									{appService?.app_code ?? ""}
								</Text>{" "}
								to confirm
							</Text>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#EC4899" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuCode size={16} />
								</Box>
								<Input
									{...INPUT_PROPS}
									pl="10"
									placeholder={appService?.app_code ?? ""}
									value={confirmCode}
									onChange={(event) => setConfirmCode(event.target.value)}
									onKeyDown={(event) => {
										if (event.key === "Enter" && codeMatches && !submitting) handleSubmit();
									}}
								/>
							</Box>
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
							onClick={() => handleClose(false)}
						>
							Cancel
						</Button>
						<Button
							variant="outline"
							h="11"
							px="4.5"
							borderRadius="glassSm"
							fontSize="sm"
							fontWeight="semibold"
							borderColor="rgba(236,72,153,0.40)"
							bg="rgba(236,72,153,0.10)"
							color="aurora.magenta"
							_hover={{ bg: "rgba(236,72,153,0.18)" }}
							disabled={!codeMatches}
							loading={submitting}
							onClick={handleSubmit}
						>
							<LuCircleX size={15} /> Terminate
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
