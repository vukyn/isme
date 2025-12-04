"use client";

import { useEffect, useState } from "react";
import { Box, Stack, Text, Heading, Spinner } from "@chakra-ui/react";
import { Button } from "@/components/ui/button";
import { getCurrentUser } from "@/apis/auth";
import { useAuth } from "@/hooks/useAuth";
import type { GetMeResponse } from "@/types";

export const Welcome = () => {
	const { logout } = useAuth();
	const [user, setUser] = useState<GetMeResponse | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		const fetchUser = async () => {
			try {
				setLoading(true);
				setError(null);
				const response = await getCurrentUser();
				// API returns { code, message, data }
				setUser((response as any).data as GetMeResponse);
			} catch (err: any) {
				setError(err?.response?.data?.message || "Failed to load user data");
			} finally {
				setLoading(false);
			}
		};

		fetchUser();
	}, []);

	if (loading) {
		return (
			<Box w="full" h="100vh" display="flex" alignItems="center" justifyContent="center">
				<Stack align="center" gap="4">
					<Spinner size="xl" color="brand.500" />
					<Text color="fg.muted">Loading...</Text>
				</Stack>
			</Box>
		);
	}

	if (error) {
		return (
			<Box w="full" h="100vh" display="flex" alignItems="center" justifyContent="center">
				<Stack align="center" gap="4">
					<Text color="error.solid">{error}</Text>
				</Stack>
			</Box>
		);
	}

	return (
		<Box w="full" h="100vh" bg="bg" position="relative">
			{/* Header with logout button */}
			<Box w="full" p="4" display="flex" justifyContent="flex-end" borderBottom="1px solid" borderColor="border">
				<Button variant="outline" disabled onClick={() => logout()} opacity={0.6} cursor="not-allowed">
					Log out
				</Button>
			</Box>

			{/* Main content */}
			<Box w="full" h="calc(100vh - 73px)" display="flex" alignItems="center" justifyContent="center">
				<Stack align="center" gap="4" textAlign="center">
					<Heading size="2xl" color="fg">
						Welcome, {user?.name || "User"}!
					</Heading>
					<Text color="fg.muted" fontSize="lg">
						You have successfully logged in.
					</Text>
					{user?.email && (
						<Text color="fg.subtle" fontSize="sm">
							{user.email}
						</Text>
					)}
				</Stack>
			</Box>
		</Box>
	);
};
