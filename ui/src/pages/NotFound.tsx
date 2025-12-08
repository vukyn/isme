"use client";

import { Box, Stack, Text, Heading } from "@chakra-ui/react";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";

export const NotFound = () => {
	const navigate = useNavigate();

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
		>
			<Stack align="center" gap="6" textAlign="center">
				<Heading size="3xl" color="fg">
					404
				</Heading>
				<Text fontSize="xl" color="fg.muted">
					Page Not Found
				</Text>
				<Text color="fg.subtle" fontSize="sm">
					The page you are looking for does not exist.
				</Text>
				<Button variant="solid" onClick={() => navigate("/")}>
					Go Home
				</Button>
			</Stack>
		</Box>
	);
};
