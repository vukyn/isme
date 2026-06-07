"use client";

import { useState } from "react";
import { Box, Button, Center, Dialog, Field, HStack, Input, NativeSelect, Text } from "@chakra-ui/react";
import { LuAppWindow, LuCode, LuKeyRound, LuLink2, LuPlus, LuX } from "react-icons/lu";
import { registerAppService } from "@/apis";
import { toaster } from "@/components/ui/toaster";
import { APP_SERVICE_CTX_INFO_OPTIONS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { AppServiceCtxInfo } from "@/types";
import { registerAppServiceSchema } from "@/validators";

interface RegisterAppServiceDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Called with the app_code + one-time plaintext secret on success. */
	onRegistered: (appCode: string, appSecret: string) => void;
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

type FieldErrors = Partial<Record<"app_code" | "app_name" | "redirect_url" | "ctx_info", string>>;

export const RegisterAppServiceDialog = ({ open, onOpenChange, onRegistered }: RegisterAppServiceDialogProps) => {
	const [appCode, setAppCode] = useState("");
	const [appName, setAppName] = useState("");
	const [redirectUrl, setRedirectUrl] = useState("");
	const [ctxInfo, setCtxInfo] = useState<AppServiceCtxInfo>("authen");
	const [errors, setErrors] = useState<FieldErrors>({});
	const [submitting, setSubmitting] = useState(false);

	const resetForm = () => {
		setAppCode("");
		setAppName("");
		setRedirectUrl("");
		setCtxInfo("authen");
		setErrors({});
	};

	const handleClose = (next: boolean) => {
		if (!next) resetForm();
		onOpenChange(next);
	};

	const handleSubmit = async () => {
		const parsed = registerAppServiceSchema.safeParse({
			app_code: appCode.trim(),
			app_name: appName.trim(),
			redirect_url: redirectUrl.trim(),
			ctx_info: ctxInfo,
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
			const response = await registerAppService(parsed.data);
			resetForm();
			onRegistered(parsed.data.app_code, response.app_secret);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			const message = err?.response?.data?.message || "";
			// Server-side uniqueness check — surface inline on the app_code field.
			if (message.includes("already exists")) {
				setErrors({ app_code: "app_code already exists — pick another code" });
			} else {
				toaster.create({ title: message || "Failed to register app service", type: "error", meta: { closable: true } });
			}
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
								bg="rgba(139,92,246,0.14)"
								borderWidth="1px"
								borderColor="rgba(139,92,246,0.40)"
								color="aurora.violet"
							>
								<LuAppWindow size={16} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Register app service
							</Dialog.Title>
						</HStack>
						<Button variant="ghost" size="xs" p="1" minW="auto" borderRadius="9px" color="fg.muted" _hover={{ bg: "bg.glass", color: "fg" }} onClick={() => handleClose(false)}>
							<LuX size={16} />
						</Button>
					</Dialog.Header>
					<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
						<Field.Root invalid={!!errors.app_code}>
							<Field.Label {...FIELD_LABEL_PROPS}>
								App code{" "}
								<Text as="span" color="aurora.magenta">
									*
								</Text>{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· unique, immutable
								</Text>
							</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color={errors.app_code ? "aurora.magenta" : "fg.muted"} pointerEvents="none" zIndex="1">
									<LuCode size={16} />
								</Box>
								<Input {...INPUT_PROPS} pl="10" placeholder="my-app" value={appCode} onChange={(event) => setAppCode(event.target.value)} />
							</Box>
							{errors.app_code && <Field.ErrorText>{errors.app_code}</Field.ErrorText>}
						</Field.Root>
						<Field.Root invalid={!!errors.app_name}>
							<Field.Label {...FIELD_LABEL_PROPS}>
								App name{" "}
								<Text as="span" color="aurora.magenta">
									*
								</Text>
							</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuAppWindow size={16} />
								</Box>
								<Input {...INPUT_PROPS} pl="10" placeholder="Human-readable display name" value={appName} onChange={(event) => setAppName(event.target.value)} />
							</Box>
							{errors.app_name && <Field.ErrorText>{errors.app_name}</Field.ErrorText>}
						</Field.Root>
						<Field.Root invalid={!!errors.redirect_url}>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Redirect URL{" "}
								<Text as="span" color="aurora.magenta">
									*
								</Text>
							</Field.Label>
							<Box w="full" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
								<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
									<LuLink2 size={16} />
								</Box>
								<Input {...INPUT_PROPS} pl="10" placeholder="https://app.example.local/auth/callback" value={redirectUrl} onChange={(event) => setRedirectUrl(event.target.value)} />
							</Box>
							{errors.redirect_url ? (
								<Field.ErrorText>{errors.redirect_url}</Field.ErrorText>
							) : (
								<Field.HelperText fontSize="12px" color="fg.muted">
									Where the SSO flow returns the user after sign-in
								</Field.HelperText>
							)}
						</Field.Root>
						<Field.Root>
							<Field.Label {...FIELD_LABEL_PROPS}>
								Context{" "}
								<Text as="span" color="aurora.magenta">
									*
								</Text>{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· ctx_info, fixed set
								</Text>
							</Field.Label>
							<NativeSelect.Root size="sm" w="full">
								<NativeSelect.Field
									{...INPUT_PROPS}
									css={{ "& option": { background: "#12122E", color: "#F4F5FF" } }}
									value={ctxInfo}
									onChange={(event) => setCtxInfo(event.target.value as AppServiceCtxInfo)}
								>
									{APP_SERVICE_CTX_INFO_OPTIONS.map((option) => (
										<option key={option.value} value={option.value}>
											{option.label}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
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
							alignItems="flex-start"
						>
							<Box color="aurora.amber" mt="0.5" flex="none">
								<LuKeyRound size={14} />
							</Box>
							<Box>
								The app secret is generated server-side and shown{" "}
								<Text as="span" color="aurora.amber" fontWeight="semibold">
									exactly once
								</Text>{" "}
								after creation. It cannot be recovered later — only rotated.
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
							onClick={() => handleClose(false)}
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
							<LuPlus size={15} /> Register
						</Button>
					</HStack>
				</Dialog.Content>
			</Dialog.Positioner>
		</Dialog.Root>
	);
};
