import { Box, Flex, Grid } from "@chakra-ui/react";
import type { ReactNode } from "react";

interface AuthLayoutProps {
	topRight?: ReactNode;
	brand: ReactNode;
	children: ReactNode;
}

export const AuthLayout = ({ topRight, brand, children }: AuthLayoutProps) => {
	return (
		<Flex w="full" minH="100vh" align="center" justify="center" px={{ base: "4", md: "6" }} py="12">
			<Box
				w="full"
				maxW="1280px"
				bg="bg.glass"
				borderWidth="1px"
				borderColor="border"
				borderRadius="3xl"
				boxShadow="glassSoft"
				overflow="hidden"
				position="relative"
			>
				<Grid
					templateColumns={{ base: "1fr", md: "1.05fr 1fr" }}
					minH={{ base: "auto", md: "720px" }}
				>
					{brand}
					<Flex
						direction="column"
						justify="center"
						p={{ base: "8", md: "14" }}
						bg="rgba(7,7,26,0.55)"
						css={{ backdropFilter: "blur(8px)", WebkitBackdropFilter: "blur(8px)" }}
					>
						{topRight && (
							<Flex justify="flex-end" mb="9">
								{topRight}
							</Flex>
						)}
						<Box w="full" maxW="440px" mx="auto">
							{children}
						</Box>
					</Flex>
				</Grid>
			</Box>
		</Flex>
	);
};
