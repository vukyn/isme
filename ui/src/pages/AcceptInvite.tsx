"use client";

import { useEffect, useState } from "react";
import { Box, Button, Center, Field, Flex, HStack, Heading, Spinner, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink, useNavigate, useSearchParams } from "react-router-dom";
import { LuArrowRight, LuInfo, LuLink2Off, LuLock, LuMail, LuShieldCheck, LuUser, LuUserCheck, LuKeyRound } from "react-icons/lu";
import { acceptInvite, getInvitationByToken } from "@/apis";
import { Input } from "@/components/ui/input";
import { toaster } from "@/components/ui/toaster";
import { PasswordField } from "@/components/ui/password-field";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import type { InviteDetailResponse } from "@/types";
import { acceptInviteSchema, type AcceptInviteFormData } from "@/validators";

const INVITE_FEATURES = [
	{ icon: <LuUserCheck />, title: "Pre-verified account", desc: "Your inviter vouched — sign in right away." },
	{ icon: <LuKeyRound />, title: "You pick the password", desc: "Set it once, only you know it." },
	{ icon: <LuShieldCheck />, title: "Single-use link", desc: "This invitation is consumed on account creation." },
];

/**
 * Public /accept-invite?token=… page (mock frames 03 + 03b).
 * Resolves the token on load: valid → locked-email form; any 4xx → one
 * generic error card (never leaks whether the link was expired, revoked
 * or already used).
 */
export const AcceptInvite = () => {
	const navigate = useNavigate();
	const [searchParams] = useSearchParams();
	const token = searchParams.get("token") ?? "";

	const [loading, setLoading] = useState(true);
	const [invalid, setInvalid] = useState(false);
	const [detail, setDetail] = useState<InviteDetailResponse | null>(null);

	const [formData, setFormData] = useState<AcceptInviteFormData>({ name: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof AcceptInviteFormData, string>>>({});
	const [submitting, setSubmitting] = useState(false);

	useEffect(() => {
		if (!token) {
			setInvalid(true);
			setLoading(false);
			return;
		}
		let active = true;
		getInvitationByToken(token)
			.then((response) => {
				if (active) setDetail(response);
			})
			.catch(() => {
				if (active) setInvalid(true);
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, [token]);

	const handleChange = (field: keyof AcceptInviteFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) setFormErrors((prev) => ({ ...prev, [field]: undefined }));
	};

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		const result = acceptInviteSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof AcceptInviteFormData, string>> = {};
			result.error.issues.forEach((issue) => {
				if (issue.path[0]) errors[issue.path[0] as keyof AcceptInviteFormData] = issue.message;
			});
			setFormErrors(errors);
			return;
		}
		setSubmitting(true);
		try {
			await acceptInvite({ token, name: result.data.name, password: result.data.password });
			toaster.create({ title: "Account created", description: "Sign in with your new password.", type: "success" });
			navigate("/login");
		} catch {
			// consumed / expired between load and submit — swap to the error card
			setInvalid(true);
		} finally {
			setSubmitting(false);
		}
	};

	if (loading) {
		return (
			<Center minH="100vh">
				<Spinner size="lg" color="accent" />
			</Center>
		);
	}

	if (invalid || !detail) {
		return (
			<Flex minH="100vh" align="center" justify="center" px="6" py="12">
				<Box
					w="full"
					maxW="460px"
					borderRadius="20px"
					borderWidth="1px"
					borderColor="rgba(236,72,153,0.30)"
					bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
					boxShadow="glassPop"
					px="8"
					py="9"
					textAlign="center"
				>
					<Center
						w="56px"
						h="56px"
						mx="auto"
						mb="4.5"
						borderRadius="18px"
						bg="rgba(236,72,153,0.12)"
						borderWidth="1px"
						borderColor="rgba(236,72,153,0.35)"
						color="aurora.magenta"
						boxShadow="0 0 28px rgba(236,72,153,0.25)"
					>
						<LuLink2Off size={24} />
					</Center>
					<Heading as="h2" fontSize="22px" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
						This invite link is no longer valid
					</Heading>
					<Text fontSize="sm" color="fg.muted" mb="6">
						It may have <b>expired</b>, been <b>revoked</b>, or <b>already been used</b>. Invite links work once and expire after 7 days.
					</Text>
					<Stack gap="2.5">
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
							textAlign="left"
							alignItems="flex-start"
						>
							<Box mt="0.5" flex="none">
								<LuInfo size={14} />
							</Box>
							<Box>
								Ask your administrator to send you a{" "}
								<Text as="span" color="aurora.amber" fontWeight="semibold">
									new invite link
								</Text>
								.
							</Box>
						</HStack>
						<Button
							variant="outline"
							h="12"
							borderRadius="glassSm"
							borderColor="border.strong"
							bg="bg.glass"
							fontSize="md"
							fontWeight="semibold"
							color="fg"
							_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
							onClick={() => navigate("/login")}
						>
							Go to sign in
						</Button>
					</Stack>
				</Box>
			</Flex>
		);
	}

	return (
		<AuthLayout
			brand={
				<BrandPanel
					pill="Invitation · expires soon"
					pillTone="violet"
					titleLead="You're invited to"
					titleGrad="isme."
					sub={`Join the workspace as ${detail.role_name || "a member"}. Set a password to finish creating your account — it only takes a minute.`}
					features={INVITE_FEATURES}
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Create your account
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Invited as{" "}
				<Text as="span" color="fg.subtle" fontWeight="semibold">
					{detail.role_name || "member"}
				</Text>{" "}
				· this link works once and can't be reused.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				{/* Email is bound to the invitation — locked, not editable */}
				<Field.Root>
					<Field.Label>Email</Field.Label>
					<Input
						type="email"
						value={detail.email}
						disabled
						aria-label="Email (locked by invitation)"
						startElement={<LuMail />}
						endElement={<Box color="aurora.amber"><LuLock size={15} /></Box>}
					/>
					<Field.HelperText fontSize="12px" color="fg.muted">
						Locked to this invitation
					</Field.HelperText>
				</Field.Root>

				<Field.Root invalid={!!formErrors.name}>
					<Field.Label>Name</Field.Label>
					<Input
						type="text"
						autoComplete="name"
						placeholder="Your full name"
						value={formData.name}
						onChange={handleChange("name")}
						startElement={<LuUser />}
					/>
					{formErrors.name && <Field.ErrorText>{formErrors.name}</Field.ErrorText>}
				</Field.Root>

				<PasswordField
					label="Password"
					value={formData.password}
					onChange={handleChange("password")}
					error={formErrors.password}
					autoComplete="new-password"
					placeholder="At least 6 characters"
				/>

				<Button
					type="submit"
					h="12"
					mt="1.5"
					loading={submitting}
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					_focusVisible={{ boxShadow: "focusRing" }}
					css={AURORA_CTA_STYLE}
				>
					<HStack gap="2.5">
						<Text>Create account</Text>
						<LuArrowRight />
					</HStack>
				</Button>

				<Text fontSize="12px" color="fg.muted" textAlign="center" mt="2">
					Already have an account?{" "}
					<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg-subtle)", fontWeight: 500 }}>
						Sign in
					</RouterLink>
				</Text>
			</Stack>
		</AuthLayout>
	);
};
