import { Box, Flex, HStack, IconButton } from "@chakra-ui/react";
import { NavLink } from "react-router-dom";
import { LuBell } from "react-icons/lu";
import { usePermissions } from "@/hooks/usePermissions";
import { BrandMark } from "./brand-mark";
import { UserChip } from "./user-chip";

export type TopbarTab = "overview" | "users" | "roles" | "appServices" | "sessions" | "team" | "settings";

interface TopbarProps {
	active: TopbarTab;
	user: { name: string; email: string };
}

// `perm` gates the nav entry — items the user can't read are hidden from the
// menu (the backend stays the source of truth; this is UX-only).
const NAV: { key: TopbarTab; label: string; to: string; perm?: string }[] = [
	{ key: "overview", label: "Overview", to: "/welcome" },
	{ key: "users", label: "Users", to: "/users", perm: "user:read" },
	{ key: "roles", label: "Roles & Permissions", to: "/roles", perm: "role:read" },
	{ key: "appServices", label: "App Services", to: "/app-services", perm: "app_service:read" },
	{ key: "sessions", label: "Sessions", to: "/sessions", perm: "user_session:read" },
	{ key: "team", label: "Team", to: "/team" },
	{ key: "settings", label: "Settings", to: "/settings", perm: "settings:read" },
];

export const Topbar = ({ active, user }: TopbarProps) => {
	const { can } = usePermissions();
	const visibleNav = NAV.filter((n) => !n.perm || can(n.perm));
	return (
		<Flex
			as="header"
			h="16"
			px="7"
			align="center"
			justify="space-between"
			borderBottomWidth="1px"
			borderColor="border"
			bg="rgba(7,7,26,0.55)"
			css={{ backdropFilter: "blur(20px) saturate(1.2)", WebkitBackdropFilter: "blur(20px) saturate(1.2)" }}
		>
			<HStack gap="6">
				<HStack gap="3">
					<BrandMark size="sm" />
					<Box fontWeight="bold" fontSize="lg" letterSpacing="-0.01em" color="fg">
						isme
					</Box>
				</HStack>
				<HStack as="nav" gap="1" aria-label="Primary">
					{visibleNav.map((n) => {
						const isActive = n.key === active;
						return (
							<NavLink
								key={n.key}
								to={n.to}
								aria-current={isActive ? "page" : undefined}
								style={{ textDecoration: "none" }}
							>
								<Box
									px="3.5"
									py="2"
									fontSize="sm"
									fontWeight="medium"
									borderRadius="md"
									color={isActive ? "fg" : "fg.muted"}
									bg={isActive ? "bg.glass" : "transparent"}
									borderWidth="1px"
									borderColor={isActive ? "border.strong" : "transparent"}
									boxShadow={isActive ? "0 0 20px rgba(99,102,241,0.20)" : "none"}
									_hover={{ color: "fg" }}
								>
									{n.label}
								</Box>
							</NavLink>
						);
					})}
				</HStack>
			</HStack>
			<HStack gap="2.5">
				<IconButton
					aria-label="Notifications"
					variant="ghost"
					size="md"
					color="fg.subtle"
					bg="bg.glass"
					borderWidth="1px"
					borderColor="border.strong"
					borderRadius="md"
					position="relative"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
					_hover={{ color: "fg", borderColor: "rgba(255,255,255,0.28)" }}
				>
					<LuBell />
					<Box
						position="absolute"
						top="1.5"
						right="1.5"
						w="2"
						h="2"
						borderRadius="full"
						bg="aurora.magenta"
						css={{ boxShadow: "0 0 8px #EC4899" }}
					/>
				</IconButton>
				<UserChip name={user.name} email={user.email} />
			</HStack>
		</Flex>
	);
};
