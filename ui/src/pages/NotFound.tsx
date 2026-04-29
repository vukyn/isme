"use client";

import { Box, Heading, HStack, Stack, Text } from "@chakra-ui/react";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";
import { LuArrowRight } from "react-icons/lu";

export const NotFound = () => {
	const navigate = useNavigate();

	return (
		<Box w="full" minH="100vh" display="flex" alignItems="center" justifyContent="center" px="4">
			<Stack align="center" gap="6" textAlign="center">
				<Heading
					as="h1"
					fontSize="6xl"
					letterSpacing="-0.04em"
					lineHeight="1.15"
					py="2"
					css={{
						backgroundImage: "linear-gradient(135deg, #22D3EE, #8B5CF6 60%, #EC4899)",
						WebkitBackgroundClip: "text",
						backgroundClip: "text",
						color: "transparent",
					}}
				>
					404
				</Heading>
				<Text fontSize="xl" color="fg">Page Not Found</Text>
				<Text color="fg.muted" fontSize="sm">The page you are looking for does not exist.</Text>
				<Button
					h="12"
					px="6"
					color="white"
					borderRadius="glassSm"
					boxShadow="ctaGlow"
					_hover={{ boxShadow: "ctaGlowHi" }}
					css={{
						background: "linear-gradient(135deg, #6366F1 0%, #8B5CF6 50%, #EC4899 100%)",
						backgroundSize: "200% 200%",
					}}
					onClick={() => navigate("/")}
				>
					<HStack gap="2.5"><Text>Go Home</Text><LuArrowRight /></HStack>
				</Button>
			</Stack>
		</Box>
	);
};
