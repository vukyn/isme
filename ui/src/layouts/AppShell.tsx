import { Box, Stack } from "@chakra-ui/react";
import type { ReactNode } from "react";
import { Topbar, type TopbarTab } from "@/components/ui/topbar";

interface AppShellProps {
	active: TopbarTab;
	user: { name: string; email: string };
	children: ReactNode;
}

export const AppShell = ({ active, user, children }: AppShellProps) => {
	return (
		<Box minH="100vh" w="full">
			<Topbar active={active} user={user} />
			<Stack as="main" gap="6" p="7" maxW="1280px" mx="auto">
				{children}
			</Stack>
		</Box>
	);
};
