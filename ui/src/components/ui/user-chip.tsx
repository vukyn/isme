import { Box, Center, HStack, Menu, Portal, Text } from "@chakra-ui/react";
import { LuChevronDown, LuLogOut } from "react-icons/lu";
import { useAuth } from "@/hooks/useAuth";

interface UserChipProps {
	name: string;
	email: string;
}

const initials = (name: string) =>
	name
		.split(" ")
		.filter(Boolean)
		.map((s) => s[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

export const UserChip = ({ name, email }: UserChipProps) => {
	const { logout } = useAuth();

	return (
		<Menu.Root>
			<Menu.Trigger asChild>
				<HStack
					as="button"
					gap="2.5"
					pl="1"
					pr="2.5"
					py="1"
					borderWidth="1px"
					borderColor="border.strong"
					borderRadius="full"
					bg="bg.glass"
					cursor="pointer"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
					_hover={{ bg: "bg.glassHi" }}
				>
					<Center
						w="8"
						h="8"
						borderRadius="full"
						color="white"
						fontWeight="bold"
						fontSize="xs"
						boxShadow="0 0 16px rgba(139,92,246,0.45)"
						css={{
							background:
								"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
						}}
					>
						{initials(name)}
					</Center>
					<Box textAlign="left" lineHeight="1.15">
						<Text fontSize="sm" fontWeight="semibold" color="fg">
							{name}
						</Text>
						<Text fontSize="xs" color="fg.muted">
							{email}
						</Text>
					</Box>
					<Box color="fg.muted" ml="1">
						<LuChevronDown size={14} />
					</Box>
				</HStack>
			</Menu.Trigger>
			<Portal>
				<Menu.Positioner>
					<Menu.Content
						bg="bg.subtle"
						borderWidth="1px"
						borderColor="border.strong"
						boxShadow="glassSoft"
						css={{ backdropFilter: "blur(20px)", WebkitBackdropFilter: "blur(20px)" }}
					>
						<Menu.Item value="logout" onClick={() => logout()} cursor="pointer">
							<LuLogOut /> Log out
						</Menu.Item>
					</Menu.Content>
				</Menu.Positioner>
			</Portal>
		</Menu.Root>
	);
};
