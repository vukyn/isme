"use client";

import { useState } from "react";
import { Button, Field, HStack, Heading, Stack, Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { LuArrowRight, LuMail, LuShieldCheck, LuClock, LuCheckCheck } from "react-icons/lu";
import { Input } from "@/components/ui/input";
import { toaster } from "@/components/ui/toaster";
import { BrandPanel } from "@/components/ui/brand-panel";
import { AuthLayout } from "@/layouts/AuthLayout";

const RECOVERY_FEATURES = [
	{ icon: <LuShieldCheck />, title: "Encrypted reset tokens", desc: "Single-use, expiring." },
	{ icon: <LuClock />, title: "5-minute window", desc: "Short-lived links." },
	{ icon: <LuCheckCheck />, title: "Audit trail", desc: "Reset attempts logged." },
];

export const ForgotPassword = () => {
	const [email, setEmail] = useState("");

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		toaster.create({
			title: "Coming soon",
			description: "Reset flow not implemented yet.",
			type: "info",
		});
	};

	return (
		<AuthLayout
			topRight={
				<Text fontSize="sm" color="fg.muted">
					Remember it?{" "}
					<RouterLink to="/login" style={{ color: "var(--chakra-colors-fg)", fontWeight: 600, borderBottom: "1px solid var(--chakra-colors-aurora-violet)", paddingBottom: 1 }}>
						Sign in
					</RouterLink>
				</Text>
			}
			brand={
				<BrandPanel
					pill="Account recovery"
					pillTone="cyan"
					titleLead="Forgot your"
					titleGrad="password?"
					sub="Drop your email — we'll send a reset link in seconds."
					features={RECOVERY_FEATURES}
				/>
			}
		>
			<Heading as="h2" fontSize="3xl" fontWeight="bold" letterSpacing="-0.02em" mb="2" color="fg">
				Reset password
			</Heading>
			<Text color="fg.muted" mb="7" fontSize="md">
				Enter the email you used to register.
			</Text>
			<Stack as="form" onSubmit={handleSubmit} gap="4">
				<Field.Root>
					<Field.Label>Email</Field.Label>
					<Input
						type="email"
						autoComplete="email"
						placeholder="you@company.com"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						startElement={<LuMail />}
					/>
				</Field.Root>
				<Button
					type="submit"
					mt="2"
					h="12"
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
						<Text>Send reset link</Text>
						<LuArrowRight />
					</HStack>
				</Button>
			</Stack>
		</AuthLayout>
	);
};
