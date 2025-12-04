import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { Box, Spinner, Stack } from "@chakra-ui/react";
import { getTokens, isTokenExpired } from "@/utils/axios";
import { getCurrentUser } from "@/apis/auth";
import { refreshToken } from "@/apis/auth";
import { saveTokens } from "@/utils/axios";

interface ProtectedRouteProps {
	children: React.ReactNode;
}

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
	const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
	const [isLoading, setIsLoading] = useState(true);

	useEffect(() => {
		const checkAuth = async () => {
			try {
				const tokens = getTokens();

				// No tokens at all
				if (!tokens.access_token) {
					setIsAuthenticated(false);
					setIsLoading(false);
					return;
				}

				// Check if token is expired
				if (isTokenExpired()) {
					// Try to refresh token
					if (tokens.refresh_token) {
						try {
							const response = await refreshToken({
								refresh_token: tokens.refresh_token,
							});
							saveTokens(response.data);
							// Verify with /me after refresh
							await getCurrentUser();
							setIsAuthenticated(true);
						} catch (refreshError) {
							// Refresh failed, redirect to login
							setIsAuthenticated(false);
						}
					} else {
						// No refresh token, redirect to login
						setIsAuthenticated(false);
					}
				} else {
					// Token not expired, verify with /me
					try {
						await getCurrentUser();
						setIsAuthenticated(true);
					} catch (error) {
						// /me failed, try refresh
						if (tokens.refresh_token) {
							try {
								const response = await refreshToken({
									refresh_token: tokens.refresh_token,
								});
								saveTokens(response.data);
								await getCurrentUser();
								setIsAuthenticated(true);
							} catch (refreshError) {
								setIsAuthenticated(false);
							}
						} else {
							setIsAuthenticated(false);
						}
					}
				}
			} catch (error) {
				setIsAuthenticated(false);
			} finally {
				setIsLoading(false);
			}
		};

		checkAuth();
	}, []);

	if (isLoading) {
		return (
			<Box w="full" h="100vh" display="flex" alignItems="center" justifyContent="center">
				<Stack align="center" gap="4">
					<Spinner size="xl" color="brand.500" />
				</Stack>
			</Box>
		);
	}

	if (!isAuthenticated) {
		return <Navigate to="/login" replace />;
	}

	return <>{children}</>;
};
