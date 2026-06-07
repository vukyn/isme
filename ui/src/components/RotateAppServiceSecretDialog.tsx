"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, Field, HStack, Input, Text } from "@chakra-ui/react";
import { LuKeyRound, LuRefreshCw, LuShieldCheck, LuTriangleAlert, LuX } from "react-icons/lu";
import { refreshAppServiceSecret } from "@/apis";
import { toaster } from "@/components/ui/toaster";
import type { AppService } from "@/types";
import { rotateAppServiceSecretSchema } from "@/validators";

interface RotateAppServiceSecretDialogProps {
	open: boolean;
	/** Row the rotation targets — supplies app_code + ctx_info for the RefreshRequest. */
	appService: AppService | null;
	onOpenChange: (open: boolean) => void;
	/** Called with the app_code + NEW one-time plaintext secret on success. */
	onRotated: (appCode: string, appSecret: string) => void;
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
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" },
	_invalid: {
		borderColor: "rgba(236,72,153,0.65)",
		boxShadow: "0 0 0 4px rgba(236,72,153,0.14), 0 0 20px rgba(236,72,153,0.18)",
	},
} as const;

export const RotateAppServiceSecretDialog = ({ open, appService, onOpenChange, onRotated }: RotateAppServiceSecretDialogProps) => {
	const [currentSecret, setCurrentSecret] = useState("");
	const [secretError, setSecretError] = useState<string | null>(null);
	const [submitting, setSubmitting] = useState(false);

	const handleClose = (next: boolean) => {
		if (!next) {
			setCurrentSecret("");
			setSecretError(null);
		}
		onOpenChange(next);
	};

	const handleSubmit = async () => {
		if (!appService) return;
		const parsed = rotateAppServiceSecretSchema.safeParse({ app_secret: currentSecret });
		if (!parsed.success) {
			setSecretError(parsed.error.issues[0]?.message ?? "Invalid input");
			return;
		}
		setSecretError(null);
		setSubmitting(true);
		try {
			const response = await refreshAppServiceSecret({
				app_code: appService.app_code,
				app_secret: parsed.data.app_secret,
				ctx_info: appService.ctx_info,
			});
			setCurrentSecret("");
			onRotated(appService.app_code, response.app_secret);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({
				title: err?.response?.data?.message || "Failed to rotate secret",
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
					borderColor="rgba(34,211,238,0.35)"
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					color="fg"
					boxShadow="0 20px 60px rgba(34,211,238,0.20), 0 4px 12px rgba(0,0,0,0.35)"
					overflow="hidden"
				>
					<Dialog.Header px="5" py="4" display="flex" alignItems="center" justifyContent="space-between" borderBottomWidth="1px" borderColor="border">
						<HStack gap="3">
							<Center
								w="9"
								h="9"
								borderRadius="11px"
								bg="rgba(34,211,238,0.12)"
								borderWidth="1px"
								borderColor="rgba(34,211,238,0.35)"
								color="aurora.cyan"
							>
								<LuRefreshCw size={16} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Rotate app secret
							</Dialog.Title>
							<Box
								display="inline-flex"
								alignItems="center"
								px="11px"
								py="1"
								borderRadius="full"
								fontSize="12px"
								fontWeight="medium"
								color="fg.subtle"
								borderWidth="1px"
								borderColor="border.strong"
								bg="bg.glass"
								whiteSpace="nowrap"
							>
								{appService?.app_code ?? ""}
							</Box>
						</HStack>
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => handleClose(false)}>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						<Field.Root invalid={!!secretError}>
							<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
								Current secret{" "}
								<Text as="span" color="aurora.magenta">
									*
								</Text>{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· required by the refresh API
								</Text>
							</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuKeyRound size={16} />
								</Box>
								<Input
									{...INPUT_PROPS}
									pl="10"
									type="password"
									placeholder="Paste the current app secret"
									value={currentSecret}
									onChange={(event) => setCurrentSecret(event.target.value)}
								/>
							</Box>
							{secretError ? (
								<Field.ErrorText>{secretError}</Field.ErrorText>
							) : (
								<Field.HelperText fontSize="12px" color="fg.muted">
									Proves possession — rotation is rejected without a valid current secret
								</Field.HelperText>
							)}
						</Field.Root>
						<HStack
							gap="2"
							px="3.5"
							py="2.5"
							borderRadius="glassSm"
							borderWidth="1px"
							borderColor="rgba(236,72,153,0.30)"
							bg="rgba(236,72,153,0.07)"
							fontSize="12px"
							color="aurora.magenta"
							alignItems="flex-start"
						>
							<Box mt="0.5" flex="none">
								<LuTriangleAlert size={14} />
							</Box>
							<Box>
								The old secret stops working <b>immediately</b>. Every consumer of <b>{appService?.app_code ?? "this app"}</b> must be updated with the new secret or its SSO verify calls will fail.
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
								<LuShieldCheck size={14} />
							</Box>
							<Box>Only an administrator or the app's creator can rotate</Box>
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
							borderColor="rgba(34,211,238,0.40)"
							bg="rgba(34,211,238,0.10)"
							color="aurora.cyan"
							_hover={{ bg: "rgba(34,211,238,0.18)" }}
							loading={submitting}
							onClick={handleSubmit}
						>
							<LuRefreshCw size={15} /> Rotate secret
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
