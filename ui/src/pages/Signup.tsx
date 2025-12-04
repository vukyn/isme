"use client";

import { useState } from "react";
import { Box, Stack, Text, Field } from "@chakra-ui/react";
import { useAuth } from "@/hooks/useAuth";
import { signupSchema, type SignupFormData } from "@/validators";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Link } from "@/components/ui/link";
import { toaster } from "@/components/ui/toaster";
import { LuMail, LuLock, LuUser } from "react-icons/lu";

export const Signup = () => {
	const { signup, loading, error } = useAuth();
	const [formData, setFormData] = useState<SignupFormData>({
		name: "",
		email: "",
		password: "",
	});
	const [formErrors, setFormErrors] = useState<Partial<Record<keyof SignupFormData, string>>>({});
	const [agreeToTerms, setAgreeToTerms] = useState(false);
	const [termsError, setTermsError] = useState("");

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setFormErrors({});
		setTermsError("");

		// Validate terms acceptance
		if (!agreeToTerms) {
			setTermsError("You must agree to the Terms and Conditions");
			return;
		}

		// Validate form
		const result = signupSchema.safeParse(formData);
		if (!result.success) {
			const errors: Partial<Record<keyof SignupFormData, string>> = {};
			result.error.issues.forEach((err) => {
				if (err.path[0]) {
					errors[err.path[0] as keyof SignupFormData] = err.message;
				}
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
		} catch (err) {
			// Error is handled by useAuth hook
			toaster.create({
				title: "Signup failed",
				description: error?.message || "Please check your information and try again",
				type: "error",
			});
		}
	};

	const handleChange = (field: keyof SignupFormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
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
						Register
					</Text>

					<Stack gap="4">
						<Field.Root invalid={!!formErrors.name}>
							<Field.Label>Name</Field.Label>
							<Input
								type="text"
								placeholder="Name"
								value={formData.name}
								onChange={handleChange("name")}
								startElement={<LuUser />}
							/>
							{formErrors.name && <Field.ErrorText>{formErrors.name}</Field.ErrorText>}
						</Field.Root>

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

					<Box>
						<Checkbox
							checked={agreeToTerms}
							onCheckedChange={(e) => {
								setAgreeToTerms(!!e.checked);
								if (termsError) {
									setTermsError("");
								}
							}}
						>
							I agree the{" "}
							<Text as="span" fontWeight="bold">
								Terms and Conditions
							</Text>
						</Checkbox>
						{termsError && (
							<Text color="red.500" fontSize="sm" mt="1">
								{termsError}
							</Text>
						)}
					</Box>

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
						Sign up
					</Button>

					<Text textAlign="center" color="fg.muted" textStyle="sm">
						Already have an account?{" "}
						<Link to="/login" color="brand.500" fontWeight="medium" _hover={{ color: "brand.600" }}>
							Login
						</Link>
					</Text>
				</Stack>
			</Card>
		</Box>
	);
};
