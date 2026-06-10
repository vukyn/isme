"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Box, Center, Flex, HStack, NativeSelect, Text } from "@chakra-ui/react";
import { LuPlus, LuX } from "react-icons/lu";
import { listRoles } from "@/apis";
import { AppRoleChip, AppTile } from "@/components/AppRoleChip";
import { toaster } from "@/components/ui/toaster";
import type { AppService, InvitationRoleAssignment, RoleListItem } from "@/types";

/**
 * Repeatable App → Role assignment rows (mock aurora-user-invite #assign-rows).
 * Each row = an App select + an app-scoped Role select + a remove button; one
 * row maps to one models.RoleAssignment { role_id, app_service_id }. RBAC is
 * app-owned, so changing the App repopulates Role from that app's roles.
 *
 * Controlled: `value` holds the assignments, `onChange` reports edits. Roles are
 * lazily fetched per app_code and cached so re-selecting an app is instant.
 */

const LABEL_PROPS = {
	fontSize: "11px",
	fontWeight: "semibold",
	letterSpacing: "0.14em",
	textTransform: "uppercase",
	color: "fg.muted",
} as const;

const SELECT_FIELD_PROPS = {
	h: "11",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "sm",
	color: "fg",
	css: {
		backdropFilter: "blur(12px)",
		WebkitBackdropFilter: "blur(12px)",
		"& option": { background: "#12122E", color: "#F4F5FF" },
	},
	_hover: { borderColor: "rgba(255,255,255,0.28)" },
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" },
} as const;

interface AppScopedRoleAssignmentsProps {
	apps: AppService[];
	value: InvitationRoleAssignment[];
	onChange: (assignments: InvitationRoleAssignment[]) => void;
}

export const AppScopedRoleAssignments = ({ apps, value, onChange }: AppScopedRoleAssignmentsProps) => {
	// Roles cached per app_code (app-owned RBAC); shared by every row. An app
	// referenced by a row but absent from this map is still loading — that drives
	// the per-row "Loading roles…" indicator without extra state.
	const [rolesByApp, setRolesByApp] = useState<Record<string, RoleListItem[]>>({});
	// Already-requested app codes — a ref so dedup never triggers a render. The
	// async setRolesByApp on completion is what re-renders rows out of "loading".
	const requestedApps = useRef<Set<string>>(new Set());

	const appById = useCallback(
		(appServiceId: string) => apps.find((app) => app.id === appServiceId) ?? null,
		[apps]
	);

	// Fetch one app's roles once, caching the result; safe to call repeatedly.
	// No synchronous setState — only the async callbacks touch state.
	const loadRolesFor = useCallback((appCode: string) => {
		if (requestedApps.current.has(appCode)) return;
		requestedApps.current.add(appCode);
		listRoles({ app_code: appCode })
			.then((items) => {
				setRolesByApp((previous) => ({ ...previous, [appCode]: items }));
			})
			.catch(() => {
				setRolesByApp((previous) => ({ ...previous, [appCode]: [] }));
				toaster.create({ title: `Failed to load roles for ${appCode}`, type: "error", meta: { closable: true } });
			});
	}, []);

	// Lazily load roles for every app currently referenced by a row (covers
	// prefilled re-issue rows and the initial seed). setState happens only in
	// loadRolesFor's async callbacks, never synchronously in this effect.
	useEffect(() => {
		for (const assignment of value) {
			const app = appById(assignment.app_service_id);
			if (app) loadRolesFor(app.app_code);
		}
	}, [value, appById, loadRolesFor]);

	// Reconcile rows whose role_id is empty/stale once the app's roles resolve.
	// A native <select> visually shows its first <option> when value="", but
	// state stays empty until an onChange fires — so a seeded/app-switched row
	// looks filled (e.g. "admin") yet never counts as a complete assignment,
	// disabling Send. Snap role_id to the first available role as it loads.
	useEffect(() => {
		let changed = false;
		const reconciled = value.map((assignment) => {
			const app = appById(assignment.app_service_id);
			if (!app) return assignment;
			const roles = rolesByApp[app.app_code];
			if (!roles || roles.length === 0) return assignment;
			if (roles.some((role) => role.id === assignment.role_id)) return assignment;
			changed = true;
			return { ...assignment, role_id: roles[0].id };
		});
		if (changed) onChange(reconciled);
	}, [value, rolesByApp, appById, onChange]);

	const updateRow = (index: number, patch: Partial<InvitationRoleAssignment>) => {
		onChange(value.map((assignment, position) => (position === index ? { ...assignment, ...patch } : assignment)));
	};

	const handleAppChange = (index: number, appServiceId: string) => {
		// Switching app invalidates the chosen role — reset to the app's first
		// role once known, otherwise leave empty until the roles resolve.
		const app = appById(appServiceId);
		const cached = app ? rolesByApp[app.app_code] : undefined;
		updateRow(index, { app_service_id: appServiceId, role_id: cached?.[0]?.id ?? "" });
	};

	const handleRemoveRow = (index: number) => {
		onChange(value.filter((_, position) => position !== index));
	};

	const handleAddRow = () => {
		const firstApp = apps[0];
		const cached = firstApp ? rolesByApp[firstApp.app_code] : undefined;
		onChange([...value, { app_service_id: firstApp?.id ?? "", role_id: cached?.[0]?.id ?? "" }]);
	};

	// "Invitee will get" preview — one chip per fully-specified assignment.
	const grantChips = value
		.map((assignment) => {
			const app = appById(assignment.app_service_id);
			if (!app) return null;
			const role = rolesByApp[app.app_code]?.find((item) => item.id === assignment.role_id);
			if (!role) return null;
			return { key: `${assignment.app_service_id}-${assignment.role_id}`, appCode: app.app_code, roleCode: role.code };
		})
		.filter((chip): chip is { key: string; appCode: string; roleCode: string } => chip !== null);

	return (
		<Box display="grid" gap="4">
			{/* Column labels */}
			<Flex gap="3" px="0.5">
				<Text flex="1" {...LABEL_PROPS}>
					App
				</Text>
				<Text flex="1" {...LABEL_PROPS}>
					Role
				</Text>
				<Box w="11" flex="none" />
			</Flex>

			{/* Repeatable rows */}
			<Box display="grid" gap="2.5">
				{value.map((assignment, index) => {
					const app = appById(assignment.app_service_id);
					const appRoles = app ? rolesByApp[app.app_code] ?? [] : [];
					// Referenced app with no cache entry yet = roles still loading.
					const rolesLoading = app ? rolesByApp[app.app_code] === undefined : false;
					return (
						<Flex key={index} gap="3" align="center">
							{/* App select with leading app tile */}
							<Box flex="1" position="relative">
								{app && (
									<Box position="absolute" left="3" top="50%" transform="translateY(-50%)" zIndex="1" pointerEvents="none">
										<AppTile appCode={app.app_code} />
									</Box>
								)}
								<NativeSelect.Root size="sm" w="full">
									<NativeSelect.Field
										{...SELECT_FIELD_PROPS}
										pl={app ? "11" : "3.5"}
										aria-label={`App for assignment ${index + 1}`}
										value={assignment.app_service_id}
										onChange={(event) => handleAppChange(index, event.target.value)}
									>
										{apps.length === 0 && <option value="">No active apps</option>}
										{apps.map((option) => (
											<option key={option.id} value={option.id}>
												{option.app_name} · {option.app_code}
											</option>
										))}
									</NativeSelect.Field>
									<NativeSelect.Indicator color="fg.muted" />
								</NativeSelect.Root>
							</Box>

							{/* App-scoped Role select */}
							<Box flex="1">
								<NativeSelect.Root size="sm" w="full" disabled={rolesLoading || appRoles.length === 0}>
									<NativeSelect.Field
										{...SELECT_FIELD_PROPS}
										aria-label={`Role for assignment ${index + 1}`}
										value={assignment.role_id}
										onChange={(event) => updateRow(index, { role_id: event.target.value })}
									>
										{rolesLoading && <option value="">Loading roles…</option>}
										{!rolesLoading && appRoles.length === 0 && <option value="">No roles in this app</option>}
										{appRoles.map((role) => (
											<option key={role.id} value={role.id}>
												{role.code}
											</option>
										))}
									</NativeSelect.Field>
									<NativeSelect.Indicator color="fg.muted" />
								</NativeSelect.Root>
							</Box>

							{/* Remove row — disabled when it's the only assignment (≥1 required) */}
							<Center
								as="button"
								w="11"
								h="11"
								flex="none"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border.strong"
								bg="bg.glass"
								color="fg.muted"
								cursor={value.length > 1 ? "pointer" : "not-allowed"}
								opacity={value.length > 1 ? 1 : 0.35}
								css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
								_hover={value.length > 1 ? { bg: "rgba(236,72,153,0.15)", borderColor: "rgba(236,72,153,0.40)", color: "aurora.magenta" } : undefined}
								title={value.length > 1 ? "Remove assignment" : "At least one assignment is required"}
								aria-label={`Remove assignment ${index + 1}`}
								onClick={() => value.length > 1 && handleRemoveRow(index)}
							>
								<LuX size={15} />
							</Center>
						</Flex>
					);
				})}
			</Box>

			{/* Add assignment */}
			<HStack
				as="button"
				gap="2"
				h="10"
				px="3.5"
				w="fit-content"
				borderRadius="glassSm"
				borderWidth="1px"
				borderStyle="dashed"
				borderColor="rgba(139,92,246,0.40)"
				bg="rgba(139,92,246,0.08)"
				color="aurora.violet"
				fontSize="13px"
				fontWeight="semibold"
				cursor={apps.length === 0 ? "not-allowed" : "pointer"}
				aria-disabled={apps.length === 0}
				opacity={apps.length === 0 ? 0.4 : 1}
				_hover={apps.length === 0 ? undefined : { bg: "rgba(139,92,246,0.18)", borderStyle: "solid" }}
				onClick={() => apps.length > 0 && handleAddRow()}
			>
				<LuPlus size={15} /> Add assignment
			</HStack>

			{/* Live "Invitee will get" preview */}
			<Flex
				align="center"
				gap="2"
				wrap="wrap"
				px="3.5"
				py="3"
				borderRadius="glassSm"
				borderWidth="1px"
				borderColor="border"
				bg="rgba(7,7,26,0.35)"
			>
				<Text {...LABEL_PROPS} letterSpacing="0.12em" mr="0.5">
					Invitee will get
				</Text>
				{grantChips.length > 0 ? (
					grantChips.map((chip) => <AppRoleChip key={chip.key} appCode={chip.appCode} role={chip.roleCode} />)
				) : (
					<Text fontSize="12px" color="fg.muted">
						pick an app and role
					</Text>
				)}
			</Flex>
		</Box>
	);
};
