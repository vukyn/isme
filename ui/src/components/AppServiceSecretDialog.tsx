"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, Flex, HStack, Text } from "@chakra-ui/react";
import { LuCheck, LuCopy, LuTriangleAlert } from "react-icons/lu";
import { Checkbox } from "@/components/ui/checkbox";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";

interface AppServiceSecretDialogProps {
	open: boolean;
	/** Header title — "App registered" or "Secret rotated". */
	title: string;
	appCode: string;
	/** One-time plaintext secret — never re-displayable after this dialog closes. */
	secret: string;
	onDone: () => void;
}

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

/**
 * One-time secret display. Deliberately NOT dismissable by backdrop click,
 * escape, or a close button — the only exit is the explicit Done button,
 * gated by the "I have stored the secret" acknowledgement, so the secret
 * cannot be lost by accident.
 */
export const AppServiceSecretDialog = ({ open, title, appCode, secret, onDone }: AppServiceSecretDialogProps) => {
	const [acknowledged, setAcknowledged] = useState(false);
	const [copied, setCopied] = useState(false);

	// Done is the only exit, so resetting here guarantees a fresh
	// acknowledgement the next time the dialog opens with a new secret.
	const handleDone = () => {
		setAcknowledged(false);
		setCopied(false);
		onDone();
	};

	const handleCopy = async () => {
		try {
			await navigator.clipboard.writeText(secret);
			setCopied(true);
			toaster.create({ title: "Secret copied to clipboard", type: "success", meta: { closable: true } });
			setTimeout(() => setCopied(false), 1500);
		} catch {
			toaster.create({ title: "Failed to copy — select the secret and copy manually", type: "error", meta: { closable: true } });
		}
	};

	return (
		<Dialog.Root open={open} onOpenChange={() => {}} placement="center" closeOnInteractOutside={false} closeOnEscape={false}>
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
					{/* Intentionally no CloseTrigger / ✕ — see component doc comment. */}
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
								{title}
							</Dialog.Title>
							<Box
								display="inline-flex"
								alignItems="center"
								gap="7px"
								px="11px"
								py="1"
								borderRadius="full"
								fontSize="12px"
								fontWeight="medium"
								color="aurora.mint"
								borderWidth="1px"
								borderColor="rgba(52,211,153,0.35)"
								bg="rgba(52,211,153,0.10)"
								whiteSpace="nowrap"
							>
								{appCode}
							</Box>
						</HStack>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						<Box>
							<Text fontSize="13px" fontWeight="medium" color="fg.subtle" mb="2">
								App secret{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									·{" "}
									<Text as="span" color="aurora.amber" fontWeight="semibold">
										shown only once
									</Text>
								</Text>
							</Text>
							<Flex align="center" gap="2.5">
								<Flex
									flex="1"
									h="52px"
									align="center"
									px="4"
									borderRadius="glassSm"
									borderWidth="1px"
									borderColor="rgba(34,211,238,0.40)"
									bg="rgba(7,7,26,0.65)"
									color="aurora.cyan"
									fontSize="17px"
									fontWeight="semibold"
									letterSpacing="5px"
									boxShadow="0 0 24px rgba(34,211,238,0.18), 0 0 0 4px rgba(34,211,238,0.06)"
									css={{ userSelect: "all", fontVariantNumeric: "tabular-nums" }}
								>
									{secret}
								</Flex>
								<Center
									as="button"
									w="52px"
									h="52px"
									borderRadius="glassSm"
									cursor="pointer"
									bg="bg.glass"
									borderWidth="1px"
									borderColor={copied ? "rgba(52,211,153,0.40)" : "border.strong"}
									color={copied ? "aurora.mint" : "fg.subtle"}
									css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
									_hover={{ bg: "bg.glassHi", color: copied ? "aurora.mint" : "fg", borderColor: copied ? "rgba(52,211,153,0.40)" : "rgba(255,255,255,0.30)" }}
									title="Copy secret"
									aria-label="Copy secret"
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
								Store this secret now. It is encrypted at rest and <b>cannot be displayed again</b>. If lost, rotate to issue a new one — the old secret stops working immediately.
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
								I have stored the secret in a safe place
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
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
