"use client";

import { useState } from "react";
import { Button, Field, Flex, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { LuArrowRight, LuMail, LuShieldCheck, LuClock, LuCheckCheck } from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";
import { loginSchema, type LoginFormData } from "@/validators";
import { Input } from "@/components/ui/input";
import { toaster } from "@/components/ui/toaster";
import { PasswordField } from "@/components/ui/password-field";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";

const LOGIN_FEATURES = [
	{ icon: <LuShieldCheck />, title: "JWT + refresh rotation", desc: "HS256, rotation built-in." },
	{ icon: <LuClock />, title: "Per-device sessions", desc: "Revoke, list, audit instantly." },
	{ icon: <LuCheckCheck />, title: "Compliant by default", desc: "WCAG AA, audit-ready logs." },
];

export const Login = () => {
	const { login, loading, error } = useAuth();
	const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof LoginFormData, string>>>({});

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
		try {
			await login(formData);
			toaster.create({ title: "Login successful", description: "Welcome back!", type: "success" });
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

	return (
		<AuthLayout
			topRight={
				<Text fontSize="sm" color="fg.muted">
					New here?{" "}
					<RouterLink to="/signup" style={{ color: "var(--chakra-colors-fg)", fontWeight: 600, borderBottom: "1px solid var(--chakra-colors-aurora-violet)", paddingBottom: 1 }}>
						Create account
					</RouterLink>
				</Text>
			}
			brand={
				<BrandPanel
					pill="Welcome back"
					pillTone="violet"
					titleLead="Build, ship, repeat —"
					titleGrad="without the chaos."
					sub="One workspace for auth, sessions, teams. Aurora-grade security, quiet by default."
					features={LOGIN_FEATURES}
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Sign in
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Use your work email and password. SSO available.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				<Field.Root invalid={!!formErrors.email}>
					<Field.Label>Email</Field.Label>
					<Input
						type="email"
						autoComplete="email"
						placeholder="you@company.com"
						value={formData.email}
						onChange={handleChange("email")}
						startElement={<LuMail />}
					/>
					{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
				</Field.Root>

				<PasswordField
					label="Password"
					value={formData.password}
					onChange={handleChange("password")}
					error={formErrors.password}
					autoComplete="current-password"
					placeholder="Enter password"
				/>

				<Flex justify="flex-end" align="center" mt="1" mb="2">
					<RouterLink
						to="/forgot-password"
						style={{ color: "var(--chakra-colors-fg-subtle)", fontSize: 14, fontWeight: 500, textDecoration: "none" }}
					>
						Forgot password?
					</RouterLink>
				</Flex>

				<Button
					type="submit"
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
						<Text>Sign in</Text>
						<LuArrowRight />
					</HStack>
				</Button>
			</Stack>
		</AuthLayout>
	);
};
