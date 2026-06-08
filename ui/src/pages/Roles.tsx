"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { Box, Button, Center, Dialog, Field, Flex, Grid, Heading, HStack, Input, Spinner, Table, Text } from "@chakra-ui/react";
import {
	LuAppWindow,
	LuChevronLeft,
	LuChevronRight,
	LuEye,
	LuGlobe,
	LuInfo,
	LuKeyRound,
	LuLock,
	LuMonitorSmartphone,
	LuPencil,
	LuPlus,
	LuSearch,
	LuShieldCheck,
	LuTrash2,
	LuUser,
	LuUsers,
	LuWrench,
	LuX,
} from "react-icons/lu";
import {
	addRoleMembers,
	deleteRole,
	getRole,
	listAppServices,
	listPermissions,
	listRoleMembers,
	listRoles,
	listUsers,
	removeRoleMember,
	setRolePermissions,
	updateRole,
} from "@/apis";
import { AppTile } from "@/components/AppRoleChip";
import { CreateRoleDialog } from "@/components/CreateRoleDialog";
import { Checkbox } from "@/components/ui/checkbox";
import { APP_SERVICE_STATUS } from "@/consts";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { useUser } from "@/hooks/useUser";
import { usePermissions } from "@/hooks/usePermissions";
import { AppShell } from "@/layouts/AppShell";
import type { AppService, PermissionItem, RoleDetailResponse, RoleListItem, RoleMemberItem, UserListItem } from "@/types";
import { formatDateOnly } from "@/utils";

const MEMBERS_PAGE_SIZE = 8;

type RoleTab = "permissions" | "members" | "apps";

interface RoleAccent {
	color: string;
	bg: string;
	glow: string;
	icon: ReactNode;
}

/** Visual identity per role — system roles keep the mock's fixed colors, custom roles are amber. */
const roleAccent = (code: string, size = 16): RoleAccent => {
	if (code === "admin") {
		return { color: "aurora.violet", bg: "rgba(139,92,246,0.15)", glow: "rgba(139,92,246,0.18)", icon: <LuShieldCheck size={size} /> };
	}
	if (code === "member") {
		return { color: "aurora.cyan", bg: "rgba(34,211,238,0.12)", glow: "rgba(34,211,238,0.18)", icon: <LuUser size={size} /> };
	}
	if (code === "viewer") {
		return { color: "aurora.mint", bg: "rgba(52,211,153,0.12)", glow: "rgba(52,211,153,0.18)", icon: <LuEye size={size} /> };
	}
	return { color: "aurora.amber", bg: "rgba(245,158,11,0.12)", glow: "rgba(245,158,11,0.18)", icon: <LuWrench size={size} /> };
};

interface MatrixResource {
	resource: string;
	label: string;
	accent: { color: string; bg: string };
	icon: ReactNode;
	/** Which CRUD actions this resource has in the app's catalog. */
	crud: Record<"read" | "create" | "update" | "delete", boolean>;
	/** Non-CRUD actions (verify, revoke, rotate_secret, …). */
	special: string[];
}

const CRUD_ACTIONS = ["read", "create", "update", "delete"] as const;
const CRUD_SET = new Set<string>(CRUD_ACTIONS);

const RESOURCE_ACCENTS = [
	{ color: "aurora.violet", bg: "rgba(139,92,246,0.15)" },
	{ color: "aurora.cyan", bg: "rgba(34,211,238,0.12)" },
	{ color: "aurora.mint", bg: "rgba(52,211,153,0.12)" },
	{ color: "aurora.amber", bg: "rgba(245,158,11,0.12)" },
];

const RESOURCE_ICONS: Record<string, ReactNode> = {
	user: <LuUsers size={14} />,
	user_session: <LuMonitorSmartphone size={14} />,
	session: <LuMonitorSmartphone size={14} />,
	app_service: <LuAppWindow size={14} />,
	role: <LuKeyRound size={14} />,
};

/** Turn a snake_case resource into a Title Case label. */
const resourceLabel = (resource: string) =>
	resource
		.split("_")
		.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
		.join(" ");

/**
 * Build the role × permission matrix from the SELECTED APP's permission catalog
 * (app-owned RBAC: each app has its own resources). Rows are grouped by
 * resource; columns are CRUD; anything non-CRUD becomes a "special" chip.
 */
const buildMatrix = (catalog: PermissionItem[]): MatrixResource[] => {
	const byResource = new Map<string, Set<string>>();
	for (const permission of catalog) {
		if (!byResource.has(permission.resource)) byResource.set(permission.resource, new Set());
		byResource.get(permission.resource)!.add(permission.action);
	}
	return [...byResource.entries()]
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([resource, actions], index) => ({
			resource,
			label: resourceLabel(resource),
			accent: RESOURCE_ACCENTS[index % RESOURCE_ACCENTS.length],
			icon: RESOURCE_ICONS[resource] ?? <LuKeyRound size={14} />,
			crud: {
				read: actions.has("read"),
				create: actions.has("create"),
				update: actions.has("update"),
				delete: actions.has("delete"),
			},
			special: [...actions].filter((action) => !CRUD_SET.has(action)).sort(),
		}));
};

const AVATAR_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
	"linear-gradient(135deg, #F59E0B, #EC4899)",
	"linear-gradient(135deg, #34D399, #22D3EE)",
];

/** Uppercase micro-label (mock .rail-head / .flabel / thead th). */
const LABEL_PROPS = {
	fontSize: "11px",
	fontWeight: "semibold",
	letterSpacing: "0.14em",
	textTransform: "uppercase",
	color: "fg.muted",
} as const;

/** Glass pill (mock .pill / .count-pill). */
const PILL_BASE = {
	display: "inline-flex",
	alignItems: "center",
	gap: "7px",
	px: "11px",
	py: "3px",
	borderRadius: "full",
	fontSize: "12px",
	fontWeight: "medium",
	borderWidth: "1px",
	borderColor: "border.strong",
	bg: "bg.glass",
	whiteSpace: "nowrap",
} as const;

/** Glass panel (mock .panel). */
const PANEL_PROPS = {
	bg: "bg.glass",
	borderWidth: "1px",
	borderColor: "border",
	borderRadius: "20px",
	boxShadow: "glassSoft",
	overflow: "hidden",
	css: { backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" },
} as const;

const INPUT_PROPS = {
	h: "11",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "sm",
	color: "fg",
	_placeholder: { color: "fg.muted" },
	_hover: { borderColor: "rgba(255,255,255,0.28)" },
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" },
} as const;

/** Aurora gradient checkbox control (mock .check input:checked). */
const CHECKBOX_CSS = {
	"& .chakra-checkbox__control": {
		borderRadius: "5px",
		borderColor: "rgba(255,255,255,0.18)",
		background: "rgba(255,255,255,0.06)",
	},
	"& .chakra-checkbox__control[data-state=checked]": {
		background: "linear-gradient(135deg, #6366F1, #8B5CF6)",
		borderColor: "#8B5CF6",
		boxShadow: "0 0 12px rgba(139,92,246,0.45)",
	},
} as const;

const GHOST_SM_BUTTON_PROPS = {
	variant: "outline",
	h: "34px",
	px: "3",
	borderRadius: "10px",
	borderColor: "border.strong",
	bg: "bg.glass",
	fontSize: "13px",
	fontWeight: "semibold",
	color: "fg",
	css: { backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" },
	_hover: { bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" },
} as const;

const DANGER_SM_BUTTON_PROPS = {
	variant: "outline",
	h: "34px",
	px: "3",
	borderRadius: "10px",
	borderColor: "rgba(236,72,153,0.35)",
	bg: "rgba(236,72,153,0.08)",
	fontSize: "13px",
	fontWeight: "semibold",
	color: "aurora.magenta",
	_hover: { bg: "rgba(236,72,153,0.16)" },
} as const;

const initials = (name: string) =>
	name
		.split(" ")
		.filter(Boolean)
		.map((part) => part[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

const avatarGradient = (id: string) => {
	let hash = 0;
	for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) % 997;
	return AVATAR_GRADIENTS[hash % AVATAR_GRADIENTS.length];
};

const errorMessage = (error: unknown, fallback: string): string => {
	const err = error as { response?: { data?: { message?: string } }; message?: string };
	return err?.response?.data?.message || err?.message || fallback;
};

const RoleIconBox = ({ code, size = "9", radius = "11px", iconSize = 16, glow = false }: { code: string; size?: string; radius?: string; iconSize?: number; glow?: boolean }) => {
	const accent = roleAccent(code, iconSize);
	return (
		<Center
			w={size}
			h={size}
			flex="none"
			borderRadius={radius}
			bg={accent.bg}
			borderWidth="1px"
			borderColor="border.strong"
			color={accent.color}
			boxShadow={glow ? `0 0 18px ${accent.glow}` : "none"}
		>
			{accent.icon}
		</Center>
	);
};

/** Member avatar — 30px conic/linear gradient circle. */
const MemberAvatar = ({ id, name }: { id: string; name: string }) => (
	<Center
		w="30px"
		h="30px"
		flex="none"
		borderRadius="full"
		fontSize="11px"
		fontWeight="bold"
		color="white"
		css={{ background: avatarGradient(id), boxShadow: "0 0 12px rgba(99,102,241,0.25)" }}
	>
		{initials(name)}
	</Center>
);

export const Roles = () => {
	const { user: currentUser } = useUser();
	const { can } = usePermissions();

	const canCreate = can("role:create");
	const canUpdate = can("role:update");
	const canDelete = can("role:delete");
	const canAssign = can("role:assign");
	const canReadUsers = can("user:read");

	// App switcher (app-owned RBAC) — selecting an app reloads its roles + catalog.
	const [apps, setApps] = useState<AppService[]>([]);
	const [selectedAppId, setSelectedAppId] = useState<string | null>(null);
	const [appsLoading, setAppsLoading] = useState(true);

	const [roles, setRoles] = useState<RoleListItem[]>([]);
	const [catalog, setCatalog] = useState<PermissionItem[]>([]);
	const [loading, setLoading] = useState(true);

	const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);
	const [detail, setDetail] = useState<RoleDetailResponse | null>(null);
	const [detailLoading, setDetailLoading] = useState(false);

	const [tab, setTab] = useState<RoleTab>("permissions");

	// Permission matrix working state — Save replaces the role's full permission set.
	const [originalPermissionIds, setOriginalPermissionIds] = useState<Set<number>>(new Set());
	const [draftPermissionIds, setDraftPermissionIds] = useState<Set<number>>(new Set());
	const [savingPermissions, setSavingPermissions] = useState(false);

	const [members, setMembers] = useState<RoleMemberItem[]>([]);
	const [membersTotal, setMembersTotal] = useState(0);
	const [membersPage, setMembersPage] = useState(1);
	const [membersLoading, setMembersLoading] = useState(false);

	const [memberSearch, setMemberSearch] = useState("");
	const [memberResults, setMemberResults] = useState<UserListItem[]>([]);
	const [memberSearching, setMemberSearching] = useState(false);

	const [createOpen, setCreateOpen] = useState(false);
	const [renameOpen, setRenameOpen] = useState(false);
	const [renameName, setRenameName] = useState("");
	const [renameDescription, setRenameDescription] = useState("");
	const [renameSaving, setRenameSaving] = useState(false);
	const [deleteOpen, setDeleteOpen] = useState(false);
	const [deleting, setDeleting] = useState(false);

	const selectedApp = useMemo(() => apps.find((app) => app.id === selectedAppId) ?? null, [apps, selectedAppId]);
	// isme is itself an app whose roles/permissions are system-managed (read-only).
	const isIsmeApp = selectedApp?.app_code === "isme";

	const refreshRoles = useCallback(
		async (selectId?: string) => {
			if (!selectedApp) return [];
			const items = await listRoles({ app_code: selectedApp.app_code });
			setRoles(items);
			if (selectId) {
				setSelectedRoleId(selectId);
			} else {
				setSelectedRoleId((previous) => (previous && items.some((role) => role.id === previous) ? previous : (items[0]?.id ?? null)));
			}
			return items;
		},
		[selectedApp]
	);

	// Load the app catalog for the switcher (active apps only).
	useEffect(() => {
		let active = true;
		listAppServices({ page: 1, page_size: 100, status: APP_SERVICE_STATUS.ACTIVE })
			.then((response) => {
				if (!active) return;
				setApps(response.items);
				// default to isme (the system app) when present, else the first app
				const isme = response.items.find((app) => app.app_code === "isme");
				setSelectedAppId(isme?.id ?? response.items[0]?.id ?? null);
			})
			.catch((error: unknown) => {
				if (active) toaster.create({ title: errorMessage(error, "Failed to load apps"), type: "error", meta: { closable: true } });
			})
			.finally(() => {
				if (active) setAppsLoading(false);
			});
		return () => {
			active = false;
		};
	}, []);

	// Per-app load: roles list + permission catalog scoped to the selected app.
	useEffect(() => {
		if (!selectedApp) return;
		let active = true;
		setLoading(true);
		Promise.all([listRoles({ app_code: selectedApp.app_code }), listPermissions({ app_code: selectedApp.app_code })])
			.then(([roleItems, permissionItems]) => {
				if (!active) return;
				setRoles(roleItems);
				setCatalog(permissionItems);
				setSelectedRoleId(roleItems[0]?.id ?? null);
			})
			.catch((error: unknown) => {
				if (active) {
					toaster.create({ title: errorMessage(error, "Failed to load roles"), type: "error", meta: { closable: true } });
				}
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, [selectedApp]);

	// Role selection → load detail, reset tab-local state.
	useEffect(() => {
		if (!selectedRoleId) {
			setDetail(null);
			return;
		}
		let active = true;
		setDetailLoading(true);
		setMembers([]);
		setMembersTotal(0);
		setMembersPage(1);
		setMemberSearch("");
		setMemberResults([]);
		getRole(selectedRoleId)
			.then((roleDetail) => {
				if (!active) return;
				setDetail(roleDetail);
				const permissionIds = new Set(roleDetail.permissions.map((permission) => permission.id));
				setOriginalPermissionIds(permissionIds);
				setDraftPermissionIds(new Set(permissionIds));
			})
			.catch((error: unknown) => {
				if (active) {
					toaster.create({ title: errorMessage(error, "Failed to load role"), type: "error", meta: { closable: true } });
				}
			})
			.finally(() => {
				if (active) setDetailLoading(false);
			});
		return () => {
			active = false;
		};
	}, [selectedRoleId]);

	const loadMembers = useCallback(
		async (roleId: string, page: number) => {
			setMembersLoading(true);
			try {
				const response = await listRoleMembers(roleId, { page, size: MEMBERS_PAGE_SIZE });
				setMembers(response.items);
				setMembersTotal(response.total);
			} catch (error: unknown) {
				toaster.create({ title: errorMessage(error, "Failed to load members"), type: "error", meta: { closable: true } });
			} finally {
				setMembersLoading(false);
			}
		},
		[]
	);

	// Members + App scope tabs need the assignment list.
	useEffect(() => {
		if (!selectedRoleId || (tab !== "members" && tab !== "apps")) return;
		loadMembers(selectedRoleId, membersPage);
	}, [selectedRoleId, tab, membersPage, loadMembers]);

	// Typeahead over users (add-member). Debounced; requires user:read.
	useEffect(() => {
		const query = memberSearch.trim();
		if (!query || !canAssign || !canReadUsers) {
			setMemberResults([]);
			return;
		}
		const timer = setTimeout(() => {
			setMemberSearching(true);
			listUsers({ page: 1, size: 8, query })
				.then((response) => {
					const memberIds = new Set(members.map((member) => member.user_id));
					setMemberResults(response.items.filter((item) => !memberIds.has(item.id)));
				})
				.catch(() => setMemberResults([]))
				.finally(() => setMemberSearching(false));
		}, 300);
		return () => clearTimeout(timer);
	}, [memberSearch, canAssign, canReadUsers, members]);

	const permissionByCode = useMemo(() => {
		const map = new Map<string, PermissionItem>();
		for (const permission of catalog) {
			map.set(`${permission.resource}:${permission.action}`, permission);
		}
		return map;
	}, [catalog]);

	const selectedRole = useMemo(() => roles.find((role) => role.id === selectedRoleId) ?? null, [roles, selectedRoleId]);
	// Read-only when the role is a system role OR the whole app is system-managed
	// (isme): definitions are seeded and locked.
	const isReadOnly = !!detail?.is_system || isIsmeApp;
	const matrixEditable = !!detail && !isReadOnly && canUpdate;
	// Dynamic matrix from the selected app's permission catalog.
	const matrixResources = useMemo(() => buildMatrix(catalog), [catalog]);

	const dirtyCount = useMemo(() => {
		let count = 0;
		for (const id of draftPermissionIds) if (!originalPermissionIds.has(id)) count += 1;
		for (const id of originalPermissionIds) if (!draftPermissionIds.has(id)) count += 1;
		return count;
	}, [draftPermissionIds, originalPermissionIds]);

	const membersTotalPages = Math.max(1, Math.ceil(membersTotal / MEMBERS_PAGE_SIZE));

	// Scope distribution of the loaded assignments (App scope tab is read-only).
	const scopeDistribution = useMemo(() => {
		let globalCount = 0;
		const perApp = new Map<string, number>();
		for (const member of members) {
			if (!member.app_service_id) globalCount += 1;
			else perApp.set(member.app_service_id, (perApp.get(member.app_service_id) ?? 0) + 1);
		}
		return { globalCount, perApp };
	}, [members]);

	const togglePermission = (permissionId: number) => {
		setDraftPermissionIds((previous) => {
			const next = new Set(previous);
			if (next.has(permissionId)) next.delete(permissionId);
			else next.add(permissionId);
			return next;
		});
	};

	const handleDiscardPermissions = () => {
		setDraftPermissionIds(new Set(originalPermissionIds));
	};

	const handleSavePermissions = async () => {
		if (!selectedRoleId) return;
		setSavingPermissions(true);
		try {
			await setRolePermissions(selectedRoleId, [...draftPermissionIds].sort((a, b) => a - b));
			setOriginalPermissionIds(new Set(draftPermissionIds));
			toaster.create({ title: "Permissions saved", type: "success", meta: { closable: true } });
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to save permissions"), type: "error", meta: { closable: true } });
		} finally {
			setSavingPermissions(false);
		}
	};

	const handleAddMember = async (user: UserListItem) => {
		if (!selectedRoleId) return;
		try {
			// TODO(backend): no app_service list endpoint exists yet (only
			// register/verify/refresh), so the scope picker is hidden and new
			// assignments default to global (app_service_id NULL).
			await addRoleMembers(selectedRoleId, { user_ids: [user.id] });
			setMemberSearch("");
			setMemberResults([]);
			setRoles((previous) =>
				previous.map((role) => (role.id === selectedRoleId ? { ...role, members_count: role.members_count + 1 } : role))
			);
			await loadMembers(selectedRoleId, membersPage);
			toaster.create({ title: `${user.name || user.email} added to ${selectedRole?.code}`, type: "success", meta: { closable: true } });
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to add member"), type: "error", meta: { closable: true } });
		}
	};

	const handleRemoveMember = async (member: RoleMemberItem) => {
		if (!selectedRoleId) return;
		try {
			await removeRoleMember(selectedRoleId, member.user_id, member.app_service_id);
			setRoles((previous) =>
				previous.map((role) =>
					role.id === selectedRoleId ? { ...role, members_count: Math.max(0, role.members_count - 1) } : role
				)
			);
			await loadMembers(selectedRoleId, membersPage);
			toaster.create({ title: `${member.name || member.email} removed`, type: "success", meta: { closable: true } });
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to remove member"), type: "error", meta: { closable: true } });
		}
	};

	const handleOpenRename = () => {
		setRenameName(detail?.name ?? "");
		setRenameDescription(detail?.description ?? "");
		setRenameOpen(true);
	};

	const handleRename = async () => {
		if (!selectedRoleId || !renameName.trim()) return;
		setRenameSaving(true);
		try {
			await updateRole(selectedRoleId, { name: renameName.trim(), description: renameDescription.trim() });
			setDetail((previous) => (previous ? { ...previous, name: renameName.trim(), description: renameDescription.trim() } : previous));
			setRoles((previous) =>
				previous.map((role) =>
					role.id === selectedRoleId ? { ...role, name: renameName.trim(), description: renameDescription.trim() } : role
				)
			);
			setRenameOpen(false);
			toaster.create({ title: "Role updated", type: "success", meta: { closable: true } });
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to update role"), type: "error", meta: { closable: true } });
		} finally {
			setRenameSaving(false);
		}
	};

	const handleDelete = async () => {
		if (!selectedRoleId) return;
		setDeleting(true);
		try {
			await deleteRole(selectedRoleId);
			setDeleteOpen(false);
			toaster.create({ title: `Role ${selectedRole?.code} deleted`, type: "success", meta: { closable: true } });
			await refreshRoles();
		} catch (error: unknown) {
			// Backend rejects deletes while members still hold the role
			// ("role has members; reassign first") — surface it as-is.
			toaster.create({ title: errorMessage(error, "Failed to delete role"), type: "error", meta: { closable: true } });
		} finally {
			setDeleting(false);
		}
	};

	const handleCreated = async (roleId: string) => {
		try {
			await refreshRoles(roleId);
			setTab("permissions");
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to reload roles"), type: "error", meta: { closable: true } });
		}
	};

	const renderPermissionCell = (resource: MatrixResource, action: (typeof CRUD_ACTIONS)[number]) => {
		if (!resource.crud[action]) {
			return (
				<Text as="span" fontSize="13px" color="fg.muted" opacity={0.6}>
					—
				</Text>
			);
		}
		const permission = permissionByCode.get(`${resource.resource}:${action}`);
		if (!permission) {
			return (
				<Text as="span" fontSize="13px" color="fg.muted" opacity={0.6}>
					—
				</Text>
			);
		}
		return (
			<Checkbox
				size="sm"
				colorPalette="purple"
				css={CHECKBOX_CSS}
				checked={draftPermissionIds.has(permission.id)}
				disabled={!matrixEditable}
				onCheckedChange={() => togglePermission(permission.id)}
			/>
		);
	};

	const tabButton = (key: RoleTab, label: string, icon: ReactNode, pill?: string) => (
		<HStack
			as="button"
			key={key}
			px="3.5"
			py="13px"
			gap="2"
			fontSize="sm"
			fontWeight="medium"
			cursor="pointer"
			color={tab === key ? "fg" : "fg.muted"}
			borderBottomWidth="2px"
			borderBottomColor={tab === key ? "aurora.violet" : "transparent"}
			mb="-1px"
			_hover={{ color: "fg" }}
			onClick={() => setTab(key)}
		>
			{icon}
			{label}
			{pill && (
				<Box {...PILL_BASE} px="2" py="1px" fontSize="11px" fontWeight="semibold" color="fg.subtle" css={{ fontVariantNumeric: "tabular-nums" }}>
					{pill}
				</Box>
			)}
		</HStack>
	);

	return (
		<AppShell active="roles" user={{ name: currentUser?.name || "User", email: currentUser?.email || "" }}>
			{/* Page head */}
			<Flex align="flex-end" justify="space-between" gap="4" wrap="wrap">
				<Box>
					<Heading fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1" color="fg">
						Roles & Permissions
					</Heading>
					<Text fontSize="sm" color="fg.muted" mt="1.5">
						App-owned RBAC · roles &amp; permissions belong to an{" "}
						<Text as="span" color="fg.subtle">
							app_service
						</Text>{" "}
						· managing{" "}
						<Text as="span" color="aurora.cyan" fontWeight="semibold">
							{selectedApp?.app_code ?? "—"}
						</Text>
					</Text>
				</Box>
				{canCreate && (
					<Button
						h="11"
						px="4.5"
						borderRadius="glassSm"
						fontSize="sm"
						fontWeight="semibold"
						color="white"
						css={AURORA_CTA_STYLE}
						boxShadow="ctaGlow"
						_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
						_focusVisible={{ boxShadow: "focusRing" }}
						disabled={isIsmeApp}
						title={isIsmeApp ? "isme roles are system-managed — read-only" : undefined}
						onClick={() => setCreateOpen(true)}
					>
						{isIsmeApp ? <LuLock size={16} /> : <LuPlus size={16} />} New role
					</Button>
				)}
			</Flex>

			{/* App switcher — drives the whole panel below (roles + catalog + read-only). */}
			<Box {...PANEL_PROPS} p="3.5">
				<Flex align="center" gap="2.5" wrap="wrap">
					<Text {...LABEL_PROPS} mr="1">
						App
					</Text>
					{apps.map((app) => {
						const on = app.id === selectedAppId;
						return (
							<HStack
								key={app.id}
								as="button"
								gap="2.5"
								h="11"
								px="3.5"
								borderRadius="glassSm"
								cursor="pointer"
								borderWidth="1px"
								borderColor={on ? "rgba(139,92,246,0.45)" : "border.strong"}
								bg={on ? "linear-gradient(135deg, rgba(99,102,241,0.30), rgba(139,92,246,0.25))" : "bg.glass"}
								boxShadow={on ? "0 0 16px rgba(139,92,246,0.25) inset" : "none"}
								color={on ? "fg" : "fg.muted"}
								_hover={{ color: "fg" }}
								onClick={() => setSelectedAppId(app.id)}
							>
								<AppTile appCode={app.app_code} size="22px" />
								<Box lineHeight="1.2" textAlign="left">
									<Text fontSize="13px" fontWeight="semibold">
										{app.app_code}
									</Text>
									<Text fontSize="11px" color="fg.muted" truncate maxW="120px">
										{app.app_name}
									</Text>
								</Box>
								{app.app_code === "isme" && (
									<Box color="aurora.amber" flex="none" title="System-managed">
										<LuLock size={12} />
									</Box>
								)}
							</HStack>
						);
					})}
				</Flex>
				{/* isme system banner — only for the isme app */}
				{isIsmeApp && (
					<HStack
						gap="2.5"
						mt="3"
						px="3.5"
						py="3"
						borderRadius="glassSm"
						borderWidth="1px"
						borderColor="rgba(245,158,11,0.30)"
						bg="rgba(245,158,11,0.07)"
						fontSize="12.5px"
						color="aurora.amber"
						align="flex-start"
					>
						<Box mt="0.5" flex="none">
							<LuLock size={14} />
						</Box>
						<Text color="fg.subtle">
							<Text as="span" color="fg" fontWeight="semibold">
								System-managed app.
							</Text>{" "}
							isme's own roles (admin / member / viewer) and its permissions are seeded and{" "}
							<Text as="span" color="aurora.amber" fontWeight="semibold">
								read-only
							</Text>{" "}
							here (is_system = true). You can still assign these roles to users — only the role &amp; permission definitions are
							locked. Platform admin = the admin role on this app.
						</Text>
					</HStack>
				)}
			</Box>

			{loading || appsLoading ? (
				<Center py="16">
					<Spinner size="lg" color="accent" />
				</Center>
			) : (
				<Grid templateColumns={{ base: "1fr", md: "300px 1fr" }} gap="5" alignItems="start">
					{/* Roles list (left rail) */}
					<Box {...PANEL_PROPS}>
						<Box px="4.5" py="13px" {...LABEL_PROPS} borderBottomWidth="1px" borderColor="border">
							Roles
						</Box>
						{roles.map((role) => {
							const selected = role.id === selectedRoleId;
							return (
								<HStack
									key={role.id}
									gap="13px"
									px="4.5"
									py="3.5"
									cursor="pointer"
									borderBottomWidth="1px"
									borderColor="border"
									borderLeftWidth="2px"
									borderLeftColor={selected ? "aurora.violet" : "transparent"}
									bg={selected ? "linear-gradient(90deg, rgba(139,92,246,0.18), rgba(139,92,246,0.04))" : "transparent"}
									boxShadow={selected ? "0 0 20px rgba(139,92,246,0.15) inset" : "none"}
									transition="border-color .15s ease-out, background .15s ease-out"
									_hover={selected ? undefined : { bg: "rgba(255,255,255,0.04)" }}
									onClick={() => setSelectedRoleId(role.id)}
								>
									<RoleIconBox code={role.is_system ? role.code : "custom"} />
									<Box minW="0" lineHeight="1.3" flex="1">
										<HStack gap="7px" fontSize="sm" fontWeight="semibold" color="fg">
											<Text truncate>{role.code}</Text>
											{role.is_system && (
												<Box color="fg.muted" flex="none" title="System role — immutable">
													<LuLock size={12} />
												</Box>
											)}
										</HStack>
										<Text fontSize="12px" color="fg.muted" truncate>
											{role.description || (role.is_system ? "system role" : "custom")}
										</Text>
									</Box>
									<Box {...PILL_BASE} px="10px" fontWeight="semibold" color="fg.subtle" css={{ fontVariantNumeric: "tabular-nums" }}>
										{role.members_count.toLocaleString()}
									</Box>
								</HStack>
							);
						})}
						<HStack px="4.5" py="13px" gap="2" fontSize="12px" color="fg.muted">
							<LuInfo size={13} /> system roles seeded · immutable
						</HStack>
					</Box>

					{/* Detail panel */}
					<Box {...PANEL_PROPS}>
						{detailLoading || !detail ? (
							<Center py="16">
								{detailLoading ? (
									<Spinner size="md" color="accent" />
								) : (
									<Text fontSize="sm" color="fg.muted">
										Select a role
									</Text>
								)}
							</Center>
						) : (
							<>
								{/* Role header */}
								<HStack gap="3.5" px="5" py="4.5" borderBottomWidth="1px" borderColor="border" wrap="wrap">
									<RoleIconBox code={detail.is_system ? detail.code : "custom"} size="11" radius="13px" iconSize={20} glow />
									<Box lineHeight="1.3" flex="1" minW="0">
										<HStack gap="2.5" fontSize="17px" fontWeight="bold" letterSpacing="-0.01em" color="fg">
											<Text>{detail.code}</Text>
											{detail.is_system ? (
												<Box {...PILL_BASE} color="fg.subtle">
													system
												</Box>
											) : (
												<Box {...PILL_BASE} color="aurora.amber" borderColor="rgba(245,158,11,0.35)" bg="rgba(245,158,11,0.10)">
													custom
												</Box>
											)}
										</HStack>
										<Text fontSize="12px" color="fg.muted" mt="3px" truncate css={{ fontVariantNumeric: "tabular-nums" }}>
											role_id ·{" "}
											<Text as="span" color="fg.subtle">
												{detail.id}
											</Text>
											{detail.name && ` · ${detail.name}`}
										</Text>
									</Box>
									{/* Rename + Delete — hidden for system/isme roles, shown for editable custom roles. */}
									{!isReadOnly && (
										<>
											<Button {...GHOST_SM_BUTTON_PROPS} disabled={!canUpdate} onClick={handleOpenRename}>
												<LuPencil size={13} /> Rename
											</Button>
											<Button {...DANGER_SM_BUTTON_PROPS} disabled={!canDelete} onClick={() => setDeleteOpen(true)}>
												<LuTrash2 size={13} /> Delete
											</Button>
										</>
									)}
								</HStack>

								{/* Tabs */}
								<HStack gap="1" px="3.5" borderBottomWidth="1px" borderColor="border">
									{tabButton("permissions", "Permissions", <LuKeyRound size={14} />)}
									{tabButton(
										"members",
										"Members",
										<LuUsers size={14} />,
										(selectedRole?.members_count ?? 0).toLocaleString()
									)}
									{tabButton("apps", "App scope", <LuAppWindow size={14} />)}
								</HStack>

								{/* Tab 1 · permission matrix */}
								{tab === "permissions" && (
									<>
										{isReadOnly && (
											<HStack gap="2" px="4.5" py="2.5" borderBottomWidth="1px" borderColor="rgba(245,158,11,0.25)" bg="rgba(245,158,11,0.06)" fontSize="12px" color="aurora.amber">
												<LuLock size={13} />
												{isIsmeApp
													? "System-managed app — isme permissions are seeded and read-only"
													: "System role — permissions are seeded and read-only"}
											</HStack>
										)}
										{/* read-only matrix gets a faint amber wash (mock .matrix-readonly) */}
										<Box bg={isReadOnly ? "rgba(245,158,11,0.03)" : "transparent"}>
											<Table.Root size="sm" css={{ "& td, & th": { borderColor: "rgba(255,255,255,0.10)" } }}>
												<Table.Header>
													<Table.Row bg="transparent">
														<Table.ColumnHeader px="4.5" py="3" bg="transparent" {...LABEL_PROPS}>
															Resource
														</Table.ColumnHeader>
														{CRUD_ACTIONS.map((action) => (
															<Table.ColumnHeader key={action} w="84px" px="2" py="3" bg="transparent" textAlign="center" {...LABEL_PROPS}>
																{action}
															</Table.ColumnHeader>
														))}
														<Table.ColumnHeader w="220px" px="4.5" py="3" bg="transparent" {...LABEL_PROPS}>
															Special
														</Table.ColumnHeader>
													</Table.Row>
												</Table.Header>
												<Table.Body>
													{matrixResources.length === 0 && (
														<Table.Row bg="transparent">
															<Table.Cell colSpan={6} px="4.5" py="8" textAlign="center" fontSize="sm" color="fg.muted">
																This app has no permissions in its catalog.
															</Table.Cell>
														</Table.Row>
													)}
													{matrixResources.map((resource) => (
														<Table.Row key={resource.resource} bg="transparent" transition="background .15s ease-out" _hover={{ bg: "rgba(255,255,255,0.03)" }}>
															<Table.Cell px="4.5" py="13px">
																<HStack gap="3">
																	<Center
																		w="8"
																		h="8"
																		flex="none"
																		borderRadius="10px"
																		bg={resource.accent.bg}
																		borderWidth="1px"
																		borderColor="border.strong"
																		color={resource.accent.color}
																	>
																		{resource.icon}
																	</Center>
																	<Box lineHeight="1.25">
																		<Text fontSize="sm" fontWeight="medium" color="fg">
																			{resource.label}
																		</Text>
																		<Text fontSize="11px" color="fg.muted">
																			{resource.resource}
																		</Text>
																	</Box>
																</HStack>
															</Table.Cell>
															{CRUD_ACTIONS.map((action) => (
																<Table.Cell key={action} px="2" py="13px" textAlign="center">
																	<Center>{renderPermissionCell(resource, action)}</Center>
																</Table.Cell>
															))}
															<Table.Cell px="4.5" py="13px">
																<HStack gap="2" wrap="wrap">
																	{resource.special.map((action) => {
																		const permission = permissionByCode.get(`${resource.resource}:${action}`);
																		if (!permission) return null;
																		const checked = draftPermissionIds.has(permission.id);
																		return (
																			<HStack
																				key={action}
																				as="label"
																				{...PILL_BASE}
																				px="3"
																				py="5px"
																				gap="2"
																				color={isReadOnly && checked ? "aurora.amber" : "fg.subtle"}
																				borderColor={isReadOnly && checked ? "rgba(245,158,11,0.35)" : "border.strong"}
																				bg={isReadOnly && checked ? "rgba(245,158,11,0.10)" : "bg.glass"}
																				cursor={matrixEditable ? "pointer" : "default"}
																				transition="border-color .15s ease-out, background .15s ease-out"
																				_hover={matrixEditable ? { borderColor: "rgba(255,255,255,0.30)", bg: "bg.glassHi" } : undefined}
																			>
																				<Checkbox
																					size="sm"
																					colorPalette={isReadOnly ? "yellow" : "purple"}
																					css={CHECKBOX_CSS}
																					checked={checked}
																					disabled={!matrixEditable}
																					onCheckedChange={() => togglePermission(permission.id)}
																				/>
																				{action}
																			</HStack>
																		);
																	})}
																</HStack>
															</Table.Cell>
														</Table.Row>
													))}
												</Table.Body>
											</Table.Root>
										</Box>

										{/* Read-only lock notice (mock .lock-notice) — replaces the unsaved bar. */}
										{isReadOnly && (
											<HStack gap="2" px="4.5" py="3" borderTopWidth="1px" borderColor="rgba(245,158,11,0.25)" bg="rgba(245,158,11,0.06)" fontSize="12px" color="aurora.amber" align="flex-start">
												<Box mt="0.5" flex="none">
													<LuLock size={13} />
												</Box>
												<Text>
													<Text as="span" fontWeight="semibold">
														System-managed
													</Text>{" "}
													— {isIsmeApp ? "isme" : "system role"} permissions are seeded and immutable. To grant or revoke access, assign a
													different role to the user from the Users screen.
												</Text>
											</HStack>
										)}

										{/* Unsaved-changes bar (editable apps only) */}
										{!isReadOnly && dirtyCount > 0 && (
											<HStack
												gap="3"
												px="4.5"
												py="3"
												borderTopWidth="1px"
												borderColor="rgba(139,92,246,0.30)"
												bg="linear-gradient(90deg, rgba(139,92,246,0.16), rgba(99,102,241,0.08))"
												fontSize="13px"
												wrap="wrap"
											>
												<Text color="aurora.violet" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
													{dirtyCount} unsaved change{dirtyCount > 1 ? "s" : ""}
												</Text>
												<Box h="18px" w="1px" bg="border.strong" />
												<Button {...GHOST_SM_BUTTON_PROPS} onClick={handleDiscardPermissions}>
													Discard
												</Button>
												<Button
													h="34px"
													px="3"
													borderRadius="10px"
													fontSize="13px"
													fontWeight="semibold"
													color="white"
													css={AURORA_CTA_STYLE}
													boxShadow="ctaGlow"
													_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
													loading={savingPermissions}
													onClick={handleSavePermissions}
												>
													Save changes
												</Button>
												<HStack ml="auto" gap="2" fontSize="12px" color="fg.muted">
													<LuInfo size={13} /> takes effect on next token refresh
												</HStack>
											</HStack>
										)}
									</>
								)}

								{/* Tab 2 · members */}
								{tab === "members" && (
									<>
										{canAssign && canReadUsers && (
											<Box p="3.5" borderBottomWidth="1px" borderColor="border" position="relative">
												<Box position="relative" css={{ "&:focus-within .search-icon": { color: "#22D3EE" } }}>
													<Box className="search-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none">
														<LuSearch size={16} />
													</Box>
													<Input
														{...INPUT_PROPS}
														pl="10"
														placeholder="Add member — search by name or email…"
														value={memberSearch}
														onChange={(event) => setMemberSearch(event.target.value)}
													/>
												</Box>
												{(memberResults.length > 0 || memberSearching) && memberSearch.trim() && (
													<Box
														position="absolute"
														left="3.5"
														right="3.5"
														top="100%"
														mt="-1"
														zIndex="20"
														borderRadius="14px"
														borderWidth="1px"
														borderColor="border.strong"
														bg="linear-gradient(180deg, rgba(18,18,46,0.96), rgba(11,11,35,0.97))"
														boxShadow="glassPop"
														overflow="hidden"
													>
														<Box px="3.5" py="2" {...LABEL_PROPS} letterSpacing="0.12em" borderBottomWidth="1px" borderColor="border">
															Users not in this role
														</Box>
														{memberSearching ? (
															<Center py="3">
																<Spinner size="sm" color="accent" />
															</Center>
														) : (
															memberResults.map((result) => (
																<HStack
																	key={result.id}
																	as="button"
																	w="full"
																	gap="3"
																	px="3.5"
																	py="2.5"
																	cursor="pointer"
																	textAlign="left"
																	transition="background .15s ease-out"
																	_hover={{ bg: "rgba(139,92,246,0.14)" }}
																	onClick={() => handleAddMember(result)}
																>
																	<MemberAvatar id={result.id} name={result.name || result.email} />
																	<Box lineHeight="1.3" minW="0" flex="1">
																		<Text fontSize="sm" fontWeight="medium" color="fg" truncate>
																			{result.name || "—"}
																		</Text>
																		<Text fontSize="12px" color="fg.muted" truncate>
																			{result.email}
																		</Text>
																	</Box>
																	<Box color="aurora.violet">
																		<LuPlus size={14} />
																	</Box>
																</HStack>
															))
														)}
													</Box>
												)}
											</Box>
										)}
										{membersLoading ? (
											<Center py="10">
												<Spinner size="md" color="accent" />
											</Center>
										) : members.length === 0 ? (
											<Center py="10">
												<Text fontSize="sm" color="fg.muted">
													No members hold this role yet.
												</Text>
											</Center>
										) : (
											<Box>
												{members.map((member, index) => (
													<HStack
														key={`${member.user_id}-${member.app_service_id ?? "global"}`}
														gap="13px"
														px="4.5"
														py="3"
														borderBottomWidth={index < members.length - 1 ? "1px" : "0"}
														borderColor="border"
													>
														<MemberAvatar id={member.user_id} name={member.name || member.email} />
														<Box lineHeight="1.3" flex="1" minW="0">
															<Text fontSize="sm" fontWeight="medium" color="fg" truncate>
																{member.name || "—"}
															</Text>
															<Text fontSize="12px" color="fg.muted" truncate>
																{member.email}
															</Text>
														</Box>
														{member.app_service_id && (
															<Box
																{...PILL_BASE}
																fontSize="11px"
																color="aurora.cyan"
																borderColor="rgba(34,211,238,0.40)"
																bg="rgba(34,211,238,0.10)"
																title={`Scoped to app_service ${member.app_service_id}`}
															>
																scoped
															</Box>
														)}
														<Text fontSize="12px" color="fg.muted" whiteSpace="nowrap" css={{ fontVariantNumeric: "tabular-nums" }}>
															added {formatDateOnly(member.created_at)}
														</Text>
														{canAssign && (
															<Center
																as="button"
																w="8"
																h="8"
																borderRadius="9px"
																bg="rgba(7,7,26,0.55)"
																borderWidth="1px"
																borderColor="border.strong"
																color="aurora.magenta"
																cursor="pointer"
																title="Remove from role"
																aria-label="Remove from role"
																css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
																_hover={{ bg: "rgba(236,72,153,0.15)", borderColor: "rgba(236,72,153,0.40)" }}
																onClick={() => handleRemoveMember(member)}
															>
																<LuX size={14} />
															</Center>
														)}
													</HStack>
												))}
											</Box>
										)}
										<Flex
											align="center"
											justify="space-between"
											px="4.5"
											py="3"
											borderTopWidth="1px"
											borderColor="border"
											fontSize="12px"
											color="fg.muted"
											css={{ fontVariantNumeric: "tabular-nums" }}
										>
											<Text>
												{members.length} of {membersTotal.toLocaleString()} shown
											</Text>
											<HStack gap="2">
												<Center
													as="button"
													h="8"
													w="8"
													borderRadius="9px"
													borderWidth="1px"
													borderColor="border.strong"
													bg="bg.glass"
													color="fg.subtle"
													cursor={membersPage > 1 ? "pointer" : "not-allowed"}
													opacity={membersPage > 1 ? 1 : 0.35}
													_hover={membersPage > 1 ? { bg: "bg.glassHi", color: "fg" } : undefined}
													aria-label="Previous page"
													onClick={() => setMembersPage((previous) => Math.max(1, previous - 1))}
												>
													<LuChevronLeft size={14} />
												</Center>
												<Center
													as="button"
													h="8"
													w="8"
													borderRadius="9px"
													borderWidth="1px"
													borderColor="border.strong"
													bg="bg.glass"
													color="fg.subtle"
													cursor={membersPage < membersTotalPages ? "pointer" : "not-allowed"}
													opacity={membersPage < membersTotalPages ? 1 : 0.35}
													_hover={membersPage < membersTotalPages ? { bg: "bg.glassHi", color: "fg" } : undefined}
													aria-label="Next page"
													onClick={() => setMembersPage((previous) => Math.min(membersTotalPages, previous + 1))}
												>
													<LuChevronRight size={14} />
												</Center>
											</HStack>
										</Flex>
									</>
								)}

								{/* Tab 3 · app scope — read-only.
								    The mock drafted per-role app toggles, but the backend stores
								    scope per ASSIGNMENT (user_roles.app_service_id, set when adding
								    a member) — there is no role-level scope to toggle. This tab
								    shows the scope distribution of current assignments instead. */}
								{tab === "apps" && (
									<>
										<HStack gap="3.5" px="4.5" py="3.5" borderBottomWidth="1px" borderColor="border">
											<Center w="9" h="9" flex="none" borderRadius="11px" bg="bg.glass" borderWidth="1px" borderColor="border.strong" color="fg.muted">
												<LuGlobe size={16} />
											</Center>
											<Box lineHeight="1.3" flex="1">
												<Text fontSize="sm" fontWeight="semibold" color="fg">
													Scope is set per assignment
												</Text>
												<Text fontSize="12px" color="fg.muted">
													each member is added globally or scoped to one app service
												</Text>
											</Box>
										</HStack>
										{membersLoading ? (
											<Center py="10">
												<Spinner size="md" color="accent" />
											</Center>
										) : (
											<Box>
												<HStack gap="3.5" px="4.5" py="3.5" borderBottomWidth="1px" borderColor="border">
													<Center
														w="9"
														h="9"
														flex="none"
														borderRadius="11px"
														bg="rgba(139,92,246,0.15)"
														borderWidth="1px"
														borderColor="border.strong"
														color="aurora.violet"
													>
														<LuGlobe size={16} />
													</Center>
													<Box lineHeight="1.3" flex="1" minW="0">
														<Text fontSize="sm" fontWeight="semibold" color="fg">
															Global (all apps)
														</Text>
														<Text fontSize="12px" color="fg.muted" truncate>
															<Text as="span" color="fg.subtle">
																app_service_id = NULL
															</Text>{" "}
															· assignment applies everywhere
														</Text>
													</Box>
													<Box w="220px" flex="none" display="grid" gap="1.5">
														<Flex justify="space-between" fontSize="12px" color="fg.subtle" css={{ fontVariantNumeric: "tabular-nums" }}>
															<Text>assignments</Text>
															<Text as="b" color="fg" fontWeight="semibold">
																{scopeDistribution.globalCount} / {members.length} loaded
															</Text>
														</Flex>
														<Box h="6px" borderRadius="3px" bg="bg.glassHi" overflow="hidden">
															<Box
																h="full"
																borderRadius="3px"
																w={`${members.length > 0 ? (scopeDistribution.globalCount / members.length) * 100 : 0}%`}
																bg="linear-gradient(90deg, #22D3EE, #8B5CF6)"
																boxShadow="0 0 10px rgba(139,92,246,0.5)"
															/>
														</Box>
													</Box>
												</HStack>
												{[...scopeDistribution.perApp.entries()].map(([appServiceId, count]) => (
													<HStack key={appServiceId} gap="3.5" px="4.5" py="3.5" borderBottomWidth="1px" borderColor="border">
														<Center
															w="9"
															h="9"
															flex="none"
															borderRadius="11px"
															bg="rgba(34,211,238,0.12)"
															borderWidth="1px"
															borderColor="border.strong"
															color="aurora.cyan"
														>
															<LuAppWindow size={16} />
														</Center>
														<Box lineHeight="1.3" flex="1" minW="0">
															<Text fontSize="sm" fontWeight="semibold" color="fg" truncate>
																{appServiceId}
															</Text>
															<Text fontSize="12px" color="fg.muted">
																app_service_id
															</Text>
														</Box>
														<Box w="220px" flex="none" display="grid" gap="1.5">
															<Flex justify="space-between" fontSize="12px" color="fg.subtle" css={{ fontVariantNumeric: "tabular-nums" }}>
																<Text>assignments</Text>
																<Text as="b" color="fg" fontWeight="semibold">
																	{count} / {members.length} loaded
																</Text>
															</Flex>
															<Box h="6px" borderRadius="3px" bg="bg.glassHi" overflow="hidden">
																<Box
																	h="full"
																	borderRadius="3px"
																	w={`${members.length > 0 ? (count / members.length) * 100 : 0}%`}
																	bg="linear-gradient(90deg, #22D3EE, #8B5CF6)"
																	boxShadow="0 0 10px rgba(139,92,246,0.5)"
																/>
															</Box>
														</Box>
													</HStack>
												))}
											</Box>
										)}
										{/* TODO(backend): new assignments default to global — an
										    app_service list endpoint is needed before a scope picker
										    (and app names instead of raw ids) can be offered here. */}
										<HStack px="4.5" py="13px" gap="2" borderTopWidth="1px" borderColor="border" fontSize="12px" color="fg.muted">
											<LuInfo size={13} style={{ flex: "0 0 auto" }} />
											<Text>
												Read-only — scope stored per assignment on{" "}
												<Text as="span" color="fg.subtle">
													user_roles.app_service_id
												</Text>{" "}
												· NULL = global
											</Text>
										</HStack>
									</>
								)}
							</>
						)}
					</Box>
				</Grid>
			)}

			<CreateRoleDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				appId={selectedApp?.id ?? ""}
				appCode={selectedApp?.app_code ?? ""}
				roles={roles}
				onCreated={handleCreated}
			/>

			{/* Rename dialog (custom roles only) */}
			<Dialog.Root open={renameOpen} onOpenChange={(details) => setRenameOpen(details.open)} placement="center">
				<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
				<Dialog.Positioner>
					<Dialog.Content
						w="480px"
						maxW="92vw"
						borderRadius="20px"
						borderWidth="1px"
						borderColor="border.strong"
						bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
						color="fg"
						boxShadow="glassPop"
						overflow="hidden"
					>
						<Dialog.Header px="5" py="4" display="flex" alignItems="center" borderBottomWidth="1px" borderColor="border">
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Rename role
							</Dialog.Title>
						</Dialog.Header>
						<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									Name
								</Field.Label>
								<Input {...INPUT_PROPS} value={renameName} onChange={(event) => setRenameName(event.target.value)} />
							</Field.Root>
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									Description
								</Field.Label>
								<Input {...INPUT_PROPS} value={renameDescription} onChange={(event) => setRenameDescription(event.target.value)} />
							</Field.Root>
						</Dialog.Body>
						<HStack justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
							<Button
								variant="outline"
								h="11"
								px="4.5"
								borderRadius="glassSm"
								borderColor="border.strong"
								bg="bg.glass"
								fontSize="sm"
								fontWeight="semibold"
								color="fg"
								_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
								onClick={() => setRenameOpen(false)}
							>
								Cancel
							</Button>
							<Button
								h="11"
								px="4.5"
								borderRadius="glassSm"
								fontSize="sm"
								fontWeight="semibold"
								color="white"
								css={AURORA_CTA_STYLE}
								boxShadow="ctaGlow"
								_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
								disabled={!renameName.trim()}
								loading={renameSaving}
								onClick={handleRename}
							>
								Save
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>

			{/* Delete confirm dialog */}
			<Dialog.Root open={deleteOpen} onOpenChange={(details) => setDeleteOpen(details.open)} placement="center">
				<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
				<Dialog.Positioner>
					<Dialog.Content
						w="480px"
						maxW="92vw"
						borderRadius="20px"
						borderWidth="1px"
						borderColor="border.strong"
						bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
						color="fg"
						boxShadow="glassPop"
						overflow="hidden"
					>
						<Dialog.Header px="5" py="4" display="flex" alignItems="center" borderBottomWidth="1px" borderColor="border">
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Delete role {selectedRole?.code}
							</Dialog.Title>
						</Dialog.Header>
						<Dialog.Body p="5">
							<Text fontSize="sm" color="fg.subtle">
								This soft-deletes the role. The backend rejects the delete while members still hold it — reassign them first.
							</Text>
						</Dialog.Body>
						<HStack justify="flex-end" gap="2.5" px="5" py="4" borderTopWidth="1px" borderColor="border">
							<Button
								variant="outline"
								h="11"
								px="4.5"
								borderRadius="glassSm"
								borderColor="border.strong"
								bg="bg.glass"
								fontSize="sm"
								fontWeight="semibold"
								color="fg"
								_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
								onClick={() => setDeleteOpen(false)}
							>
								Cancel
							</Button>
							<Button
								h="11"
								px="4.5"
								variant="outline"
								borderRadius="glassSm"
								borderColor="rgba(236,72,153,0.35)"
								bg="rgba(236,72,153,0.08)"
								fontSize="sm"
								fontWeight="semibold"
								color="aurora.magenta"
								_hover={{ bg: "rgba(236,72,153,0.16)" }}
								loading={deleting}
								onClick={handleDelete}
							>
								<LuTrash2 size={14} /> Delete role
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>
		</AppShell>
	);
};
