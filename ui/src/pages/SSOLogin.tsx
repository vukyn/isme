"use client";

import { useCallback, useEffect, useState } from "react";
import { Box, Center, Field, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { LuArrowRight, LuMail, LuLock, LuCircleAlert } from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";
import { loginSchema, type LoginFormData } from "@/validators";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { GlassCard } from "@/components/ui/glass-card";
import { BrandMark } from "@/components/ui/brand-mark";
import { PasswordField } from "@/components/ui/password-field";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { ssoCheck, ssoConsent } from "@/apis/auth";
import { getTokens } from "@/utils/axios";
import type { SSOScope } from "@/types";
import { SSOConsent } from "./SSOConsent";

// Requesting-app name shown in the SSO context header when no isme session
// exists yet (password form). The app initiating the handshake is only resolved
// server-side via session_id, which /sso/check returns — but the password form
// renders before that resolves, so it keeps a generic fallback.
const REQUESTING_APP_NAME = "this application";

/** State driving which screen the SSO page shows. */
type CheckState =
	| { phase: "checking" }
	| { phase: "password" }
	| {
			phase: "consent";
			appName: string;
			redirectURL: string;
			userName: string;
			userEmail: string;
			scopes: SSOScope[];
			nonce: string;
	  };

/**
 * Hand the authorization_code back to the requesting app. Shared by the
 * consent-allow path and the password-submit path so the two can't drift:
 * post to the opener tab (cross-origin via the redirect_url's origin) and close,
 * else fall back to a same-tab redirect carrying the code in the query string.
 */
const deliverAuthorizationCode = (redirectURL: string, authorizationCode: string) => {
	const targetOrigin = new URL(redirectURL).origin;
	if (window.opener && !window.opener.closed) {
		window.opener.postMessage({ type: "SSO_AUTH_SUCCESS", authorization_code: authorizationCode }, targetOrigin);
		window.close();
	} else {
		const url = new URL(redirectURL);
		url.searchParams.set("authorization_code", authorizationCode);
		window.location.href = url.toString();
	}
};

export const SSOLogin = () => {
	const { login, loading, error } = useAuth();
	const [searchParams] = useSearchParams();
	const navigate = useNavigate();
	const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof LoginFormData, string>>>({});
	const [check, setCheck] = useState<CheckState>({ phase: "checking" });
	const [consentLoading, setConsentLoading] = useState(false);
	// Requesting-app name shown on the password form. Resolved server-side from
	// session_id (returned by /sso/check even when no isme session exists), so it
	// shows "continue to <app>" instead of the generic fallback.
	const [appName, setAppName] = useState(REQUESTING_APP_NAME);

	const sessionId = searchParams.get("session_id") ?? "";

	// On mount: guard the session_id, then run the read-only silent-authorize
	// probe. A live isme session → consent screen; otherwise → password form.
	useEffect(() => {
		if (!sessionId || sessionId.trim() === "") {
			navigate("/404", { replace: true });
			return;
		}

		let cancelled = false;
		(async () => {
			const tokens = getTokens();
			try {
				// Always probe — tokens are optional. With no isme session the
				// response is valid=false but still carries the app name, so the
				// password form header reads "continue to <app>".
				const res = await ssoCheck({
					session_id: sessionId,
					access_token: tokens.access_token || undefined,
					refresh_token: tokens.refresh_token || undefined,
				});
				if (cancelled) return;
				if (res.data.app?.name) setAppName(res.data.app.name);
				if (res.data.valid) {
					setCheck({
						phase: "consent",
						appName: res.data.app.name,
						redirectURL: res.data.app.redirect_url,
						userName: res.data.user.name,
						userEmail: res.data.user.email,
						scopes: res.data.scopes,
						nonce: res.data.nonce,
					});
				} else {
					setCheck({ phase: "password" });
				}
			} catch {
				// probe failed (e.g. invalid session) → fall back to the password form
				if (!cancelled) setCheck({ phase: "password" });
			}
		})();

		return () => {
			cancelled = true;
		};
	}, [sessionId, navigate]);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		const result = loginSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof LoginFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) errors[err.path[0] as keyof LoginFormData] = err.message;
			});
			setFormErrors(errors);
			return;
		}
		if (!sessionId || sessionId.trim() === "") {
			toaster.create({ title: "Invalid session", description: "Session ID is required", type: "error" });
			return;
		}
		try {
			const response = await login({ ...formData, session_id: sessionId });
			if (response.data.redirect_url && response.data.authorization_code) {
				toaster.create({ title: "Login successful", description: "Redirecting to application...", type: "success" });
				deliverAuthorizationCode(response.data.redirect_url, response.data.authorization_code);
			} else {
				toaster.create({ title: "Login successful", description: "Welcome back!", type: "success" });
			}
		} catch {
			toaster.create({
				title: "Login failed",
				description: error?.message || "Please check your credentials",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof LoginFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) setFormErrors((prev) => ({ ...prev, [field]: undefined }));
	};

	// Consent "Allow": authorize via the nonce, then hand the code back exactly
	// like the password path does.
	const handleConsentAllow = useCallback(async () => {
		if (check.phase !== "consent") return;
		setConsentLoading(true);
		try {
			const tokens = getTokens();
			const res = await ssoConsent({
				session_id: sessionId,
				access_token: tokens.access_token || undefined,
				refresh_token: tokens.refresh_token || undefined,
				nonce: check.nonce,
			});
			toaster.create({ title: "Authorized", description: "Redirecting to application...", type: "success" });
			deliverAuthorizationCode(res.data.redirect_url, res.data.authorization_code);
		} catch (err) {
			const message = (err as { response?: { data?: { message?: string } } })?.response?.data?.message;
			toaster.create({ title: "Authorization failed", description: message || "Please try again.", type: "error" });
			setConsentLoading(false);
		}
	}, [check, sessionId]);

	// Consent "Cancel/Deny": no code issued — close the popup or go back.
	const handleConsentDeny = useCallback(() => {
		if (window.opener && !window.opener.closed) {
			window.close();
		} else {
			navigate("/login", { replace: true });
		}
	}, [navigate]);

	// "Not you? Switch account": drop to the password form for this session.
	const handleSwitchAccount = useCallback(() => {
		setCheck({ phase: "password" });
	}, []);

	// While the silent-authorize probe is in flight, render nothing structural
	// (the aurora background still shows) to avoid flashing the password form.
	if (check.phase === "checking") {
		return <Box minH="100dvh" display="grid" placeItems="center" px="5" />;
	}

	if (check.phase === "consent") {
		return (
			<SSOConsent
				appName={check.appName}
				userName={check.userName}
				userEmail={check.userEmail}
				scopes={check.scopes}
				loading={consentLoading}
				onAllow={handleConsentAllow}
				onSwitchAccount={handleSwitchAccount}
				onDeny={handleConsentDeny}
			/>
		);
	}

	return (
		<Box minH="100dvh" display="grid" placeItems="center" px="5" pt="10" pb="18">
			<Stack w="full" maxW="440px" gap="18px">
				{/* isme brand — the identity provider doing the auth */}
				<HStack justify="center" gap="2.5">
					<BrandMark />
					<Text fontWeight="bold" fontSize="lg" letterSpacing="-0.01em" color="fg">
						isme
					</Text>
				</HStack>

				<GlassCard tone="strong" p="7" pb="6" borderRadius="20px" position="relative" overflow="hidden">
					{/* ambient orbs */}
					<Box
						position="absolute"
						w="220px"
						h="220px"
						right="-70px"
						top="-90px"
						borderRadius="full"
						pointerEvents="none"
						css={{ filter: "blur(46px)", background: "radial-gradient(circle, rgba(99,102,241,0.50), transparent 70%)" }}
					/>
					<Box
						position="absolute"
						w="200px"
						h="200px"
						left="-70px"
						bottom="-90px"
						borderRadius="full"
						pointerEvents="none"
						css={{ filter: "blur(46px)", background: "radial-gradient(circle, rgba(236,72,153,0.42), transparent 70%)" }}
					/>

					<Box position="relative" zIndex={1}>
						{/* ===== SSO context header: which app you're signing in to ===== */}
						<Stack align="center" textAlign="center" gap="3.5" mb="22px">
							<HStack gap="3.5" aria-hidden="true">
								{/* requesting app (generic placeholder tile) */}
								<Center
									w="52px"
									h="52px"
									borderRadius="15px"
									color="accent"
									bg="rgba(139,92,246,0.18)"
									borderWidth="1px"
									borderColor="rgba(139,92,246,0.45)"
									css={{ boxShadow: "0 0 22px rgba(139,92,246,0.25)" }}
								>
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
										<path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z" />
									</svg>
								</Center>
								<Center color="fg.muted">
									<svg width="26" height="16" viewBox="0 0 26 16" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round">
										<path d="M2 8h22" />
										<path d="m7 3-5 5 5 5" />
										<path d="m19 3 5 5-5 5" />
									</svg>
								</Center>
								{/* isme (IdP) — lg = 52px to match the medioa tile */}
								<BrandMark size="lg" />
							</HStack>
							<Heading as="h1" fontSize="22px" fontWeight="bold" letterSpacing="-0.02em" lineHeight="1.25" color="fg">
								Sign in to continue to {appName}
							</Heading>
							<Text fontSize="sm" color="fg.muted" maxW="340px">
								<Text as="b" color="fg.subtle" fontWeight="semibold">
									{appName}
								</Text>{" "}
								uses isme single sign-on. Log in with your isme account to authorize access.
							</Text>
						</Stack>

						<Stack as="form" onSubmit={handleSubmit} gap="0">
							{/* form-level error banner */}
							{error && (
								<HStack
									role="alert"
									aria-live="polite"
									gap="2.5"
									px="3.5"
									py="2.5"
									mb="4"
									borderRadius="glassSm"
									bg="rgba(248,113,113,0.10)"
									borderWidth="1px"
									borderColor="rgba(248,113,113,0.34)"
									color="danger"
									fontSize="sm"
								>
									<Box flex="0 0 auto" display="grid" placeItems="center">
										<LuCircleAlert size={16} />
									</Box>
									<Text>{error.message || "Incorrect email or password. Please try again."}</Text>
								</HStack>
							)}

							<Box mb="3.5">
								<Field.Root invalid={!!formErrors.email}>
									<Field.Label>Email</Field.Label>
									<Input
										type="email"
										autoComplete="email"
										inputMode="email"
										placeholder="you@company.com"
										value={formData.email}
										onChange={handleChange("email")}
										disabled={loading}
										startElement={<LuMail />}
									/>
									{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
								</Field.Root>
							</Box>

							<Box mb="3.5">
								<PasswordField
									label="Password"
									value={formData.password}
									onChange={handleChange("password")}
									error={formErrors.password}
									autoComplete="current-password"
									placeholder="Enter your password"
								/>
							</Box>

							<Button
								type="submit"
								mt="5"
								h="50px"
								loading={loading}
								loadingText="Signing in…"
								color="white"
								borderRadius="glassSm"
								boxShadow="ctaGlow"
								_hover={{ boxShadow: "ctaGlowHi" }}
								_focusVisible={{ boxShadow: "focusRing" }}
								css={AURORA_CTA_STYLE}
							>
								<HStack gap="2.5">
									<Text>Continue</Text>
									<LuArrowRight />
								</HStack>
							</Button>
						</Stack>
					</Box>
				</GlassCard>

				{/* trust footer: secured-by pill + privacy line */}
				<HStack
					alignSelf="center"
					gap="2"
					px="3"
					py="1.5"
					borderRadius="full"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border.strong"
					color="fg.subtle"
					fontSize="xs"
					fontWeight="medium"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
				>
					<Box color="success" display="grid" placeItems="center">
						<LuLock size={13} />
					</Box>
					Secured by isme SSO
				</HStack>
				<Text textAlign="center" fontSize="xs" color="fg.muted">
					Your password is never shared with {appName}.
				</Text>
			</Stack>
		</Box>
	);
};
