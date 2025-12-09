"use client";

import { useState, useEffect } from "react";
import { Box, Stack, Text, Field } from "@chakra-ui/react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { loginSchema, type LoginFormData } from "@/validators";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toaster } from "@/components/ui/toaster";
import { LuMail, LuLock } from "react-icons/lu";

export const SSOLogin = () => {
	const { login, loading, error } = useAuth();
	const [searchParams] = useSearchParams();
	const navigate = useNavigate();
	const [formData, setFormData] = useState<LoginFormData>({
		email: "",
		password: "",
	});
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

		// Validate form
		const result = loginSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof LoginFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) {
					errors[err.path[0] as keyof LoginFormData] = err.message;
				}
			});
			setFormErrors(errors);
			return;
		}

		// Get session_id from query params
		const sessionId = searchParams.get("session_id");
		if (!sessionId || sessionId.trim() === "") {
			toaster.create({
				title: "Invalid session",
				description: "Session ID is required",
				type: "error",
			});
			return;
		}

		try {
			const response = await login({
				...formData,
				session_id: sessionId,
			});

			// Check if this is an SSO login with redirect
			if (response.data.redirect_url && response.data.authorization_code) {
				// Open new tab with redirect URL and authorization code
				const redirectUrl = new URL(response.data.redirect_url);
				redirectUrl.searchParams.set("authorization_code", response.data.authorization_code);
				window.open(redirectUrl.toString(), "_blank");
				toaster.create({
					title: "Login successful",
					description: "Redirecting to application...",
					type: "success",
				});
			} else {
				toaster.create({
					title: "Login successful",
					description: "Welcome back!",
					type: "success",
				});
			}
		} catch (err) {
			// Error is handled by useAuth hook
			toaster.create({
				title: "Login failed",
				description: error?.message || "Please check your credentials",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof LoginFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
		setFormData((prev) => ({ ...prev, [field]: e.target.value }));
		if (formErrors[field]) {
			setFormErrors((prev) => ({ ...prev, [field]: undefined }));
		}
	};

	return (
		<Box
			w="full"
			h="100vh"
			display="flex"
			alignItems="center"
			justifyContent="center"
			bgGradient="to-br"
			gradientFrom="brand.50"
			gradientTo="brand.200"
			_dark={{
				bgGradient: "to-br",
				gradientFrom: "brand.950",
				gradientTo: "brand.900",
			}}
			position="relative"
			overflow="hidden"
		>
			<Card w="full" maxW="md" p="8" bg="bg" rounded="xl" shadow="xl" position="relative" zIndex="1">
				<Stack gap="6" as="form" onSubmit={handleSubmit}>
					<Text fontSize="2xl" fontWeight="bold" textAlign="center" color="fg">
						Login
					</Text>

					<Stack gap="4">
						<Field.Root invalid={!!formErrors.email}>
							<Field.Label>Email</Field.Label>
							<Input
								type="email"
								placeholder="Email"
								value={formData.email}
								onChange={handleChange("email")}
								startElement={<LuMail />}
							/>
							{formErrors.email && <Field.ErrorText>{formErrors.email}</Field.ErrorText>}
						</Field.Root>

						<Field.Root invalid={!!formErrors.password}>
							<Field.Label>Password</Field.Label>
							<Input
								type="password"
								placeholder="Password"
								value={formData.password}
								onChange={handleChange("password")}
								startElement={<LuLock />}
							/>
							{formErrors.password && <Field.ErrorText>{formErrors.password}</Field.ErrorText>}
						</Field.Root>
					</Stack>

					<Button
						type="submit"
						variant="solid"
						w="full"
						loading={loading}
						bgGradient="to-r"
						gradientFrom="brand.500"
						gradientTo="brand.600"
						color="white"
						_hover={{
							bgGradient: "to-r",
							gradientFrom: "brand.600",
							gradientTo: "brand.700",
						}}
					>
						Login
					</Button>
				</Stack>
			</Card>
		</Box>
	);
};
