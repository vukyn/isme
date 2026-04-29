"use client";

import { useEffect, useState } from "react";
import { Box, HStack, Stack, Text, Field } from "@chakra-ui/react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { loginSchema, type LoginFormData } from "@/validators";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toaster } from "@/components/ui/toaster";
import { LuArrowRight, LuMail } from "react-icons/lu";
import { PasswordField } from "@/components/ui/password-field";

export const SSOLogin = () => {
	const { login, loading, error } = useAuth();
	const [searchParams] = useSearchParams();
	const navigate = useNavigate();
	const [formData, setFormData] = useState<LoginFormData>({ email: "", password: "" });
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof LoginFormData, string>>>({});

	useEffect(() => {
		const sessionId = searchParams.get("session_id");
		if (!sessionId || sessionId.trim() === "") {
			navigate("/404", { replace: true });
		}
	}, [searchParams, navigate]);

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
		const sessionId = searchParams.get("session_id");
		if (!sessionId || sessionId.trim() === "") {
			toaster.create({ title: "Invalid session", description: "Session ID is required", type: "error" });
			return;
		}
		try {
			const response = await login({ ...formData, session_id: sessionId });
			if (response.data.redirect_url && response.data.authorization_code) {
				const redirectUrl = new URL(response.data.redirect_url);
				redirectUrl.searchParams.set("authorization_code", response.data.authorization_code);
				window.open(redirectUrl.toString(), "_blank");
				toaster.create({ title: "Login successful", description: "Redirecting to application...", type: "success" });
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

	return (
		<Box w="full" minH="100vh" display="flex" alignItems="center" justifyContent="center" px="4">
			<Card w="full" maxW="md" p="8" bg="bg.glass" borderColor="border.strong" borderWidth="1px" borderRadius="2xl" boxShadow="glassSoft" position="relative" zIndex="1" css={{ backdropFilter: "blur(20px)", WebkitBackdropFilter: "blur(20px)" }}>
				<Stack gap="6" as="form" onSubmit={handleSubmit}>
					<Text fontSize="2xl" fontWeight="bold" textAlign="center" color="fg">
						Sign in
					</Text>
					<Stack gap="4">
						<Field.Root invalid={!!formErrors.email}>
							<Field.Label>Email</Field.Label>
							<Input
								type="email"
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
							placeholder="Password"
						/>
					</Stack>
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
						<HStack gap="2.5"><Text>Sign in</Text><LuArrowRight /></HStack>
					</Button>
				</Stack>
			</Card>
		</Box>
	);
};
