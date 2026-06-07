"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, Field, HStack, Input, NativeSelect, Text } from "@chakra-ui/react";
import { LuBadgeCheck, LuCircleCheck, LuCircleX, LuCode, LuKeyRound, LuX } from "react-icons/lu";
import { verifyAppService } from "@/apis";
import { toaster } from "@/components/ui/toaster";
import { APP_SERVICE_CTX_INFO_OPTIONS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { AppServiceCtxInfo } from "@/types";
import { verifyAppServiceSchema } from "@/validators";

interface VerifyAppServiceDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
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
	_invalid: {
		borderColor: "rgba(236,72,153,0.65)",
		boxShadow: "0 0 0 4px rgba(236,72,153,0.14), 0 0 20px rgba(236,72,153,0.18)",
	},
} as const;

type FieldErrors = Partial<Record<"app_code" | "app_secret", string>>;

/**
 * Admin utility over the public POST /verify endpoint — sanity-check a
 * credential pair without leaving the console. The result strip shows the
 * last response; ok:false is intentionally vague server-side (wrong code,
 * ctx mismatch, or bad secret all return the same).
 */
export const VerifyAppServiceDialog = ({ open, onOpenChange }: VerifyAppServiceDialogProps) => {
	const [appCode, setAppCode] = useState("");
	const [ctxInfo, setCtxInfo] = useState<AppServiceCtxInfo>("authen");
	const [appSecret, setAppSecret] = useState("");
	const [errors, setErrors] = useState<FieldErrors>({});
	const [result, setResult] = useState<boolean | null>(null);
	const [submitting, setSubmitting] = useState(false);

	const handleClose = (next: boolean) => {
		if (!next) {
			setAppCode("");
			setCtxInfo("authen");
			setAppSecret("");
			setErrors({});
			setResult(null);
		}
		onOpenChange(next);
	};

	const handleSubmit = async () => {
		const parsed = verifyAppServiceSchema.safeParse({
			app_code: appCode.trim(),
			ctx_info: ctxInfo,
			app_secret: appSecret,
		});
		if (!parsed.success) {
			const next: FieldErrors = {};
			for (const issue of parsed.error.issues) {
				const key = issue.path[0] as keyof FieldErrors;
				if (!next[key]) next[key] = issue.message;
			}
			setErrors(next);
			return;
		}
		setErrors({});
		setSubmitting(true);
		try {
			const response = await verifyAppService(parsed.data);
			setResult(response.ok);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({
				title: err?.response?.data?.message || "Verify request failed",
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
								<LuBadgeCheck size={16} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Verify credentials
							</Dialog.Title>
						</HStack>
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => handleClose(false)}>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						<Field.Root invalid={!!errors.app_code}>
							<Field.Label {...FIELD_LABEL_PROPS}>App code</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuCode size={16} />
								</Box>
								<Input {...INPUT_PROPS} pl="10" placeholder="my-app" value={appCode} onChange={(event) => setAppCode(event.target.value)} />
							</Box>
							{errors.app_code && <Field.ErrorText>{errors.app_code}</Field.ErrorText>}
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>Context</Field.Label>
							<NativeSelect.Root size="sm" w="full">
								<NativeSelect.Field
									{...INPUT_PROPS}
									css={{ "& option": { background: "#12122E", color: "#F4F5FF" } }}
									value={ctxInfo}
									onChange={(event) => setCtxInfo(event.target.value as AppServiceCtxInfo)}
								>
									{APP_SERVICE_CTX_INFO_OPTIONS.map((option) => (
										<option key={option.value} value={option.value}>
											{option.value}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Field.Root>
						<Field.Root invalid={!!errors.app_secret}>
							<Field.Label {...FIELD_LABEL_PROPS}>App secret</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuKeyRound size={16} />
								</Box>
								<Input
									{...INPUT_PROPS}
									pl="10"
									type="password"
									placeholder="App secret to test"
									value={appSecret}
									onChange={(event) => setAppSecret(event.target.value)}
								/>
							</Box>
							{errors.app_secret && <Field.ErrorText>{errors.app_secret}</Field.ErrorText>}
						</Field.Root>
						{result !== null && (
							<HStack
								gap="2"
								px="3.5"
								py="2.5"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor={result ? "rgba(52,211,153,0.30)" : "rgba(236,72,153,0.30)"}
								bg={result ? "rgba(52,211,153,0.07)" : "rgba(236,72,153,0.07)"}
								fontSize="12px"
								color={result ? "aurora.mint" : "aurora.magenta"}
								align="center"
							>
								<Box flex="none">{result ? <LuCircleCheck size={14} /> : <LuCircleX size={14} />}</Box>
								<Text>{result ? "ok: true · credentials valid" : "ok: false · code, context or secret did not match"}</Text>
							</HStack>
						)}
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
							Close
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
							<LuBadgeCheck size={15} /> Verify
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
