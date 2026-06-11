"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { Box, Button, Center, Dialog, Field, Flex, Grid, Heading, HStack, Input, Spinner, Table, Text } from "@chakra-ui/react";
import {
	LuAppWindow,
	LuChevronLeft,
	LuChevronRight,
	LuEye,
	LuFileText,
	LuInfo,
	LuKeyRound,
	LuLibrary,
	LuLock,
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
	createPermissions,
	deletePermission,
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
import { AppTile } from "@/components/AppTile";
import { CreateRoleDialog } from "@/components/CreateRoleDialog";
import { EditPermissionAppearanceDialog } from "@/components/EditPermissionAppearanceDialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Tooltip } from "@/components/ui/tooltip";
import {
	APP_COLORS,
	APP_SERVICE_STATUS,
	PERMISSION_ICON_KEYS,
	renderDefaultPermissionIcon,
	renderPermissionIcon,
	resolveResourceIconKey,
} from "@/consts";
import { ColorSwatchPicker } from "@/components/ColorSwatchPicker";
import { IconPicker } from "@/components/IconPicker";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { useUser } from "@/hooks/useUser";
import { usePermissions } from "@/hooks/usePermissions";
import { AppShell } from "@/layouts/AppShell";
import type { AppService, PermissionItem, PermissionPair, RoleDetailResponse, RoleListItem, RoleMemberItem, UserListItem } from "@/types";
import { formatDateOnly } from "@/utils";

const MEMBERS_PAGE_SIZE = 8;

const NO_PERMISSION_TOOLTIP = "You don't have permission";

type RoleTab = "permissions" | "catalog" | "members";

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
	/** The RAW stored icon key (empty when unset) — seeds the appearance editor,
	 *  NOT the resolved/fallback key used for rendering. */
	iconKey: string;
	/** The RAW stored color key (empty when unset) — seeds the appearance editor,
	 *  NOT the cycling-palette fallback used for rendering. */
	colorKey: string;
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

/** Sanitize a resource/action segment: lowercase, strip anything not [a-z0-9_]. */
const sanitizeSegment = (value: string) => value.toLowerCase().replace(/[^a-z0-9_]/g, "");

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
	// the persisted icon + color keys per resource — all rows of a resource share
	// them, so the first row seen (catalog is id-ordered) is authoritative.
	const iconByResource = new Map<string, string>();
	const colorByResource = new Map<string, string>();
	for (const permission of catalog) {
		if (!byResource.has(permission.resource)) byResource.set(permission.resource, new Set());
		byResource.get(permission.resource)!.add(permission.action);
		if (!iconByResource.has(permission.resource)) iconByResource.set(permission.resource, permission.icon);
		if (!colorByResource.has(permission.resource)) colorByResource.set(permission.resource, permission.color);
	}
	return [...byResource.entries()]
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([resource, actions], index) => {
			const iconKey = resolveResourceIconKey(resource, iconByResource.get(resource));
			// a stored per-resource color wins; otherwise fall back to the cycling
			// palette so resources that have never been styled look unchanged.
			const storedColorKey = colorByResource.get(resource);
			const accent =
				storedColorKey && storedColorKey in APP_COLORS
					? {
							color: APP_COLORS[storedColorKey as keyof typeof APP_COLORS].hex,
							bg: `rgba(${APP_COLORS[storedColorKey as keyof typeof APP_COLORS].rgb},0.14)`,
						}
					: RESOURCE_ACCENTS[index % RESOURCE_ACCENTS.length];
			return {
			resource,
			label: resourceLabel(resource),
			accent,
			icon: renderPermissionIcon(iconKey, 14),
			// raw stored keys (empty when unset) — for seeding the appearance editor
			iconKey: iconByResource.get(resource) ?? "",
			colorKey: storedColorKey ?? "",
			crud: {
				read: actions.has("read"),
				create: actions.has("create"),
				update: actions.has("update"),
				delete: actions.has("delete"),
			},
			special: [...actions].filter((action) => !CRUD_SET.has(action)).sort(),
			};
		});
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

const RoleIconBox = ({
	code,
	icon,
	color,
	size = "9",
	radius = "11px",
	iconSize = 16,
	glow = false,
}: {
	code: string;
	/** Stored role icon key — wins over the code heuristic when present. */
	icon?: string;
	/** Stored role color palette key — wins over the code heuristic when present. */
	color?: string;
	size?: string;
	radius?: string;
	iconSize?: number;
	glow?: boolean;
}) => {
	// the code heuristic gives every role a tasteful default; a stored color/icon
	// overrides it so existing roles (empty appearance) look unchanged.
	const heuristic = roleAccent(code, iconSize);
	const hasStoredColor = !!color && color in APP_COLORS;
	const palette = hasStoredColor ? APP_COLORS[color as keyof typeof APP_COLORS] : null;
	const accentColor = palette ? palette.hex : heuristic.color;
	const accentBg = palette ? `rgba(${palette.rgb},0.15)` : heuristic.bg;
	const accentGlow = palette ? `rgba(${palette.rgb},0.18)` : heuristic.glow;
	const node = icon ? renderPermissionIcon(icon, iconSize) : heuristic.icon;
	return (
		<Center
			w={size}
			h={size}
			flex="none"
			borderRadius={radius}
			bg={accentBg}
			borderWidth="1px"
			borderColor="border.strong"
			color={accentColor}
			boxShadow={glow ? `0 0 18px ${accentGlow}` : "none"}
		>
			{node}
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

/**
 * "First run" empty state for an app whose permission catalog is empty (mock
 * .empty.first). Shown in both the matrix and catalog tabs; the CTA is hidden
 * for read-only apps (isme always has its seeded perms, so it never reaches here).
 */
const EmptyCatalogState = ({
	icon,
	title,
	description,
	canAdd,
	onAdd,
}: {
	icon: ReactNode;
	title: string;
	description: ReactNode;
	canAdd: boolean;
	onAdd: () => void;
}) => (
	<Center flexDirection="column" gap="3.5" px="6" py="14" textAlign="center">
		<Center
			w="14"
			h="14"
			borderRadius="16px"
			bg="linear-gradient(135deg, rgba(99,102,241,0.22), rgba(139,92,246,0.16))"
			borderWidth="1px"
			borderColor="aurora.violet"
			color="aurora.violet"
			boxShadow="0 0 26px rgba(139,92,246,0.22)"
		>
			{icon}
		</Center>
		<Box>
			<Heading fontSize="16px" fontWeight="semibold" color="fg">
				{title}
			</Heading>
			<Text fontSize="13px" color="fg.muted" mt="1.5" maxW="380px">
				{description}
			</Text>
		</Box>
		{canAdd && (
			<Button
				mt="1"
				h="11"
				px="4.5"
				borderRadius="glassSm"
				fontSize="sm"
				fontWeight="semibold"
				color="white"
				css={AURORA_CTA_STYLE}
				boxShadow="ctaGlow"
				_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
				onClick={onAdd}
			>
				<LuPlus size={16} /> Add your first permission
			</Button>
		)}
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
	const [renameIcon, setRenameIcon] = useState("");
	const [renameColor, setRenameColor] = useState("");
	const [renameSaving, setRenameSaving] = useState(false);
	const [deleteOpen, setDeleteOpen] = useState(false);
	const [deleting, setDeleting] = useState(false);

	// Add-permission dialog (create one or more resource:action pairs in the
	// selected app's catalog). `permissionQueue` holds extra pairs staged via
	// "Add another"; the live inputs are submitted alongside the queue.
	const [addPermissionOpen, setAddPermissionOpen] = useState(false);
	const [newResource, setNewResource] = useState("");
	const [newAction, setNewAction] = useState("");
	const [permissionQueue, setPermissionQueue] = useState<PermissionPair[]>([]);
	const [creatingPermission, setCreatingPermission] = useState(false);
	// Icon chosen for a BRAND-NEW resource; when the typed resource matches an
	// existing one the icon is locked to that resource's stored icon instead.
	const [newResourceIcon, setNewResourceIcon] = useState("");
	// Color chosen for a BRAND-NEW resource; when the typed resource matches an
	// existing one the color is locked to that resource's stored color instead.
	const [newResourceColor, setNewResourceColor] = useState("");

	// Remove-permission confirm dialog (delete a resource:action from the catalog).
	const [permissionToDelete, setPermissionToDelete] = useState<PermissionItem | null>(null);
	const [deletingPermission, setDeletingPermission] = useState(false);

	// Edit-appearance dialog target — the resource whose icon + color are edited
	// (shared by all its resource:action rows). Carries the RAW stored keys so the
	// editor seeds from the persisted appearance, not buildMatrix's fallback color.
	const [appearanceResource, setAppearanceResource] = useState<{
		resource: string;
		label: string;
		iconKey: string;
		colorKey: string;
	} | null>(null);

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
				// isme (system app) pinned first; the rest sorted by newest created first
				const sorted = [...response.items].sort((a, b) => {
					if (a.app_code === "isme") return -1;
					if (b.app_code === "isme") return 1;
					return (b.created_at ?? "").localeCompare(a.created_at ?? "");
				});
				setApps(sorted);
				// default to isme (the system app) when present, else the first app
				const isme = sorted.find((app) => app.app_code === "isme");
				setSelectedAppId(isme?.id ?? sorted[0]?.id ?? null);
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

	// Members tab needs the assignment list.
	useEffect(() => {
		if (!selectedRoleId || tab !== "members") return;
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
	// Selected app has an empty permission catalog → drive the "Add your first
	// permission" empty states (matrix + catalog tabs).
	const catalogEmpty = matrixResources.length === 0;

	const dirtyCount = useMemo(() => {
		let count = 0;
		for (const id of draftPermissionIds) if (!originalPermissionIds.has(id)) count += 1;
		for (const id of originalPermissionIds) if (!draftPermissionIds.has(id)) count += 1;
		return count;
	}, [draftPermissionIds, originalPermissionIds]);

	const membersTotalPages = Math.max(1, Math.ceil(membersTotal / MEMBERS_PAGE_SIZE));

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

	// Sanitized live inputs (lowercase, [a-z0-9_] only — same rule the backend enforces).
	const cleanResource = useMemo(() => sanitizeSegment(newResource), [newResource]);
	const cleanAction = useMemo(() => sanitizeSegment(newAction), [newAction]);

	// Resolved icon key per EXISTING resource (from the loaded catalog) — used to
	// detect when the typed resource already exists and to lock its icon.
	const existingResourceIconKeys = useMemo(() => {
		const byResource = new Map<string, string>();
		for (const permission of catalog) {
			if (!byResource.has(permission.resource)) {
				byResource.set(permission.resource, resolveResourceIconKey(permission.resource, permission.icon));
			}
		}
		return byResource;
	}, [catalog]);
	// Stored color key per EXISTING resource — locks the color the same way the
	// icon is locked when the typed resource already exists.
	const existingResourceColorKeys = useMemo(() => {
		const byResource = new Map<string, string>();
		for (const permission of catalog) {
			if (!byResource.has(permission.resource)) {
				byResource.set(permission.resource, permission.color ?? "");
			}
		}
		return byResource;
	}, [catalog]);
	// Does the typed resource match an existing one? If so its icon/color are locked.
	const matchedExistingResource = cleanResource ? existingResourceIconKeys.has(cleanResource) : false;
	const lockedResourceIconKey = matchedExistingResource ? (existingResourceIconKeys.get(cleanResource) ?? "") : "";
	const lockedResourceColorKey = matchedExistingResource ? (existingResourceColorKeys.get(cleanResource) ?? "") : "";
	// The icon/color that will actually be sent: locked for an existing resource,
	// the user's pick for a new one.
	const effectiveIconKey = matchedExistingResource ? lockedResourceIconKey : newResourceIcon;
	const effectiveColorKey = matchedExistingResource ? lockedResourceColorKey : newResourceColor;
	const livePermission = cleanResource && cleanAction ? `${cleanResource}:${cleanAction}` : "";
	// A live pair counts only when valid AND not already queued.
	const liveValid = !!livePermission && !permissionQueue.some((pair) => `${pair.resource}:${pair.action}` === livePermission);
	// Total pairs that would be created on submit (queue + the live pair, if any).
	const pendingPermissionCount = permissionQueue.length + (liveValid ? 1 : 0);

	const resetPermissionDialog = () => {
		setNewResource("");
		setNewAction("");
		setNewResourceIcon("");
		setNewResourceColor("");
		setPermissionQueue([]);
	};

	// "+ Add another" — stage the current inputs (with the resolved icon/color)
	// and clear them for the next pair.
	const handleQueuePermission = () => {
		if (!liveValid) return;
		setPermissionQueue((previous) => [
			...previous,
			{ resource: cleanResource, action: cleanAction, icon: effectiveIconKey, color: effectiveColorKey },
		]);
		setNewResource("");
		setNewAction("");
		setNewResourceIcon("");
		setNewResourceColor("");
	};

	const handleUnqueuePermission = (index: number) => {
		setPermissionQueue((previous) => previous.filter((_, position) => position !== index));
	};

	const handleCreatePermission = async () => {
		if (!selectedApp) return;
		// Submit every queued pair plus the live inputs (if a valid, unqueued pair).
		const payload: PermissionPair[] = [...permissionQueue];
		if (liveValid) payload.push({ resource: cleanResource, action: cleanAction, icon: effectiveIconKey, color: effectiveColorKey });
		if (payload.length === 0) return;
		setCreatingPermission(true);
		try {
			await createPermissions(selectedApp.id, payload);
			// re-pull the catalog so the new permissions appear in the matrix and can
			// be toggled onto roles via the existing save flow
			const refreshed = await listPermissions({ app_code: selectedApp.app_code });
			setCatalog(refreshed);
			setAddPermissionOpen(false);
			resetPermissionDialog();
			toaster.create({
				title:
					payload.length === 1
						? `Permission ${payload[0].resource}:${payload[0].action} created`
						: `${payload.length} permissions added to ${selectedApp.app_code}`,
				type: "success",
				meta: { closable: true },
			});
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to create permission"), type: "error", meta: { closable: true } });
		} finally {
			setCreatingPermission(false);
		}
	};

	const handleDeletePermission = async () => {
		if (!selectedApp || !permissionToDelete) return;
		const permission = permissionToDelete;
		setDeletingPermission(true);
		try {
			await deletePermission(permission.id);
			// Re-pull the catalog so the permission disappears from both the catalog
			// list and the matrix; drop it from the working draft so a stale grant
			// can't be re-saved onto the role.
			const refreshed = await listPermissions({ app_code: selectedApp.app_code });
			setCatalog(refreshed);
			setOriginalPermissionIds((previous) => {
				const next = new Set(previous);
				next.delete(permission.id);
				return next;
			});
			setDraftPermissionIds((previous) => {
				const next = new Set(previous);
				next.delete(permission.id);
				return next;
			});
			setPermissionToDelete(null);
			toaster.create({
				title: `Permission ${permission.resource}:${permission.action} removed`,
				type: "success",
				meta: { closable: true },
			});
		} catch (error: unknown) {
			toaster.create({ title: errorMessage(error, "Failed to remove permission"), type: "error", meta: { closable: true } });
		} finally {
			setDeletingPermission(false);
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
		setRenameIcon(detail?.icon ?? "");
		setRenameColor(detail?.color ?? "");
		setRenameOpen(true);
	};

	const handleRename = async () => {
		if (!selectedRoleId || !renameName.trim()) return;
		setRenameSaving(true);
		try {
			const trimmedName = renameName.trim();
			const trimmedDescription = renameDescription.trim();
			await updateRole(selectedRoleId, {
				name: trimmedName,
				description: trimmedDescription,
				icon: renameIcon,
				color: renameColor,
			});
			setDetail((previous) =>
				previous ? { ...previous, name: trimmedName, description: trimmedDescription, icon: renameIcon, color: renameColor } : previous
			);
			setRoles((previous) =>
				previous.map((role) =>
					role.id === selectedRoleId
						? { ...role, name: trimmedName, description: trimmedDescription, icon: renameIcon, color: renameColor }
						: role
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
				<Tooltip
					content={canCreate ? "isme roles are system-managed — read-only" : NO_PERMISSION_TOOLTIP}
					disabled={canCreate && !isIsmeApp}
					positioning={{ placement: "top" }}
				>
					<Box as="span">
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
							disabled={!canCreate || isIsmeApp}
							onClick={() => setCreateOpen(true)}
						>
							{isIsmeApp ? <LuLock size={16} /> : <LuPlus size={16} />} New role
						</Button>
					</Box>
				</Tooltip>
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
								<AppTile iconKey={app.icon} colorKey={app.color} size="chip" fallbackSeed={app.id} appCode={app.app_code} />
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
							isme's own roles and its permissions are seeded and{" "}
							<Text as="span" color="aurora.amber" fontWeight="semibold">
								read-only
							</Text>{" "}
							here. You can still assign these roles to users — only the role &amp; permission definitions are
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
									<RoleIconBox code={role.is_system ? role.code : "custom"} icon={role.icon} color={role.color} />
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
									<RoleIconBox code={detail.is_system ? detail.code : "custom"} icon={detail.icon} color={detail.color} size="11" radius="13px" iconSize={20} glow />
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
											<Tooltip content={NO_PERMISSION_TOOLTIP} disabled={canUpdate} positioning={{ placement: "top" }}>
												<Box as="span">
													<Button {...GHOST_SM_BUTTON_PROPS} disabled={!canUpdate} onClick={handleOpenRename}>
														<LuPencil size={13} /> Rename
													</Button>
												</Box>
											</Tooltip>
											{/* Delete is blocked while the role has members — the backend
											    rejects it too ("reassign first"); disabling here makes the
											    constraint visible without a round-trip. */}
											<Tooltip
												content={
													(selectedRole?.members_count ?? 0) > 0
														? `Role has ${selectedRole?.members_count} member${selectedRole?.members_count === 1 ? "" : "s"} — reassign them first`
														: NO_PERMISSION_TOOLTIP
												}
												disabled={canDelete && (selectedRole?.members_count ?? 0) === 0}
												positioning={{ placement: "top" }}
											>
												<Box as="span">
													<Button
														{...DANGER_SM_BUTTON_PROPS}
														disabled={!canDelete || (selectedRole?.members_count ?? 0) > 0}
														onClick={() => setDeleteOpen(true)}
													>
														<LuTrash2 size={13} /> Delete
													</Button>
												</Box>
											</Tooltip>
										</>
									)}
								</HStack>

								{/* Tabs */}
								<HStack gap="1" px="3.5" borderBottomWidth="1px" borderColor="border">
									{tabButton("permissions", "Permissions", <LuKeyRound size={14} />)}
										{tabButton("catalog", "Permission catalog", <LuLibrary size={14} />)}
									{tabButton(
										"members",
										"Members",
										<LuUsers size={14} />,
										(selectedRole?.members_count ?? 0).toLocaleString()
									)}
								</HStack>

								{/* Tab 1 · permission matrix */}
								{tab === "permissions" && (
									<>
										{/* Catalog toolbar — add a resource:action to the app's catalog.
										    Hidden for the isme system app (read-only catalog) and when
										    the catalog is empty (the empty-state CTA covers it). The button
										    is disabled (with a perm tooltip) when role:create is missing. */}
										{!isIsmeApp && !catalogEmpty && (
											<Flex
												align="center"
												justify="space-between"
												gap="3"
												px="4.5"
												py="2.5"
												borderBottomWidth="1px"
												borderColor="border"
												wrap="wrap"
											>
												<Text fontSize="12px" color="fg.muted">
													Permissions belong to{" "}
													<Text as="span" color="aurora.cyan" fontWeight="semibold">
														{selectedApp?.app_code}
													</Text>
													's catalog · add new resource:action pairs, then grant them to a role.
												</Text>
												<Tooltip content={NO_PERMISSION_TOOLTIP} disabled={canCreate} positioning={{ placement: "top" }}>
													<Box as="span">
														<Button
															h="9"
															px="3.5"
															flex="none"
															borderRadius="glassSm"
															fontSize="13px"
															fontWeight="semibold"
															color="white"
															css={AURORA_CTA_STYLE}
															boxShadow="ctaGlow"
															_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
															_focusVisible={{ boxShadow: "focusRing" }}
															disabled={!canCreate}
															onClick={() => setAddPermissionOpen(true)}
														>
															<LuPlus size={14} /> Add permission
														</Button>
													</Box>
												</Tooltip>
											</Flex>
										)}
										{isReadOnly && (
											<HStack gap="2" px="4.5" py="2.5" borderBottomWidth="1px" borderColor="rgba(245,158,11,0.25)" bg="rgba(245,158,11,0.06)" fontSize="12px" color="aurora.amber">
												<LuLock size={13} />
												{isIsmeApp
													? "System-managed app — isme permissions are seeded and read-only"
													: "System role — permissions are seeded and read-only"}
											</HStack>
										)}
										{/* Empty catalog → "first permission" CTA instead of an empty grid (mock renderMatrix empty state) */}
										{catalogEmpty ? (
											<EmptyCatalogState
												icon={<LuKeyRound size={26} />}
												title="No permissions yet"
												description={
													<>
														This app starts with an empty permission catalog. Create your first{" "}
														<Text as="span" color="aurora.cyan">
															resource:action
														</Text>{" "}
														permission — then toggle it onto the{" "}
														<Text as="span" color="fg.subtle" fontWeight="semibold">
															{selectedRole?.code}
														</Text>{" "}
														role here.
													</>
												}
												canAdd={!isReadOnly && canCreate}
												onAdd={() => setAddPermissionOpen(true)}
											/>
										) : (
										/* read-only matrix gets a faint amber wash (mock .matrix-readonly) */
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
										)}

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

								{/* Tab 2 · permission catalog — flat resource:action list for the app */}
								{tab === "catalog" && (
									<>
										{/* Catalog header bar — perm count + add-permission affordance.
										    isme renders a locked amber button (read-only catalog). */}
										<Flex
											align="center"
											justify="space-between"
											gap="3"
											px="4.5"
											py="3"
											borderBottomWidth="1px"
											borderColor="border"
											wrap="wrap"
										>
											<Text fontSize="12px" color="fg.muted">
												{isIsmeApp ? (
													<>
														<Text as="span" color="fg.subtle" fontWeight="semibold">
															{catalog.length} permission{catalog.length === 1 ? "" : "s"}
														</Text>{" "}
														· seeded &amp; system-managed (read-only)
													</>
												) : (
													<>
														<Text as="span" color="fg.subtle" fontWeight="semibold">
															{catalog.length} permission{catalog.length === 1 ? "" : "s"}
														</Text>{" "}
														· add{" "}
														<Text as="span" color="aurora.cyan">
															resource:action
														</Text>{" "}
														pairs, then toggle them onto roles
													</>
												)}
											</Text>
											{isIsmeApp ? (
												<Button
													{...GHOST_SM_BUTTON_PROPS}
													color="aurora.amber"
													borderColor="rgba(245,158,11,0.35)"
													bg="rgba(245,158,11,0.08)"
													disabled
													title="isme permissions are system-managed — read-only"
												>
													<LuPlus size={13} /> Add permission <LuLock size={12} />
												</Button>
											) : (
												<Tooltip content={NO_PERMISSION_TOOLTIP} disabled={canCreate} positioning={{ placement: "top" }}>
													<Box as="span">
														<Button
															h="9"
															px="3.5"
															flex="none"
															borderRadius="glassSm"
															fontSize="13px"
															fontWeight="semibold"
															color="white"
															css={AURORA_CTA_STYLE}
															boxShadow="ctaGlow"
															_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
															_focusVisible={{ boxShadow: "focusRing" }}
															disabled={!canCreate}
															onClick={() => setAddPermissionOpen(true)}
														>
															<LuPlus size={14} /> Add permission
														</Button>
													</Box>
												</Tooltip>
											)}
										</Flex>

										{catalogEmpty ? (
											<EmptyCatalogState
												icon={<LuLibrary size={26} />}
												title="Empty permission catalog"
												description={
													<>
														New apps don&apos;t auto-seed permissions. Create{" "}
														<Text as="span" color="aurora.cyan">
															resource:action
														</Text>{" "}
														entries (e.g.{" "}
														<Text as="span" color="aurora.cyan">
															track:read
														</Text>
														) and they become available to toggle onto roles.
													</>
												}
												canAdd={!isIsmeApp && canCreate}
												onAdd={() => setAddPermissionOpen(true)}
											/>
										) : (
											<>
												<Table.Root size="sm" css={{ "& td, & th": { borderColor: "rgba(255,255,255,0.10)" } }}>
													<Table.Header>
														<Table.Row bg="transparent">
															<Table.ColumnHeader px="4.5" py="3" bg="transparent" {...LABEL_PROPS}>
																Resource
															</Table.ColumnHeader>
															<Table.ColumnHeader px="2" py="3" bg="transparent" {...LABEL_PROPS}>
																Action
															</Table.ColumnHeader>
															<Table.ColumnHeader px="4.5" py="3" bg="transparent" {...LABEL_PROPS}>
																Permission
															</Table.ColumnHeader>
															{!isIsmeApp && (
																<Table.ColumnHeader w="56px" px="4.5" py="3" bg="transparent" textAlign="right" {...LABEL_PROPS} />
															)}
														</Table.Row>
													</Table.Header>
													<Table.Body>
														{matrixResources.map((resource) => {
															const actions = [
																...CRUD_ACTIONS.filter((action) => resource.crud[action]),
																...resource.special,
															];
															return actions.map((action, index) => (
																<Table.Row
																	key={`${resource.resource}:${action}`}
																	bg="transparent"
																	transition="background .15s ease-out"
																	_hover={{ bg: "rgba(255,255,255,0.03)" }}
																>
																	<Table.Cell px="4.5" py="13px">
																		{index === 0 ? (
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
																				{!isIsmeApp &&
																					(() => {
																						const pencil = (
																							<Center
																								as="button"
																								display="inline-flex"
																								flex="none"
																								w="7"
																								h="7"
																								borderRadius="8px"
																								bg="bg.glass"
																								borderWidth="1px"
																								borderColor="border.strong"
																								color="fg.muted"
																								opacity={canUpdate ? 0.5 : 0.35}
																								cursor={canUpdate ? "pointer" : "not-allowed"}
																								title={canUpdate ? "Edit appearance" : undefined}
																								aria-label={`Edit appearance for ${resource.resource}`}
																								css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
																								_hover={
																									canUpdate
																										? {
																												opacity: 1,
																												color: "aurora.violet",
																												borderColor: "aurora.violet",
																												bg: "rgba(139,92,246,0.12)",
																											}
																										: undefined
																								}
																								onClick={() => {
																									if (canUpdate && !isIsmeApp) {
																										setAppearanceResource({
																											resource: resource.resource,
																											label: resource.label,
																											iconKey: resource.iconKey,
																											colorKey: resource.colorKey,
																										});
																									}
																								}}
																							>
																								<LuPencil size={13} />
																							</Center>
																						);
																						if (canUpdate) return pencil;
																						return (
																							<Tooltip content={NO_PERMISSION_TOOLTIP} positioning={{ placement: "top" }}>
																								{pencil}
																							</Tooltip>
																						);
																					})()}
																			</HStack>
																		) : null}
																	</Table.Cell>
																	<Table.Cell px="2" py="13px">
																		<Text fontSize="13px" color="fg.subtle">
																			{action}
																		</Text>
																	</Table.Cell>
																	<Table.Cell px="4.5" py="13px">
																		<Text as="code" fontSize="13px" color="aurora.cyan" fontFamily="inherit">
																			{resource.resource}:{action}
																		</Text>
																	</Table.Cell>
																	{!isIsmeApp && (
																		<Table.Cell px="4.5" py="13px" textAlign="right">
																			{(() => {
																				const permission = permissionByCode.get(`${resource.resource}:${action}`);
																				if (!permission) return null;
																				const trash = (
																					<Center
																						as="button"
																						display="inline-flex"
																						w="8"
																						h="8"
																						borderRadius="9px"
																						bg="rgba(7,7,26,0.55)"
																						borderWidth="1px"
																						borderColor="border.strong"
																						color="aurora.magenta"
																						cursor={canDelete ? "pointer" : "not-allowed"}
																						opacity={canDelete ? 1 : 0.35}
																						title={canDelete ? "Remove permission" : undefined}
																						aria-label={`Remove permission ${resource.resource}:${action}`}
																						css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
																						_hover={canDelete ? { bg: "rgba(236,72,153,0.15)", borderColor: "rgba(236,72,153,0.40)" } : undefined}
																						onClick={() => {
																							if (canDelete) setPermissionToDelete(permission);
																						}}
																					>
																						<LuTrash2 size={13} />
																					</Center>
																				);
																				if (canDelete) return trash;
																				return (
																					<Tooltip content={NO_PERMISSION_TOOLTIP} positioning={{ placement: "top" }}>
																						{trash}
																					</Tooltip>
																				);
																			})()}
																		</Table.Cell>
																	)}
																</Table.Row>
															));
														})}
													</Table.Body>
												</Table.Root>
												{/* Catalog foot — lock notice (isme) vs neutral hint (editable).
												    Editable apps get a per-row trash button (DELETE /permissions/:id);
												    isme's catalog is read-only so no delete affordance is shown. */}
												<HStack
													gap="2"
													px="4.5"
													py="3"
													borderTopWidth="1px"
													borderColor={isIsmeApp ? "rgba(245,158,11,0.25)" : "border"}
													bg={isIsmeApp ? "rgba(245,158,11,0.06)" : "rgba(7,7,26,0.35)"}
													fontSize="12px"
													color={isIsmeApp ? "aurora.amber" : "fg.muted"}
													align="flex-start"
												>
													<Box mt="0.5" flex="none">
														{isIsmeApp ? <LuLock size={13} /> : <LuInfo size={13} />}
													</Box>
													<Text>
														{isIsmeApp ? (
															<>
																<Text as="span" fontWeight="semibold">
																	System-managed
																</Text>{" "}
																— isme&apos;s permissions are seeded and cannot be added or removed. Permission strings are{" "}
																<Text as="span" color="fg.subtle">
																	resource:action
																</Text>
																, never app-prefixed.
															</>
														) : (
															<>
																Permission strings are{" "}
																<Text as="span" color="fg.subtle">
																	resource:action
																</Text>{" "}
																— not app-prefixed. Use &quot;Add permission&quot; above to create more, then toggle them onto roles.
															</>
														)}
													</Text>
												</HStack>
											</>
										)}
									</>
								)}

								{/* Tab 3 · members */}
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
														<Tooltip content={NO_PERMISSION_TOOLTIP} disabled={canAssign} positioning={{ placement: "top" }}>
															<Center
																as="button"
																w="8"
																h="8"
																borderRadius="9px"
																bg="rgba(7,7,26,0.55)"
																borderWidth="1px"
																borderColor="border.strong"
																color="aurora.magenta"
																cursor={canAssign ? "pointer" : "not-allowed"}
																opacity={canAssign ? 1 : 0.35}
																title={canAssign ? "Remove from role" : undefined}
																aria-label="Remove from role"
																css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
																_hover={canAssign ? { bg: "rgba(236,72,153,0.15)", borderColor: "rgba(236,72,153,0.40)" } : undefined}
																onClick={() => {
																	if (canAssign) handleRemoveMember(member);
																}}
															>
																<LuX size={14} />
															</Center>
														</Tooltip>
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
								Edit role
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
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									Icon{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										· role badge
									</Text>
								</Field.Label>
								<IconPicker value={renameIcon} onChange={setRenameIcon} ariaLabel="Role icon" />
							</Field.Root>
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									Color{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										· aurora palette
									</Text>
								</Field.Label>
								<ColorSwatchPicker value={renameColor} onChange={setRenameColor} ariaLabel="Role color" />
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

			{/* Add-permission dialog — creates a resource:action in the app's catalog */}
			<Dialog.Root
				open={addPermissionOpen}
				onOpenChange={(details) => {
					setAddPermissionOpen(details.open);
					if (!details.open) resetPermissionDialog();
				}}
				placement="center"
			>
				<Dialog.Backdrop bg="rgba(4,4,14,0.70)" css={{ backdropFilter: "blur(10px)", WebkitBackdropFilter: "blur(10px)" }} />
				<Dialog.Positioner>
					<Dialog.Content
						w="520px"
						maxW="92vw"
						borderRadius="20px"
						borderWidth="1px"
						borderColor="border.strong"
						bg="linear-gradient(180deg, rgba(18,18,46,0.92), rgba(11,11,35,0.94))"
						color="fg"
						boxShadow="glassPop"
						overflow="hidden"
					>
						<Dialog.Header px="5" py="4" display="flex" alignItems="center" gap="2.5" borderBottomWidth="1px" borderColor="border">
							<Center w="7" h="7" flex="none" borderRadius="9px" bg="rgba(139,92,246,0.15)" borderWidth="1px" borderColor="border.strong" color="aurora.violet">
								<LuKeyRound size={15} />
							</Center>
							<Dialog.Title fontSize="15px" fontWeight="semibold">
								Add permission to{" "}
								<Text as="span" color="aurora.violet">
									{selectedApp?.app_code}
								</Text>
							</Dialog.Title>
						</Dialog.Header>
						<Dialog.Body p="5" display="flex" flexDirection="column" gap="4">
							{/* App field — bound to the selected app, read-only (app-owned RBAC). */}
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									App{" "}
									<Text as="span" color="fg.muted" fontWeight="normal">
										· bound to the selected app, not editable
									</Text>
								</Field.Label>
								<Box position="relative" width="full" opacity={0.7}>
									<Box position="absolute" left="3.5" top="50%" transform="translateY(-50%)" color="fg.muted" pointerEvents="none">
										<LuAppWindow size={16} />
									</Box>
									<Input {...INPUT_PROPS} w="full" pl="10" readOnly value={selectedApp ? `${selectedApp.app_name} · ${selectedApp.app_code}` : ""} />
								</Box>
							</Field.Root>

							{/* Resource + Action — composed into resource:action. */}
							<Grid templateColumns="1fr 1fr" gap="3.5">
								<Field.Root>
									<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
										Resource
									</Field.Label>
									<Box position="relative" width="full">
										<Box position="absolute" left="3.5" top="50%" transform="translateY(-50%)" color="fg.muted" pointerEvents="none">
											<LuFileText size={16} />
										</Box>
										<Input
											{...INPUT_PROPS}
											w="full"
											pl="10"
											placeholder="report"
											autoComplete="off"
											value={newResource}
											onChange={(event) => setNewResource(sanitizeSegment(event.target.value))}
										/>
									</Box>
								</Field.Root>
								<Field.Root>
									<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
										Action
									</Field.Label>
									<Box position="relative" width="full">
										<Box position="absolute" left="3.5" top="50%" transform="translateY(-50%)" color="fg.muted" pointerEvents="none">
											<LuKeyRound size={16} />
										</Box>
										<Input
											{...INPUT_PROPS}
											w="full"
											pl="10"
											placeholder="export"
											autoComplete="off"
											value={newAction}
											onChange={(event) => setNewAction(sanitizeSegment(event.target.value))}
										/>
									</Box>
								</Field.Root>
							</Grid>

							{/* Resource icon — picker for a new resource, locked to the
							    existing resource's icon when the name matches one. */}
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									Icon{" "}
									<Text as="span" color="fg.muted" fontWeight="normal">
										· shown next to this resource everywhere
									</Text>
								</Field.Label>
								{matchedExistingResource ? (
									<HStack
										gap="2.5"
										px="3.5"
										py="2.5"
										borderRadius="glassSm"
										borderWidth="1px"
										borderColor="border"
										bg="bg.glass"
										color="fg.muted"
									>
										<Center flex="none" w="30px" h="30px" borderRadius="8px" bg="bg.glassHi" color="aurora.cyan">
											{renderPermissionIcon(lockedResourceIconKey, 16)}
										</Center>
										<Box>
											<Text fontSize="13px" fontWeight="medium" color="fg">
												Locked
											</Text>
											<Text fontSize="12px" color="fg.muted">
												icon set by existing resource
											</Text>
										</Box>
										<Box ml="auto" flex="none" color="fg.muted">
											<LuLock size={14} />
										</Box>
									</HStack>
								) : (
									<Grid w="full" templateColumns="repeat(auto-fill, 38px)" justifyContent="start" gap="2">
										{/* Neutral default option (empty key). */}
										<Center
											as="button"
											aria-label="No icon (default)"
											aria-pressed={newResourceIcon === ""}
											title="Default"
											w="38px"
											h="38px"
											borderRadius="9px"
											borderWidth="1px"
											cursor="pointer"
											borderColor={newResourceIcon === "" ? "aurora.cyan" : "border"}
											bg={newResourceIcon === "" ? "rgba(34,211,238,0.12)" : "bg.glass"}
											color={newResourceIcon === "" ? "aurora.cyan" : "fg.muted"}
											_hover={{ borderColor: "border.strong", color: "fg" }}
											onClick={() => setNewResourceIcon("")}
										>
											{renderDefaultPermissionIcon(16)}
										</Center>
										{PERMISSION_ICON_KEYS.map((key) => {
											const selected = newResourceIcon === key;
											return (
												<Center
													key={key}
													as="button"
													aria-label={key}
													aria-pressed={selected}
													title={key}
													w="38px"
													h="38px"
													borderRadius="9px"
													borderWidth="1px"
													cursor="pointer"
													borderColor={selected ? "aurora.cyan" : "border"}
													bg={selected ? "rgba(34,211,238,0.12)" : "bg.glass"}
													color={selected ? "aurora.cyan" : "fg.muted"}
													_hover={{ borderColor: "border.strong", color: "fg" }}
													onClick={() => setNewResourceIcon(key)}
												>
													{renderPermissionIcon(key, 16)}
												</Center>
											);
										})}
									</Grid>
								)}
							</Field.Root>

							{/* Resource color — picker for a new resource, locked to the
							    existing resource's color when the name matches one (mirrors icon). */}
							<Field.Root>
								<Field.Label fontSize="13px" fontWeight="medium" color="fg.subtle">
									Color{" "}
									<Text as="span" color="fg.muted" fontWeight="normal">
										· tints this resource everywhere
									</Text>
								</Field.Label>
								{matchedExistingResource ? (
									<HStack
										gap="2.5"
										px="3.5"
										py="2.5"
										borderRadius="glassSm"
										borderWidth="1px"
										borderColor="border"
										bg="bg.glass"
										color="fg.muted"
									>
										<Center
											flex="none"
											w="30px"
											h="30px"
											borderRadius="full"
											css={{
												background:
													lockedResourceColorKey && lockedResourceColorKey in APP_COLORS
														? APP_COLORS[lockedResourceColorKey as keyof typeof APP_COLORS].hex
														: "rgba(255,255,255,0.10)",
											}}
										/>
										<Box>
											<Text fontSize="13px" fontWeight="medium" color="fg">
												Locked
											</Text>
											<Text fontSize="12px" color="fg.muted">
												color set by existing resource
											</Text>
										</Box>
										<Box ml="auto" flex="none" color="fg.muted">
											<LuLock size={14} />
										</Box>
									</HStack>
								) : (
									<ColorSwatchPicker value={newResourceColor} onChange={setNewResourceColor} ariaLabel="Resource color" />
								)}
							</Field.Root>

							{/* Live resource:action preview pill (dashed container). */}
							<HStack
								gap="2.5"
								wrap="wrap"
								px="3.5"
								py="13px"
								borderRadius="glassSm"
								borderWidth="1px"
								borderStyle="dashed"
								borderColor="border.strong"
								bg="rgba(7,7,26,0.35)"
							>
								<Text {...LABEL_PROPS} letterSpacing="0.12em">
									Preview
								</Text>
								{livePermission ? (
									<Box
										{...PILL_BASE}
										px="3"
										py="5px"
										fontSize="14px"
										fontWeight="semibold"
										color="aurora.cyan"
										borderColor="rgba(34,211,238,0.35)"
										bg="rgba(34,211,238,0.10)"
										css={{ fontVariantNumeric: "tabular-nums" }}
									>
										<Box as="span" display="inline-flex" alignItems="center" gap="2">
											{renderPermissionIcon(effectiveIconKey, 14)}
											<Box as="span">
												{cleanResource}
												<Text as="span" color="fg.muted" fontWeight="bold">
													:
												</Text>
												{cleanAction}
											</Box>
										</Box>
									</Box>
								) : (
									<Box {...PILL_BASE} px="3" py="5px" fontSize="14px">
										resource:action
									</Box>
								)}
							</HStack>

							{/* Validation hint — namespace is implicit, no app prefix. */}
							<Text fontSize="12px" color="fg.muted">
								Lowercase letters, digits and underscore only · no spaces or{" "}
								<Text as="code" color="fg.subtle" fontFamily="inherit">
									:
								</Text>{" "}
								(it&apos;s the separator). The app namespace is implicit — don&apos;t prefix with the app name.
							</Text>

							{/* "+ Add another" — stage multiple pairs before submitting. */}
							<Box>
								<Button {...GHOST_SM_BUTTON_PROPS} disabled={!liveValid} onClick={handleQueuePermission}>
									<LuPlus size={13} /> Add another
								</Button>
								{permissionQueue.length > 0 && (
									<Box display="grid" gap="2" mt="2.5">
										{permissionQueue.map((entry, index) => {
											const entryCode = `${entry.resource}:${entry.action}`;
											return (
											<HStack
												key={entryCode}
												gap="2.5"
												px="3"
												py="2"
												borderRadius="10px"
												borderWidth="1px"
												borderColor="border"
												bg="bg.glass"
											>
												<Center flex="none" w="22px" h="22px" borderRadius="6px" bg="bg.glassHi" color="aurora.cyan">
													{renderPermissionIcon(entry.icon, 13)}
												</Center>
												<Text as="code" flex="1" fontSize="13px" fontWeight="semibold" color="aurora.cyan" fontFamily="inherit">
													{entryCode}
												</Text>
												<Center
													as="button"
													w="26px"
													h="26px"
													borderRadius="7px"
													color="fg.muted"
													cursor="pointer"
													title="Remove"
													aria-label={`Remove ${entryCode}`}
													_hover={{ bg: "rgba(236,72,153,0.12)", color: "aurora.magenta" }}
													onClick={() => handleUnqueuePermission(index)}
												>
													<LuX size={13} />
												</Center>
											</HStack>
											);
										})}
									</Box>
								)}
							</Box>

							{/* Footer explainer (mock .modal-note). */}
							<HStack
								gap="2.5"
								px="3.5"
								py="11px"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border"
								bg="bg.glass"
								fontSize="12px"
								color="fg.muted"
								align="flex-start"
							>
								<Box mt="0.5" flex="none">
									<LuInfo size={14} />
								</Box>
								<Text>
									<Text as="span" color="aurora.amber" fontWeight="semibold">
										resource:action
									</Text>{" "}
									— once created, this permission becomes toggle-able onto any role in this app via the matrix. It grants nothing
									until a role is checked.
								</Text>
							</HStack>
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
								onClick={() => setAddPermissionOpen(false)}
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
								disabled={pendingPermissionCount === 0}
								loading={creatingPermission}
								onClick={handleCreatePermission}
							>
								<LuPlus size={14} />{" "}
								{pendingPermissionCount > 1 ? `Add ${pendingPermissionCount} permissions` : "Add permission"}
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>

			{/* Edit-appearance dialog — changes a resource's icon + color across all
			    its catalog rows (rejected for the isme system app). Seeds from the
			    RAW stored keys carried on the matrix resource. */}
			<EditPermissionAppearanceDialog
				open={!!appearanceResource}
				onOpenChange={(open) => {
					if (!open) setAppearanceResource(null);
				}}
				appId={selectedApp?.id ?? ""}
				resource={appearanceResource?.resource ?? ""}
				label={appearanceResource?.label ?? ""}
				savedIcon={appearanceResource?.iconKey ?? ""}
				savedColor={appearanceResource?.colorKey ?? ""}
				canEdit={canUpdate}
				onSaved={async () => {
					if (selectedApp) setCatalog(await listPermissions({ app_code: selectedApp.app_code }));
					setAppearanceResource(null);
				}}
			/>

			{/* Remove-permission confirm dialog — deletes a resource:action from the
			    app's catalog (and clears any role grants referencing it). */}
			<Dialog.Root
				open={!!permissionToDelete}
				onOpenChange={(details) => {
					if (!details.open) setPermissionToDelete(null);
				}}
				placement="center"
			>
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
								Remove permission
							</Dialog.Title>
						</Dialog.Header>
						<Dialog.Body p="5" display="flex" flexDirection="column" gap="3">
							<Text fontSize="sm" color="fg.subtle">
								This deletes{" "}
								<Text as="code" color="aurora.cyan" fontFamily="inherit">
									{permissionToDelete?.resource}:{permissionToDelete?.action}
								</Text>{" "}
								from{" "}
								<Text as="span" color="aurora.cyan" fontWeight="semibold">
									{selectedApp?.app_code}
								</Text>
								's catalog and removes it from every role that currently grants it.
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
								onClick={() => setPermissionToDelete(null)}
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
								loading={deletingPermission}
								onClick={handleDeletePermission}
							>
								<LuTrash2 size={14} /> Remove permission
							</Button>
						</HStack>
					</Dialog.Content>
				</Dialog.Positioner>
			</Dialog.Root>
		</AppShell>
	);
};
