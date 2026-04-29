"use client";

import { useState } from "react";
import { Box, Button, Field, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import {
	LuArrowRight,
	LuCheckCheck,
	LuLock,
	LuMail,
	LuUser,
	LuUsers,
} from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";
import { signupSchema, type SignupFormData } from "@/validators";
import { Input } from "@/components/ui/input";
import { toaster } from "@/components/ui/toaster";
import { PasswordField } from "@/components/ui/password-field";
import { PasswordStrength } from "@/components/ui/password-strength";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";

const SIGNUP_FEATURES = [
	{ icon: <LuUsers />, title: "One-click team invites", desc: "Email or magic link." },
	{ icon: <LuLock />, title: "Encryption at rest", desc: "RSA 2048 keys, never logged." },
	{ icon: <LuCheckCheck />, title: "Audit-ready", desc: "Every session, every token, logged." },
];

export const Signup = () => {
	const { signup, loading, error } = useAuth();
	const [formData, setFormData] = useState<SignupFormData>({ name: "", email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof SignupFormData, string>>>({});

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		const result = signupSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof SignupFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) errors[err.path[0] as keyof SignupFormData] = err.message;
			});
			setFormErrors(errors);
			return;
		}
		try {
			await signup(formData);
			toaster.create({
				title: "Signup successful",
				description: "Your account has been created. Please login.",
				type: "success",
			});
		} catch {
			toaster.create({
				title: "Signup failed",
				description: error?.message || "Please check your information and try again",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof SignupFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) setFormErrors((prev) => ({ ...prev, [field]: undefined }));
	};

	return (
		<AuthLayout
			topRight={
				<Text fontSize="sm" color="fg.muted">
					Already have an account?{" "}
					<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg)", fontWeight: 600, borderBottom: "1px solid var(--chakra-colors-aurora-violet)", paddingBottom: 1 }}>
						Sign in
					</RouterLink>
				</Text>
			}
			brand={
				<BrandPanel
					pill="Free for solo devs"
					pillTone="cyan"
					titleLead="Start in"
					titleGrad="under 2 minutes."
					sub="No credit card. Cancel anytime. Your tokens, your rules."
					features={SIGNUP_FEATURES}
					footerRight="SOC 2 in progress"
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Create your account
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Spin up a workspace and bring your team in.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				<Field.Root invalid={!!formErrors.name}>
					<Field.Label>Full name</Field.Label>
					<Input
						type="text"
						autoComplete="name"
						placeholder="Test Me"
						value={formData.name}
						onChange={handleChange("name")}
						startElement={<LuUser />}
					/>
					{formErrors.name && <Field.ErrorText>{formErrors.name}</Field.ErrorText>}
				</Field.Root>

				<Field.Root invalid={!!formErrors.email}>
					<Field.Label>Work email</Field.Label>
					<Input
						type="email"
						autoComplete="email"
						placeholder="you@company.com"
						value={formData.email}
						onChange={handleChange("email")}
						startElement={<LuMail />}
					/>
					{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
					<Field.HelperText>We'll send a verification link.</Field.HelperText>
				</Field.Root>

				<Box>
					<PasswordField
						label="Password"
						value={formData.password}
						onChange={handleChange("password")}
						error={formErrors.password}
						autoComplete="new-password"
						placeholder="At least 8 characters"
					/>
					<PasswordStrength value={formData.password} />
					<Text fontSize="xs" color="fg.muted" mt="2">
						Use 8+ characters with letters, numbers, and a symbol.
					</Text>
				</Box>

				<Button
					type="submit"
					mt="3"
					h="12"
					loading={loading}
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					_focusVisible={{ boxShadow: "focusRing" }}
					css={{
						background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
						backgroundSize: "200% 200%",
					}}
				>
					<HStack gap="2.5">
						<Text>Create account</Text>
						<LuArrowRight />
					</HStack>
				</Button>
			</Stack>
		</AuthLayout>
	);
};
