"use client";

import { Fragment, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
	Box,
	Button,
	Center,
	Flex,
	Heading,
	HStack,
	Input,
	Menu,
	NativeSelect,
	Portal,
	Spinner,
	Table,
	Text,
} from "@chakra-ui/react";
import {
	LuBadgeCheck,
	LuBan,
	LuChevronLeft,
	LuChevronRight,
	LuCircleCheck,
	LuCircleSlash,
	LuDownload,
	LuFilter,
	LuInfo,
	LuKeyRound,
	LuLink,
	LuLogOut,
	LuMail,
	LuMonitor,
	LuMonitorSmartphone,
	LuPencil,
	LuRefreshCw,
	LuSearch,
	LuShieldQuestion,
	LuTrash2,
	LuUserPlus,
	LuUsers,
	LuX,
} from "react-icons/lu";
import {
	assignUserRole,
	listInvitations,
	listRoles,
	listUsers,
	listUserSessions,
	revokeInvitation,
	revokeUserSession,
	softDeleteUser,
	updateUserStatus,
} from "@/apis";
import { AppRoleChip } from "@/components/AppRoleChip";
import { InviteUserDialog } from "@/components/InviteUserDialog";
import { VerifyUserDialog } from "@/components/VerifyUserDialog";
import { Checkbox } from "@/components/ui/checkbox";
import { toaster } from "@/components/ui/toaster";
import {
	USER_LAST_LOGIN_FILTER_OPTIONS,
	USER_STATUS,
	USER_STATUS_FILTER_OPTIONS,
	USER_VERIFIED_FILTER_OPTIONS,
	USER_ROLE_OPTIONS,
	USERS_PAGE_SIZE_OPTIONS,
} from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { useUser } from "@/hooks/useUser";
import { usePermissions } from "@/hooks/usePermissions";
import { AppShell } from "@/layouts/AppShell";
import { INVITATION_STATUS_LABELS } from "@/types";
import type {
	AppRole,
	InvitationDisplayStatus,
	InvitationListItem,
	InvitationRoleAssignment,
	RoleListItem,
	UserListItem,
	UserSessionItem,
	UserStatus,
} from "@/types";
import { formatDateOnly, formatShortDateTime, isZeroTime, parseUserAgent } from "@/utils";

type StatusFilter = (typeof USER_STATUS_FILTER_OPTIONS)[number];
type VerifiedFilter = (typeof USER_VERIFIED_FILTER_OPTIONS)[number];
type LastLoginFilter = (typeof USER_LAST_LOGIN_FILTER_OPTIONS)[number]["value"];

const STATUS_FILTER_TO_VALUE: Record<Exclude<StatusFilter, "all">, UserStatus> = {
	active: 1,
	inactive: 2,
	pending: 3,
};

const STATUS_FILTER_ICONS: Partial<Record<StatusFilter, React.ReactNode>> = {
	active: <LuCircleCheck size={13} />,
	inactive: <LuCircleSlash size={13} />,
	pending: <LuMail size={13} />,
};

const VERIFIED_FILTER_ICONS: Partial<Record<VerifiedFilter, React.ReactNode>> = {
	verified: <LuBadgeCheck size={13} />,
	unverified: <LuShieldQuestion size={13} />,
};

const LAST_LOGIN_WINDOW_MS: Record<string, number> = {
	"24h": 24 * 60 * 60 * 1000,
	"7d": 7 * 24 * 60 * 60 * 1000,
	"30d": 30 * 24 * 60 * 60 * 1000,
};

const AVATAR_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
	"linear-gradient(135deg, #F59E0B, #EC4899)",
	"linear-gradient(135deg, #34D399, #22D3EE)",
];

/** Uppercase micro-label (mock .flabel / thead th). */
const LABEL_PROPS = {
	fontSize: "11px",
	fontWeight: "semibold",
	letterSpacing: "0.14em",
	textTransform: "uppercase",
	color: "fg.muted",
} as const;

/** Glass pill (mock .pill). */
const PILL_BASE = {
	display: "inline-flex",
	alignItems: "center",
	gap: "7px",
	px: "11px",
	py: "1",
	borderRadius: "full",
	fontSize: "12px",
	fontWeight: "medium",
	borderWidth: "1px",
	borderColor: "border.strong",
	bg: "bg.glass",
	whiteSpace: "nowrap",
} as const;

/** Tabular numerals (mock .num). */
const NUM_PROPS = {
	fontSize: "13px",
	color: "fg.subtle",
	css: { fontVariantNumeric: "tabular-nums" },
} as const;

/** Aurora gradient checkbox control (mock .check input:checked). */
const CHECKBOX_CSS = {
	"& .chakra-checkbox__control": {
		borderRadius: "5px",
		borderColor: "rgba(255,255,255,0.18)",
		background: "rgba(255,255,255,0.06)",
	},
	"& .chakra-checkbox__control[data-state=checked], & .chakra-checkbox__control[data-state=indeterminate]": {
		background: "linear-gradient(135deg, #6366F1, #8B5CF6)",
		borderColor: "#8B5CF6",
		boxShadow: "0 0 12px rgba(139,92,246,0.45)",
	},
} as const;

const GHOST_BUTTON_PROPS = {
	variant: "outline",
	borderRadius: "glassSm",
	borderColor: "border.strong",
	bg: "bg.glass",
	color: "fg",
	fontWeight: "semibold",
	css: { backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" },
	_hover: { bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" },
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

const StatusPill = ({ status }: { status: UserStatus }) => {
	if (status === USER_STATUS.ACTIVE) {
		return (
			<Box {...PILL_BASE} color="aurora.mint" borderColor="rgba(52,211,153,0.35)" bg="rgba(52,211,153,0.10)">
				<Box w="6px" h="6px" borderRadius="full" bg="aurora.mint" css={{ boxShadow: "0 0 10px #34D399" }} />
				active
			</Box>
		);
	}
	if (status === USER_STATUS.INACTIVE) {
		return (
			<Box {...PILL_BASE} color="fg.muted">
				<Box w="6px" h="6px" borderRadius="full" bg="fg.muted" />
				inactive
			</Box>
		);
	}
	return (
		<Box {...PILL_BASE} color="aurora.amber" borderColor="rgba(245,158,11,0.35)" bg="rgba(245,158,11,0.10)">
			<Box w="6px" h="6px" borderRadius="full" bg="aurora.amber" css={{ boxShadow: "0 0 10px #F59E0B" }} />
			pending
		</Box>
	);
};

/** Mint badge-check = verified · amber shield-question = login blocked until verified. */
const VerifiedPill = ({ verified }: { verified: boolean }) => {
	if (verified) {
		return (
			<Box {...PILL_BASE} color="aurora.mint" borderColor="rgba(52,211,153,0.35)" bg="rgba(52,211,153,0.10)">
				<LuBadgeCheck size={12} /> verified
			</Box>
		);
	}
	return (
		<Box
			{...PILL_BASE}
			color="aurora.amber"
			borderColor="rgba(245,158,11,0.35)"
			bg="rgba(245,158,11,0.10)"
			title="Login blocked until verified"
		>
			<LuShieldQuestion size={12} /> unverified
		</Box>
	);
};

const ROLES_PREVIEW_COUNT = 3;

/**
 * "Roles (by app)" cell. Renders one app_code:role_code chip per app-scoped role
 * the user holds (UserListItem.roles), overflowing to "+N" past
 * ROLES_PREVIEW_COUNT. Empty roles → "no roles". admin on isme = platform
 * administrator (no separate is_admin badge anymore).
 */
const RolesByAppCell = ({ roles }: { roles: AppRole[] }) => {
	if (roles.length === 0) {
		return (
			<Text fontSize="12px" color="fg.muted" fontStyle="italic">
				no roles
			</Text>
		);
	}
	const visible = roles.slice(0, ROLES_PREVIEW_COUNT);
	const hidden = roles.length - visible.length;
	return (
		<HStack gap="1.5" wrap="wrap" maxW="320px">
			{visible.map((role) => (
				<AppRoleChip
					key={`${role.app_code}:${role.role_code}`}
					appCode={role.app_code}
					role={role.role_code}
					title={`${role.app_name} · ${role.role_name}`}
				/>
			))}
			{hidden > 0 && (
				<Box
					{...PILL_BASE}
					px="9px"
					py="3px"
					color="fg.subtle"
					title={roles
						.slice(ROLES_PREVIEW_COUNT)
						.map((role) => `${role.app_code}:${role.role_code}`)
						.join(", ")}
				>
					+{hidden} more
				</Box>
			)}
		</HStack>
	);
};

/** "expired" is derived — pending rows past expires_at; never stored server-side. */
const invitationDisplayStatus = (invitation: InvitationListItem): InvitationDisplayStatus => {
	if (invitation.status === 1 && new Date(invitation.expires_at).getTime() < Date.now()) return "expired";
	return INVITATION_STATUS_LABELS[invitation.status];
};

const INVITATION_STATUS_FILTER_OPTIONS = ["all", "pending", "accepted", "revoked", "expired"] as const;

const InvitationStatusPill = ({ status }: { status: InvitationDisplayStatus }) => {
	if (status === "pending") {
		return (
			<Box {...PILL_BASE} color="aurora.amber" borderColor="rgba(245,158,11,0.35)" bg="rgba(245,158,11,0.10)">
				<Box w="6px" h="6px" borderRadius="full" bg="aurora.amber" css={{ boxShadow: "0 0 10px #F59E0B" }} />
				pending
			</Box>
		);
	}
	if (status === "accepted") {
		return (
			<Box {...PILL_BASE} color="aurora.mint" borderColor="rgba(52,211,153,0.35)" bg="rgba(52,211,153,0.10)">
				<Box w="6px" h="6px" borderRadius="full" bg="aurora.mint" css={{ boxShadow: "0 0 10px #34D399" }} />
				accepted
			</Box>
		);
	}
	if (status === "revoked") {
		return (
			<Box {...PILL_BASE} color="aurora.magenta" borderColor="rgba(236,72,153,0.35)" bg="rgba(236,72,153,0.10)">
				<Box w="6px" h="6px" borderRadius="full" bg="aurora.magenta" css={{ boxShadow: "0 0 10px #EC4899" }} />
				revoked
			</Box>
		);
	}
	return (
		<Box {...PILL_BASE} color="fg.muted">
			<Box w="6px" h="6px" borderRadius="full" bg="fg.muted" />
			expired
		</Box>
	);
};

/** Mock .num.warn cell — "in 71h · 2026-06-11" for live pending invites. */
const formatExpiry = (expiresAt: string) => {
	const remainingMs = new Date(expiresAt).getTime() - Date.now();
	const hours = Math.floor(remainingMs / (60 * 60 * 1000));
	const relative = hours >= 48 ? `in ${Math.floor(hours / 24)}d` : `in ${Math.max(hours, 1)}h`;
	return `${relative} · ${formatDateOnly(expiresAt)}`;
};

const emailInitials = (email: string) =>
	email
		.split("@")[0]
		.split(/[._-]/)
		.filter(Boolean)
		.map((part) => part[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

/** Dashed avatar = no account yet; gradient = accepted (account exists). */
const InvitationAvatar = ({ invitation }: { invitation: InvitationListItem }) => {
	if (invitation.status === 2) {
		return (
			<Center
				w="8"
				h="8"
				flex="none"
				borderRadius="full"
				fontSize="12px"
				fontWeight="bold"
				color="white"
				css={{ background: "linear-gradient(135deg, #22D3EE, #34D399)", boxShadow: "0 0 14px rgba(52,211,153,0.30)" }}
			>
				{emailInitials(invitation.email)}
			</Center>
		);
	}
	return (
		<Center
			w="8"
			h="8"
			flex="none"
			borderRadius="full"
			bg="bg.glass"
			borderWidth="1.5px"
			borderStyle="dashed"
			borderColor="border.strong"
			fontSize="12px"
			fontWeight="bold"
			color="fg.muted"
		>
			{emailInitials(invitation.email)}
		</Center>
	);
};

const UserAvatar = ({ user }: { user: UserListItem }) => {
	if (user.status === USER_STATUS.PENDING) {
		return (
			<Center
				w="8"
				h="8"
				flex="none"
				borderRadius="full"
				bg="bg.glass"
				borderWidth="1.5px"
				borderStyle="dashed"
				borderColor="border.strong"
				fontSize="12px"
				fontWeight="bold"
				color="fg.muted"
			>
				?
			</Center>
		);
	}
	const inactive = user.status === USER_STATUS.INACTIVE;
	return (
		<Center
			w="8"
			h="8"
			flex="none"
			borderRadius="full"
			fontSize="12px"
			fontWeight="bold"
			color="white"
			css={{
				background: inactive ? "linear-gradient(135deg, #3A3F6E, #23264A)" : avatarGradient(user.id),
				boxShadow: inactive ? "none" : "0 0 14px rgba(99,102,241,0.30)",
			}}
		>
			{initials(user.name)}
		</Center>
	);
};

const QuickActionButton = ({
	label,
	color,
	danger,
	onClick,
	disabled,
	children,
}: {
	label: string;
	color: string;
	danger?: boolean;
	onClick?: () => void;
	disabled?: boolean;
	children: React.ReactNode;
}) => (
	<Center
		as="button"
		w="8"
		h="8"
		borderRadius="9px"
		bg="rgba(7,7,26,0.55)"
		borderWidth="1px"
		borderColor="border.strong"
		color={color}
		cursor={disabled ? "not-allowed" : "pointer"}
		opacity={disabled ? 0.35 : 1}
		title={label}
		aria-label={label}
		css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
		_hover={
			disabled
				? undefined
				: danger
					? { bg: "rgba(236,72,153,0.15)", borderColor: "rgba(236,72,153,0.40)" }
					: { bg: "bg.glassHi", color: "fg", borderColor: "rgba(255,255,255,0.30)" }
		}
		onClick={(event) => {
			event.stopPropagation();
			if (!disabled) onClick?.();
		}}
	>
		{children}
	</Center>
);

const SessionsPanel = ({
	sessions,
	loading,
	onRevoke,
}: {
	sessions: UserSessionItem[];
	loading: boolean;
	onRevoke: (sessionId: string) => void;
}) => (
	<Box
		borderRadius="14px"
		borderWidth="1px"
		borderColor="border.strong"
		bg="bg.glass"
		overflow="hidden"
		css={{ backdropFilter: "blur(16px)", WebkitBackdropFilter: "blur(16px)" }}
	>
		<HStack px="3.5" py="2" gap="2" {...LABEL_PROPS} borderBottomWidth="1px" borderColor="border">
			<LuMonitorSmartphone size={13} /> Active sessions · {loading ? "…" : sessions.length}
		</HStack>
		{loading ? (
			<Center py="3">
				<Spinner size="sm" color="accent" />
			</Center>
		) : sessions.length === 0 ? (
			<Box px="3.5" py="2.5" fontSize="13px" color="fg.muted">
				No active sessions
			</Box>
		) : (
			sessions.map((session, index) => (
				<HStack
					key={session.id}
					gap="4.5"
					px="3.5"
					py="2.5"
					fontSize="13px"
					borderTopWidth={index > 0 ? "1px" : "0"}
					borderColor="border"
				>
					<HStack gap="2" color="fg" fontWeight="medium">
						<Box color="aurora.cyan">
							<LuMonitor size={15} />
						</Box>
						<Text>{parseUserAgent(session.user_agent)}</Text>
					</HStack>
					<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
						{session.client_ip}
					</Text>
					<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
						last login {formatShortDateTime(session.last_login_at)}
					</Text>
					<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
						expires {formatShortDateTime(session.expires_at)}
					</Text>
					<Button
						ml="auto"
						h="8"
						px="3"
						borderRadius="9px"
						variant="outline"
						borderColor="rgba(236,72,153,0.35)"
						bg="rgba(236,72,153,0.08)"
						fontSize="12px"
						fontWeight="semibold"
						color="aurora.magenta"
						_hover={{ bg: "rgba(236,72,153,0.16)" }}
						onClick={(event) => {
							event.stopPropagation();
							onRevoke(session.id);
						}}
					>
						<LuLogOut size={13} /> Revoke
					</Button>
				</HStack>
			))
		)}
	</Box>
);

export const Users = () => {
	const { user: currentUser } = useUser();
	const { can } = usePermissions();
	const navigate = useNavigate();

	const canAssignRole = can("role:assign") && can("role:read");
	const canVerify = can("user:verify");

	const [users, setUsers] = useState<UserListItem[]>([]);
	const [total, setTotal] = useState(0);
	const [loading, setLoading] = useState(true);
	// Bumped to force a server refetch after destructive mutations.
	const [reloadKey, setReloadKey] = useState(0);

	const [roleOptions, setRoleOptions] = useState<RoleListItem[]>([]);

	const [search, setSearch] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const searchRef = useRef<HTMLInputElement>(null);

	const [filterOpen, setFilterOpen] = useState(false);
	const [draftStatus, setDraftStatus] = useState<StatusFilter>("all");
	const [draftVerified, setDraftVerified] = useState<VerifiedFilter>("all");
	const [draftRole, setDraftRole] = useState("all");
	const [draftLastLogin, setDraftLastLogin] = useState<LastLoginFilter>("any");
	// `at` is captured when filters are applied so the last-login window is
	// computed against a stable timestamp (render must stay pure).
	const [applied, setApplied] = useState<{
		status: StatusFilter;
		verified: VerifiedFilter;
		role: string;
		lastLogin: LastLoginFilter;
		at: number;
	}>({ status: "all", verified: "all", role: "all", lastLogin: "any", at: 0 });

	const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
	const [expandedId, setExpandedId] = useState<string | null>(null);
	const [sessionsByUser, setSessionsByUser] = useState<Record<string, UserSessionItem[]>>({});
	const [sessionsLoadingFor, setSessionsLoadingFor] = useState<string | null>(null);

	const [page, setPage] = useState(1);
	const [pageSize, setPageSize] = useState(USERS_PAGE_SIZE_OPTIONS[0].value);

	const [inviteOpen, setInviteOpen] = useState(false);
	// Prefill for the "New invite" re-issue action on revoked/expired invitations
	// — carries the original email + app-scoped assignments.
	const [invitePrefill, setInvitePrefill] = useState<{
		email: string;
		assignments: InvitationRoleAssignment[];
	} | null>(null);

	// Users | Invitations sub-tab (mock frame 02).
	const [activeTab, setActiveTab] = useState<"users" | "invitations">("users");
	const [invitations, setInvitations] = useState<InvitationListItem[]>([]);
	const [invitationsLoading, setInvitationsLoading] = useState(true);
	const [invitationSearch, setInvitationSearch] = useState("");
	const [invitationStatusFilter, setInvitationStatusFilter] =
		useState<(typeof INVITATION_STATUS_FILTER_OPTIONS)[number]>("all");

	// One-way verify confirm (mock #verify-modal) — row action targets one
	// user, the bulk button targets every selected unverified user.
	const [verifyOpen, setVerifyOpen] = useState(false);
	const [verifyTargets, setVerifyTargets] = useState<UserListItem[]>([]);
	const [verifySkipped, setVerifySkipped] = useState(0);

	// Global "pending verification" count — honest data from the server-side
	// verified=false query's total (not derived from the loaded page).
	const [pendingVerification, setPendingVerification] = useState<number | null>(null);

	// Debounce search before hitting the server.
	useEffect(() => {
		const timer = setTimeout(() => {
			setDebouncedSearch(search.trim());
			setPage(1);
		}, 300);
		return () => clearTimeout(timer);
	}, [search]);

	// "pending" (status 3) is forward-looking — the backend neither stores nor
	// accepts it as a filter, so the result set is known-empty without a call.
	const pendingFilter = applied.status === "pending";

	// Server-side list: query/status/role/page/size are pushed to GET /api/v1/users.
	// Refetches keep the previous page visible (no spinner) — `loading` only
	// covers the initial load.
	useEffect(() => {
		if (applied.status === "pending") return;
		let active = true;
		listUsers({
			page,
			size: pageSize,
			query: debouncedSearch || undefined,
			status: applied.status !== "all" ? STATUS_FILTER_TO_VALUE[applied.status] : undefined,
			role: applied.role !== "all" ? applied.role : undefined,
			verified: applied.verified !== "all" ? applied.verified === "verified" : undefined,
		})
			.then((response) => {
				if (!active) return;
				setUsers(response.items);
				setTotal(response.total);
			})
			.catch(() => {
				if (active) toaster.create({ title: "Failed to load users", type: "error", meta: { closable: true } });
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, [page, pageSize, debouncedSearch, applied, reloadKey]);

	// Pending-verification subtitle count — a 1-row verified=false query just
	// for its total. Refetched alongside the list (reloadKey bumps).
	useEffect(() => {
		let active = true;
		listUsers({ page: 1, size: 1, verified: false })
			.then((response) => {
				if (active) setPendingVerification(response.total);
			})
			.catch(() => {
				// non-fatal — the subtitle simply omits the count
			});
		return () => {
			active = false;
		};
	}, [reloadKey]);

	// Invitations list — loaded alongside the users list so the tab badge and
	// subtitle pending count are always current. Refetched on reloadKey bumps.
	useEffect(() => {
		let active = true;
		listInvitations()
			.then((items) => {
				if (active) setInvitations(items);
			})
			.catch(() => {
				if (active) toaster.create({ title: "Failed to load invitations", type: "error", meta: { closable: true } });
			})
			.finally(() => {
				if (active) setInvitationsLoading(false);
			});
		return () => {
			active = false;
		};
	}, [reloadKey]);

	// Role options for the filter select + bulk assign (needs role:read).
	// Scoped to the isme app — the user list only carries isme-app roles, so the
	// role filter + bulk-assign operate on isme roles (platform roles).
	useEffect(() => {
		if (!can("role:read")) return;
		let active = true;
		listRoles({ app_code: "isme" })
			.then((items) => {
				if (active) setRoleOptions(items);
			})
			.catch(() => {
				// non-fatal — the role filter falls back to the static options
			});
		return () => {
			active = false;
		};
	}, [can]);

	useEffect(() => {
		const handler = (event: KeyboardEvent) => {
			if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
				event.preventDefault();
				searchRef.current?.focus();
			}
		};
		window.addEventListener("keydown", handler);
		return () => window.removeEventListener("keydown", handler);
	}, []);

	// Search/status/role/pagination are server-side now.
	// TODO(backend): last-login has no server param yet — it is applied
	// client-side to the fetched page only (counts may undershoot).
	const paged = useMemo(() => {
		if (pendingFilter) return [];
		const now = applied.at;
		return users.filter((user) => {
			if (applied.lastLogin === "never" && !isZeroTime(user.last_login_at)) return false;
			const windowMs = LAST_LOGIN_WINDOW_MS[applied.lastLogin];
			if (windowMs) {
				if (isZeroTime(user.last_login_at)) return false;
				if (now - new Date(user.last_login_at).getTime() > windowMs) return false;
			}
			return true;
		});
	}, [users, applied, pendingFilter]);

	const effectiveTotal = pendingFilter ? 0 : total;
	const totalPages = Math.max(1, Math.ceil(effectiveTotal / pageSize));
	const safePage = Math.min(page, totalPages);

	const appliedCount =
		(applied.status !== "all" ? 1 : 0) +
		(applied.verified !== "all" ? 1 : 0) +
		(applied.role !== "all" ? 1 : 0) +
		(applied.lastLogin !== "any" ? 1 : 0);
	const totalSessions = users.reduce((sum, user) => sum + user.sessions_count, 0);
	const allOnPageSelected = paged.length > 0 && paged.every((user) => selectedIds.has(user.id));
	const someOnPageSelected = paged.some((user) => selectedIds.has(user.id));

	const mutateUsers = (ids: string[], patch: Partial<UserListItem>) => {
		setUsers((previous) => previous.map((user) => (ids.includes(user.id) ? { ...user, ...patch } : user)));
	};

	const toggleSelect = (id: string) => {
		setSelectedIds((previous) => {
			const next = new Set(previous);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			return next;
		});
	};

	const toggleSelectAll = () => {
		setSelectedIds((previous) => {
			const next = new Set(previous);
			if (allOnPageSelected) paged.forEach((user) => next.delete(user.id));
			else paged.forEach((user) => next.add(user.id));
			return next;
		});
	};

	const handleRowClick = (user: UserListItem) => {
		if (expandedId === user.id) {
			setExpandedId(null);
			return;
		}
		setExpandedId(user.id);
		if (!sessionsByUser[user.id]) {
			setSessionsLoadingFor(user.id);
			listUserSessions(user.id)
				.then((sessions) => setSessionsByUser((previous) => ({ ...previous, [user.id]: sessions })))
				.finally(() => setSessionsLoadingFor(null));
		}
	};

	const handleToggleStatus = async (user: UserListItem) => {
		const nextStatus: UserStatus = user.status === USER_STATUS.ACTIVE ? USER_STATUS.INACTIVE : USER_STATUS.ACTIVE;
		await updateUserStatus(user.id, nextStatus);
		mutateUsers([user.id], { status: nextStatus });
		toaster.create({
			title: `${user.name || user.email} ${nextStatus === USER_STATUS.ACTIVE ? "activated" : "deactivated"}`,
			type: "success",
			meta: { closable: true },
		});
	};

	const handleRevokeInvitation = async (invitation: InvitationListItem) => {
		if (!window.confirm(`Revoke invite for ${invitation.email}? The link stops working immediately.`)) return;
		try {
			await revokeInvitation(invitation.id);
			setReloadKey((previous) => previous + 1);
			toaster.create({ title: `Invite for ${invitation.email} revoked`, type: "success", meta: { closable: true } });
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({ title: err?.response?.data?.message || "Failed to revoke invite", type: "error", meta: { closable: true } });
		}
	};

	// Re-issue: opens the invite dialog prefilled — accepting creates a NEW row.
	const handleReissueInvite = (invitation: InvitationListItem) => {
		setInvitePrefill({
			email: invitation.email,
			assignments: invitation.assignments.map((assignment) => ({
				role_id: assignment.role_id,
				app_service_id: assignment.app_service_id,
			})),
		});
		setInviteOpen(true);
	};

	const handleDelete = async (user: UserListItem) => {
		try {
			await softDeleteUser(user.id);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({ title: err?.response?.data?.message || "Failed to delete user", type: "error", meta: { closable: true } });
			return;
		}
		setSelectedIds((previous) => {
			const next = new Set(previous);
			next.delete(user.id);
			return next;
		});
		// Step back when the last row on this page was removed.
		if (users.length === 1 && page > 1) setPage(page - 1);
		setReloadKey((previous) => previous + 1);
		toaster.create({ title: `${user.name || user.email} deleted`, type: "success", meta: { closable: true } });
	};

	const handleEdit = () => {
		// TODO(backend): no update-user endpoint yet — edit screen comes with it.
		toaster.create({ title: "Edit user is not available yet", type: "info", meta: { closable: true } });
	};

	const openVerifyDialog = (targets: UserListItem[], skipped: number) => {
		setVerifyTargets(targets);
		setVerifySkipped(skipped);
		setVerifyOpen(true);
	};

	// Bulk verify applies ONLY to selected unverified rows — already-verified
	// selections are skipped (reported in the success toast).
	const handleBulkVerify = () => {
		const selectedUsers = users.filter((user) => selectedIds.has(user.id));
		const targets = selectedUsers.filter((user) => !user.is_verified);
		openVerifyDialog(targets, selectedUsers.length - targets.length);
	};

	const handleVerified = (ids: string[]) => {
		mutateUsers(ids, { is_verified: true });
		// refetch list + pending-verification count from the server
		setReloadKey((previous) => previous + 1);
	};

	const handleBulkStatus = async (status: UserStatus) => {
		const ids = [...selectedIds];
		await Promise.all(ids.map((id) => updateUserStatus(id, status)));
		mutateUsers(ids, { status });
		toaster.create({
			title: `${ids.length} user${ids.length > 1 ? "s" : ""} ${status === USER_STATUS.ACTIVE ? "activated" : "deactivated"}`,
			type: "success",
			meta: { closable: true },
		});
	};

	const handleBulkDelete = async () => {
		const ids = [...selectedIds];
		await Promise.all(ids.map((id) => softDeleteUser(id)));
		setSelectedIds(new Set());
		setPage(1);
		setReloadKey((previous) => previous + 1);
		toaster.create({ title: `${ids.length} user${ids.length > 1 ? "s" : ""} deleted`, type: "success", meta: { closable: true } });
	};

	const handleBulkAssignRole = async (roleCode: string) => {
		const ids = [...selectedIds];
		try {
			await assignUserRole(ids, roleCode);
			setReloadKey((previous) => previous + 1);
			toaster.create({
				title: `${roleCode} assigned to ${ids.length} user${ids.length > 1 ? "s" : ""}`,
				type: "success",
				meta: { closable: true },
			});
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } }; message?: string };
			toaster.create({
				title: err?.response?.data?.message || err?.message || "Failed to assign role",
				type: "error",
				meta: { closable: true },
			});
		}
	};

	const handleRevokeSession = async (userId: string, sessionId: string) => {
		await revokeUserSession(userId, sessionId);
		setSessionsByUser((previous) => ({
			...previous,
			[userId]: (previous[userId] ?? []).filter((session) => session.id !== sessionId),
		}));
		setUsers((previous) =>
			previous.map((user) =>
				user.id === userId ? { ...user, sessions_count: Math.max(0, user.sessions_count - 1) } : user
			)
		);
		toaster.create({ title: "Session revoked", type: "success", meta: { closable: true } });
	};

	const handleApplyFilters = () => {
		setApplied({ status: draftStatus, verified: draftVerified, role: draftRole, lastLogin: draftLastLogin, at: Date.now() });
		setPage(1);
	};

	const handleResetFilters = () => {
		setDraftStatus("all");
		setDraftVerified("all");
		setDraftRole("all");
		setDraftLastLogin("any");
		setApplied({ status: "all", verified: "all", role: "all", lastLogin: "any", at: 0 });
		setFilterOpen(false);
		setPage(1);
	};

	const handleExportCsv = () => {
		const header = "id,name,email,status,is_verified,roles,sessions,last_login_at,created_at";
		const rows = paged.map((user) =>
			[
				user.id,
				user.name,
				user.email,
				user.status,
				user.is_verified,
				(user.roles ?? []).map((role) => `${role.app_code}:${role.role_code}`).join(" "),
				user.sessions_count,
				user.last_login_at,
				user.created_at,
			]
				.map((value) => `"${String(value).replace(/"/g, '""')}"`)
				.join(",")
		);
		const blob = new Blob([[header, ...rows].join("\n")], { type: "text/csv" });
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement("a");
		anchor.href = url;
		anchor.download = "users.csv";
		anchor.click();
		URL.revokeObjectURL(url);
	};

	const filterButtonActive = filterOpen || appliedCount > 0;

	// Invitations are a small set — filtered client-side (search + display status).
	const pendingInvitationsCount = invitations.filter(
		(invitation) => invitationDisplayStatus(invitation) === "pending"
	).length;
	const visibleInvitations = invitations.filter((invitation) => {
		if (invitationSearch && !invitation.email.toLowerCase().includes(invitationSearch.trim().toLowerCase())) {
			return false;
		}
		if (invitationStatusFilter !== "all" && invitationDisplayStatus(invitation) !== invitationStatusFilter) {
			return false;
		}
		return true;
	});

	return (
		<AppShell active="users" user={{ name: currentUser?.name || "User", email: currentUser?.email || "" }}>
			{/* Page head */}
			<Flex align="flex-end" justify="space-between" gap="4" wrap="wrap">
				<Box>
					<Heading fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1" color="fg">
						Users
					</Heading>
					<Text fontSize="sm" color="fg.muted" mt="1.5">
						<Text as="span" color="aurora.cyan" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
							{effectiveTotal.toLocaleString()}
						</Text>{" "}
						accounts ·{" "}
						<Text as="span" color="aurora.mint" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
							{totalSessions.toLocaleString()}
						</Text>{" "}
						active sessions on this page
						{pendingVerification !== null && (
							<>
								{" "}
								·{" "}
								<Text as="span" color="aurora.amber" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
									{pendingVerification.toLocaleString()}
								</Text>{" "}
								pending verification
							</>
						)}
						{pendingInvitationsCount > 0 && (
							<>
								{" "}
								·{" "}
								<Text as="span" color="aurora.amber" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
									{pendingInvitationsCount.toLocaleString()}
								</Text>{" "}
								pending invitation{pendingInvitationsCount > 1 ? "s" : ""}
							</>
						)}
					</Text>
				</Box>
			</Flex>

			{/* Sub-tabs (left) + page actions (right) on one row */}
			<Flex align="center" justify="space-between" gap="4" wrap="wrap">
				{/* Sub-tabs: Users (existing table) | Invitations (mock frame 02) */}
				<HStack
					gap="1"
					p="1"
					borderWidth="1px"
					borderColor="border.strong"
					borderRadius="glassSm"
					bg="bg.glass"
					w="fit-content"
					role="tablist"
					aria-label="Users page sections"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
				>
					{(
						[
							{ key: "users", label: "Users", icon: <LuUsers size={14} /> },
							{ key: "invitations", label: "Invitations", icon: <LuLink size={14} /> },
						] as const
					).map((tab) => {
						const on = activeTab === tab.key;
						return (
							<HStack
								key={tab.key}
								as="button"
								role="tab"
								aria-selected={on}
								h="9"
								px="4"
								gap="2"
								borderRadius="9px"
								cursor="pointer"
								fontSize="13px"
								fontWeight="semibold"
								color={on ? "fg" : "fg.muted"}
								bg={on ? "linear-gradient(135deg, rgba(99,102,241,0.30), rgba(139,92,246,0.25))" : "transparent"}
								boxShadow={on ? "0 0 16px rgba(139,92,246,0.25) inset" : "none"}
								_hover={{ color: "fg" }}
								onClick={() => setActiveTab(tab.key)}
							>
								{tab.icon}
								{tab.label}
								{tab.key === "invitations" && pendingInvitationsCount > 0 && (
									<Center minW="18px" h="18px" px="5px" borderRadius="full" fontSize="11px" bg="rgba(245,158,11,0.22)" color="aurora.amber">
										{pendingInvitationsCount}
									</Center>
								)}
							</HStack>
						);
					})}
				</HStack>

				<HStack gap="2.5">
					<Button h="11" px="4.5" fontSize="sm" {...GHOST_BUTTON_PROPS} onClick={handleExportCsv} disabled>
						<LuDownload size={16} /> Export CSV
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
						_focusVisible={{ boxShadow: "focusRing" }}
						onClick={() => navigate("/users/invite")}
					>
						<LuUserPlus size={16} /> Invite user
					</Button>
				</HStack>
			</Flex>

			{/* Users glass panel */}
			{activeTab === "users" && (
			<Box
				bg="bg.glass"
				borderWidth="1px"
				borderColor="border"
				borderRadius="20px"
				boxShadow="glassSoft"
				overflow="hidden"
				css={{ backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" }}
			>
				{/* Toolbar */}
				<Flex align="center" gap="2.5" p="3.5" wrap="wrap" borderBottomWidth="1px" borderColor="border">
					<Box
						flex="1"
						minW="240px"
						position="relative"
						css={{ "&:focus-within .search-icon": { color: "#22D3EE" } }}
					>
						<Box className="search-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none">
							<LuSearch size={16} />
						</Box>
						<Input
							ref={searchRef}
							h="11"
							pl="10"
							pr="16"
							borderRadius="glassSm"
							bg="bg.glass"
							borderColor="border.strong"
							fontSize="sm"
							color="fg"
							placeholder="Search by name or email…"
							css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
							_placeholder={{ color: "fg.muted" }}
							_hover={{ borderColor: "rgba(255,255,255,0.28)" }}
							_focus={{ borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" }}
							value={search}
							onChange={(event) => {
								setSearch(event.target.value);
								setPage(1);
							}}
						/>
						<Box
							position="absolute"
							right="3.5"
							top="2.5"
							fontSize="11px"
							px="7px"
							py="2px"
							borderRadius="6px"
							borderWidth="1px"
							borderColor="border.strong"
							bg="bg.glass"
							color="fg.muted"
						>
							⌘K
						</Box>
					</Box>
					<Button
						h="11"
						px="3.5"
						variant="outline"
						borderRadius="glassSm"
						fontSize="13px"
						fontWeight="semibold"
						borderColor={filterButtonActive ? "rgba(139,92,246,0.45)" : "border.strong"}
						bg={filterButtonActive ? "rgba(139,92,246,0.14)" : "bg.glass"}
						color={filterButtonActive ? "aurora.violet" : "fg.subtle"}
						boxShadow={filterButtonActive ? "0 0 18px rgba(139,92,246,0.20)" : "none"}
						css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
						_hover={{ bg: filterButtonActive ? "rgba(139,92,246,0.20)" : "bg.glassHi" }}
						onClick={() => setFilterOpen((previous) => !previous)}
					>
						<LuFilter size={14} /> Filter
						{appliedCount > 0 && (
							<Center minW="18px" h="18px" px="5px" borderRadius="full" fontSize="11px" bg="rgba(139,92,246,0.25)" color="white">
								{appliedCount}
							</Center>
						)}
					</Button>
				</Flex>

				{/* Filter row */}
				{filterOpen && (
					<Flex align="flex-end" gap="4" px="3.5" py="4" wrap="wrap" borderBottomWidth="1px" borderColor="border" bg="rgba(7,7,26,0.35)">
						<Box>
							<Text {...LABEL_PROPS} mb="2">
								Status
							</Text>
							<HStack
								gap="0"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border.strong"
								overflow="hidden"
								w="fit-content"
								bg="bg.glass"
								css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
							>
								{USER_STATUS_FILTER_OPTIONS.map((option, index) => (
									<HStack
										key={option}
										as="button"
										h="10"
										px="3.5"
										gap="1.5"
										fontSize="13px"
										fontWeight="medium"
										cursor="pointer"
										bg={draftStatus === option ? "linear-gradient(135deg, rgba(99,102,241,0.30), rgba(139,92,246,0.25))" : "transparent"}
										boxShadow={draftStatus === option ? "0 0 16px rgba(139,92,246,0.25) inset" : "none"}
										color={draftStatus === option ? "fg" : "fg.muted"}
										borderRightWidth={index < USER_STATUS_FILTER_OPTIONS.length - 1 ? "1px" : "0"}
										borderColor="border"
										_hover={{ color: "fg" }}
										onClick={() => setDraftStatus(option)}
									>
										{STATUS_FILTER_ICONS[option]}
										{option}
									</HStack>
								))}
							</HStack>
						</Box>
						<Box>
							<Text {...LABEL_PROPS} mb="2">
								Verified
							</Text>
							{/* maps to user.is_verified — unverified accounts are blocked at
							    login until an admin / user:verify holder verifies them */}
							<HStack
								gap="0"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border.strong"
								overflow="hidden"
								w="fit-content"
								bg="bg.glass"
								css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
							>
								{USER_VERIFIED_FILTER_OPTIONS.map((option, index) => (
									<HStack
										key={option}
										as="button"
										h="10"
										px="3.5"
										gap="1.5"
										fontSize="13px"
										fontWeight="medium"
										cursor="pointer"
										bg={draftVerified === option ? "linear-gradient(135deg, rgba(99,102,241,0.30), rgba(139,92,246,0.25))" : "transparent"}
										boxShadow={draftVerified === option ? "0 0 16px rgba(139,92,246,0.25) inset" : "none"}
										color={draftVerified === option ? "fg" : "fg.muted"}
										borderRightWidth={index < USER_VERIFIED_FILTER_OPTIONS.length - 1 ? "1px" : "0"}
										borderColor="border"
										_hover={{ color: "fg" }}
										onClick={() => setDraftVerified(option)}
									>
										{VERIFIED_FILTER_ICONS[option]}
										{option}
									</HStack>
								))}
							</HStack>
						</Box>
						<Box>
							<Text {...LABEL_PROPS} mb="2">
								Role
							</Text>
							<NativeSelect.Root size="sm" w="40">
								<NativeSelect.Field
									h="10"
									borderRadius="glassSm"
									bg="bg.glass"
									borderColor="border.strong"
									fontSize="13px"
									color="fg"
									css={{
										backdropFilter: "blur(12px)",
										WebkitBackdropFilter: "blur(12px)",
										"& option": { background: "#12122E", color: "#F4F5FF" },
									}}
									_focus={{ borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none" }}
									value={draftRole}
									onChange={(event) => setDraftRole(event.target.value)}
								>
									<option value="all">all</option>
									{(roleOptions.length > 0 ? roleOptions.map((role) => role.code) : USER_ROLE_OPTIONS).map((role) => (
										<option key={role} value={role}>
											{role}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Box>
						<Box>
							<Text {...LABEL_PROPS} mb="2">
								Last login
							</Text>
							<NativeSelect.Root size="sm" w="40">
								<NativeSelect.Field
									h="10"
									borderRadius="glassSm"
									bg="bg.glass"
									borderColor="border.strong"
									fontSize="13px"
									color="fg"
									css={{
										backdropFilter: "blur(12px)",
										WebkitBackdropFilter: "blur(12px)",
										"& option": { background: "#12122E", color: "#F4F5FF" },
									}}
									_focus={{ borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none" }}
									value={draftLastLogin}
									onChange={(event) => setDraftLastLogin(event.target.value as LastLoginFilter)}
								>
									{USER_LAST_LOGIN_FILTER_OPTIONS.map((option) => (
										<option key={option.value} value={option.value}>
											{option.label}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Box>
						<HStack gap="2.5" ml="auto">
							<Button {...GHOST_BUTTON_PROPS} h="34px" px="3" fontSize="13px" borderRadius="10px" onClick={handleResetFilters}>
								Reset
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
								onClick={handleApplyFilters}
							>
								Apply
							</Button>
						</HStack>
						<HStack w="full" gap="2" fontSize="12px" color="fg.muted">
							<LuInfo size={13} />
							<Text>
								status ·{" "}
								<Text as="span" color="fg.subtle">
									user.status
								</Text>{" "}
								&nbsp;·&nbsp; verified ·{" "}
								<Text as="span" color="fg.subtle">
									user.is_verified
								</Text>{" "}
								&nbsp;·&nbsp; role · rbac assignment &nbsp;·&nbsp; last login ·{" "}
								<Text as="span" color="fg.subtle">
									last_login_at
								</Text>
							</Text>
						</HStack>
					</Flex>
				)}

				{/* Table */}
				{loading ? (
					<Center py="16">
						<Spinner size="lg" color="accent" />
					</Center>
				) : (
					<Table.Root size="sm" css={{ "& td, & th": { borderColor: "rgba(255,255,255,0.10)" } }}>
						<Table.Header>
							<Table.Row bg="transparent">
								<Table.ColumnHeader w="44px" px="3.5" py="3" bg="transparent">
									<Checkbox
										size="sm"
										colorPalette="purple"
										css={CHECKBOX_CSS}
										checked={allOnPageSelected ? true : someOnPageSelected ? "indeterminate" : false}
										onCheckedChange={toggleSelectAll}
									/>
								</Table.ColumnHeader>
								{["User", "Status", "Verified", "Roles (by app)", "Sessions", "Last login", "Created"].map((label) => (
									<Table.ColumnHeader key={label} px="3.5" py="3" bg="transparent" {...LABEL_PROPS}>
										{label}
									</Table.ColumnHeader>
								))}
								<Table.ColumnHeader w="170px" px="3.5" py="3" bg="transparent" />
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{paged.length === 0 && (
								<Table.Row bg="transparent">
									<Table.Cell colSpan={9} px="3.5" py="8" textAlign="center" fontSize="sm" color="fg.muted">
										No users match the current search / filters.
									</Table.Cell>
								</Table.Row>
							)}
							{paged.map((user) => {
								const selected = selectedIds.has(user.id);
								const expanded = expandedId === user.id;
								const pending = user.status === USER_STATUS.PENDING;
								const inactive = user.status === USER_STATUS.INACTIVE;
								return (
									<Fragment key={user.id}>
										<Table.Row
											cursor="pointer"
											transition="background .15s ease-out"
											opacity={inactive ? 0.65 : 1}
											bg={selected ? "rgba(139,92,246,0.10)" : expanded ? "rgba(255,255,255,0.03)" : "transparent"}
											_hover={{ bg: selected ? "rgba(139,92,246,0.10)" : "rgba(255,255,255,0.04)" }}
											css={{
												"& .row-actions": { opacity: selected || expanded ? 1 : 0, transition: "opacity .15s ease-out" },
												"&:hover .row-actions": { opacity: 1 },
												"& td": expanded ? { borderBottomWidth: "0px" } : undefined,
											}}
											onClick={() => handleRowClick(user)}
										>
											<Table.Cell px="3.5" py="13px" onClick={(event) => event.stopPropagation()}>
												<Checkbox
													size="sm"
													colorPalette="purple"
													css={CHECKBOX_CSS}
													checked={selected}
													onCheckedChange={() => toggleSelect(user.id)}
												/>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<HStack gap="3" minW="0" maxW="300px">
													<UserAvatar user={user} />
													<Box minW="0" lineHeight="1.25">
														<Text
															fontSize="sm"
															fontWeight="semibold"
															truncate
															color={pending ? "fg.muted" : "fg"}
															title={user.name || undefined}
														>
															{user.name || "—"}
														</Text>
														<Text fontSize="12px" color="fg.muted" truncate>
															{user.email}
														</Text>
													</Box>
												</HStack>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<StatusPill status={user.status} />
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<VerifiedPill verified={user.is_verified} />
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<RolesByAppCell roles={user.roles ?? []} />
											</Table.Cell>
											<Table.Cell
												px="3.5"
												py="13px"
												{...NUM_PROPS}
												fontWeight={user.sessions_count > 0 ? "semibold" : "normal"}
												color={user.sessions_count > 0 ? "aurora.cyan" : "fg.muted"}
											>
												{user.sessions_count}
											</Table.Cell>
											<Table.Cell px="3.5" py="13px" {...NUM_PROPS} color={isZeroTime(user.last_login_at) ? "fg.muted" : "fg.subtle"}>
												{formatShortDateTime(user.last_login_at)}
											</Table.Cell>
											<Table.Cell px="3.5" py="13px" {...NUM_PROPS} color="fg.muted">
												{formatDateOnly(user.created_at)}
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<HStack className="row-actions" gap="1.5" justify="flex-end">
													{/* verify (one-way) — rendered ONLY when is_verified=false AND the
													    caller holds user:verify; verified rows never show it */}
													{!user.is_verified && canVerify && (
														<QuickActionButton label="Verify account" color="aurora.mint" onClick={() => openVerifyDialog([user], 0)}>
															<LuBadgeCheck size={14} />
														</QuickActionButton>
													)}
													<QuickActionButton label="Edit" color="fg.subtle" onClick={handleEdit}>
														<LuPencil size={14} />
													</QuickActionButton>
													{user.status === USER_STATUS.ACTIVE ? (
														<QuickActionButton label="Deactivate" color="aurora.amber" onClick={() => handleToggleStatus(user)}>
															<LuCircleSlash size={14} />
														</QuickActionButton>
													) : (
														<QuickActionButton label="Activate" color="aurora.mint" onClick={() => handleToggleStatus(user)}>
															<LuCircleCheck size={14} />
														</QuickActionButton>
													)}
													{/* TODO(backend): reset-password deferred — needs email infra to send the link. */}
													<QuickActionButton label="Reset password — available when email lands" color="fg.subtle" disabled>
														<LuKeyRound size={14} />
													</QuickActionButton>
													<QuickActionButton label="Delete" color="aurora.magenta" danger onClick={() => handleDelete(user)}>
														<LuTrash2 size={14} />
													</QuickActionButton>
												</HStack>
											</Table.Cell>
										</Table.Row>
										{expanded && (
											<Table.Row bg="rgba(7,7,26,0.30)">
												<Table.Cell px="0" py="0" border="none" />
												<Table.Cell colSpan={8} px="3.5" pt="0" pb="4">
													<SessionsPanel
														sessions={sessionsByUser[user.id] ?? []}
														loading={sessionsLoadingFor === user.id}
														onRevoke={(sessionId) => handleRevokeSession(user.id, sessionId)}
													/>
												</Table.Cell>
											</Table.Row>
										)}
									</Fragment>
								);
							})}
						</Table.Body>
					</Table.Root>
				)}

				{/* Bulk-select bar */}
				{selectedIds.size > 0 && (
					<HStack
						gap="3"
						px="4"
						py="3"
						borderTopWidth="1px"
						borderColor="rgba(139,92,246,0.30)"
						bg="linear-gradient(90deg, rgba(139,92,246,0.16), rgba(99,102,241,0.08))"
						fontSize="13px"
					>
						<Text color="aurora.violet" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
							{selectedIds.size} selected
						</Text>
						<Box h="18px" w="1px" bg="border.strong" />
						{/* bulk verify — only selected unverified rows are sent; verified
						    selections are skipped (reported in the toast). Hidden without user:verify. */}
						{canVerify && (
							<Button
								h="8"
								px="3"
								variant="outline"
								borderRadius="9px"
								borderColor="rgba(52,211,153,0.35)"
								bg="rgba(52,211,153,0.08)"
								fontSize="12px"
								fontWeight="semibold"
								color="aurora.mint"
								_hover={{ bg: "rgba(52,211,153,0.16)" }}
								disabled={!users.some((user) => selectedIds.has(user.id) && !user.is_verified)}
								title={
									users.some((user) => selectedIds.has(user.id) && !user.is_verified)
										? undefined
										: "All selected accounts are already verified"
								}
								onClick={handleBulkVerify}
							>
								<LuBadgeCheck size={13} /> Verify
							</Button>
						)}
						<Button {...GHOST_BUTTON_PROPS} h="8" px="3" fontSize="12px" borderRadius="9px" color="fg.subtle" onClick={() => handleBulkStatus(USER_STATUS.ACTIVE)}>
							<LuCircleCheck size={13} /> Activate
						</Button>
						<Button {...GHOST_BUTTON_PROPS} h="8" px="3" fontSize="12px" borderRadius="9px" color="fg.subtle" onClick={() => handleBulkStatus(USER_STATUS.INACTIVE)}>
							<LuCircleSlash size={13} /> Deactivate
						</Button>
						<Menu.Root onSelect={(details) => handleBulkAssignRole(details.value)}>
							<Menu.Trigger asChild>
								<Button
									{...GHOST_BUTTON_PROPS}
									h="8"
									px="3"
									fontSize="12px"
									borderRadius="9px"
									color="fg.subtle"
									disabled={!canAssignRole || roleOptions.length === 0}
									title={canAssignRole ? undefined : "Requires the role:assign permission"}
								>
									<LuKeyRound size={13} /> Assign role
								</Button>
							</Menu.Trigger>
							<Portal>
								<Menu.Positioner>
									<Menu.Content
										bg="linear-gradient(180deg, rgba(18,18,46,0.96), rgba(11,11,35,0.97))"
										borderWidth="1px"
										borderColor="border.strong"
										borderRadius="14px"
										boxShadow="glassPop"
										minW="40"
									>
										{roleOptions.map((role) => (
											<Menu.Item
												key={role.id}
												value={role.code}
												fontSize="13px"
												color="fg.subtle"
												_hover={{ bg: "rgba(139,92,246,0.14)", color: "fg" }}
												_highlighted={{ bg: "rgba(139,92,246,0.14)", color: "fg" }}
											>
												{role.code}
											</Menu.Item>
										))}
									</Menu.Content>
								</Menu.Positioner>
							</Portal>
						</Menu.Root>
						<Button
							h="8"
							px="3"
							variant="outline"
							borderRadius="9px"
							borderColor="rgba(236,72,153,0.35)"
							bg="rgba(236,72,153,0.08)"
							fontSize="12px"
							fontWeight="semibold"
							color="aurora.magenta"
							_hover={{ bg: "rgba(236,72,153,0.16)" }}
							onClick={handleBulkDelete}
						>
							<LuTrash2 size={13} /> Delete
						</Button>
						<Center
							as="button"
							ml="auto"
							w="8"
							h="8"
							borderRadius="9px"
							bg="rgba(7,7,26,0.55)"
							borderWidth="1px"
							borderColor="border.strong"
							color="fg.subtle"
							cursor="pointer"
							_hover={{ bg: "bg.glassHi", color: "fg" }}
							aria-label="Clear selection"
							onClick={() => setSelectedIds(new Set())}
						>
							<LuX size={14} />
						</Center>
					</HStack>
				)}

				{/* Pagination */}
				<Flex
					align="center"
					justify="space-between"
					px="4"
					py="3"
					borderTopWidth="1px"
					borderColor="border"
					fontSize="12px"
					color="fg.muted"
					css={{ fontVariantNumeric: "tabular-nums" }}
				>
					<Text>
						Page {safePage} / {totalPages} · {paged.length} of {effectiveTotal.toLocaleString()} shown
					</Text>
					<HStack gap="2">
						<Text>Rows</Text>
						<NativeSelect.Root size="xs" w="16">
							<NativeSelect.Field
								h="8"
								borderRadius="9px"
								bg="bg.glass"
								borderColor="border.strong"
								fontSize="12px"
								color="fg"
								css={{ "& option": { background: "#12122E", color: "#F4F5FF" } }}
								value={String(pageSize)}
								onChange={(event) => {
									setPageSize(Number(event.target.value));
									setPage(1);
								}}
							>
								{USERS_PAGE_SIZE_OPTIONS.map((option) => (
									<option key={option.value} value={option.value}>
										{option.label}
									</option>
								))}
							</NativeSelect.Field>
							<NativeSelect.Indicator color="fg.muted" />
						</NativeSelect.Root>
						<Center
							as="button"
							h="8"
							w="8"
							borderRadius="9px"
							borderWidth="1px"
							borderColor="border.strong"
							bg="bg.glass"
							color="fg.subtle"
							cursor={safePage > 1 ? "pointer" : "not-allowed"}
							opacity={safePage > 1 ? 1 : 0.35}
							_hover={safePage > 1 ? { bg: "bg.glassHi", color: "fg" } : undefined}
							aria-label="Previous page"
							onClick={() => setPage((previous) => Math.max(1, previous - 1))}
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
							cursor={safePage < totalPages ? "pointer" : "not-allowed"}
							opacity={safePage < totalPages ? 1 : 0.35}
							_hover={safePage < totalPages ? { bg: "bg.glassHi", color: "fg" } : undefined}
							aria-label="Next page"
							onClick={() => setPage((previous) => Math.min(totalPages, previous + 1))}
						>
							<LuChevronRight size={14} />
						</Center>
					</HStack>
				</Flex>
			</Box>
			)}

			{/* Invitations glass panel (mock frame 02) */}
			{activeTab === "invitations" && (
			<Box
				bg="bg.glass"
				borderWidth="1px"
				borderColor="border"
				borderRadius="20px"
				boxShadow="glassSoft"
				overflow="hidden"
				css={{ backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" }}
			>
				{invitationsLoading ? (
					<Center py="16">
						<Spinner size="lg" color="accent" />
					</Center>
				) : invitations.length === 0 ? (
					<Center flexDirection="column" gap="1" py="14" px="6" textAlign="center">
						<Center
							w="52px"
							h="52px"
							borderRadius="16px"
							mb="2.5"
							bg="linear-gradient(135deg, rgba(99,102,241,0.20), rgba(236,72,153,0.15))"
							borderWidth="1px"
							borderColor="border.strong"
							color="aurora.cyan"
						>
							<LuLink size={22} />
						</Center>
						<Text fontSize="md" fontWeight="semibold" color="fg">
							No invitations yet
						</Text>
						<Text fontSize="13px" color="fg.muted" maxW="360px" mb="3.5">
							Invite links are the only way to create accounts — public signup is disabled. Create one to get a teammate on board.
						</Text>
						<Button
							h="10"
							px="4"
							borderRadius="glassSm"
							fontSize="13px"
							fontWeight="semibold"
							color="white"
							css={AURORA_CTA_STYLE}
							boxShadow="ctaGlow"
							_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
							onClick={() => setInviteOpen(true)}
						>
							<LuUserPlus size={15} /> Invite user
						</Button>
					</Center>
				) : (
					<>
						{/* Toolbar: email search + display-status filter (client-side) */}
						<Flex align="center" gap="2.5" p="3.5" wrap="wrap" borderBottomWidth="1px" borderColor="border">
							<Box flex="1" minW="240px" position="relative" css={{ "&:focus-within .search-icon": { color: "#22D3EE" } }}>
								<Box className="search-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none">
									<LuSearch size={16} />
								</Box>
								<Input
									h="11"
									pl="10"
									borderRadius="glassSm"
									bg="bg.glass"
									borderColor="border.strong"
									fontSize="sm"
									color="fg"
									placeholder="Search by email…"
									css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
									_placeholder={{ color: "fg.muted" }}
									_hover={{ borderColor: "rgba(255,255,255,0.28)" }}
									_focus={{ borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" }}
									value={invitationSearch}
									onChange={(event) => setInvitationSearch(event.target.value)}
								/>
							</Box>
							<NativeSelect.Root size="sm" w="40">
								<NativeSelect.Field
									h="11"
									borderRadius="glassSm"
									bg="bg.glass"
									borderColor="border.strong"
									fontSize="13px"
									color="fg"
									aria-label="Status filter"
									css={{
										backdropFilter: "blur(12px)",
										WebkitBackdropFilter: "blur(12px)",
										"& option": { background: "#12122E", color: "#F4F5FF" },
									}}
									_focus={{ borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none" }}
									value={invitationStatusFilter}
									onChange={(event) =>
										setInvitationStatusFilter(event.target.value as (typeof INVITATION_STATUS_FILTER_OPTIONS)[number])
									}
								>
									{INVITATION_STATUS_FILTER_OPTIONS.map((option) => (
										<option key={option} value={option}>
											{option === "all" ? "all statuses" : option}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Flex>

						<Table.Root size="sm" css={{ "& td, & th": { borderColor: "rgba(255,255,255,0.10)" } }}>
							<Table.Header>
								<Table.Row bg="transparent">
									{["Email", "Role", "Status", "Expires", "Created"].map((label) => (
										<Table.ColumnHeader key={label} px="3.5" py="3" bg="transparent" {...LABEL_PROPS}>
											{label}
										</Table.ColumnHeader>
									))}
									<Table.ColumnHeader w="120px" px="3.5" py="3" bg="transparent" />
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{visibleInvitations.length === 0 && (
									<Table.Row bg="transparent">
										<Table.Cell colSpan={6} px="3.5" py="8" textAlign="center" fontSize="sm" color="fg.muted">
											No invitations match the current search / filters.
										</Table.Cell>
									</Table.Row>
								)}
								{visibleInvitations.map((invitation) => {
									const displayStatus = invitationDisplayStatus(invitation);
									const muted = displayStatus === "revoked" || displayStatus === "expired";
									return (
										<Table.Row
											key={invitation.id}
											transition="background .15s ease-out"
											bg="transparent"
											_hover={{ bg: "rgba(255,255,255,0.04)" }}
											css={{
												"& > td": muted ? { opacity: 0.6 } : undefined,
												"& > td:last-of-type": { opacity: 1 },
												"& .row-actions": { opacity: 0, transition: "opacity .15s ease-out" },
												"&:hover .row-actions": { opacity: 1 },
											}}
										>
											<Table.Cell px="3.5" py="13px">
												<HStack gap="3" minW="0" maxW="320px">
													<InvitationAvatar invitation={invitation} />
													<Box minW="0" lineHeight="1.25">
														<Text fontSize="sm" fontWeight="semibold" truncate color="fg" title={invitation.email}>
															{invitation.email}
														</Text>
														{invitation.accepted_at && (
															<Text fontSize="12px" color="fg.muted" truncate>
																accepted {formatDateOnly(invitation.accepted_at)}
															</Text>
														)}
													</Box>
												</HStack>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												{/* Invitations carry one or more app-scoped role assignments —
												    one app_code:role_code chip each, overflowing past 3. */}
												{invitation.assignments.length === 0 ? (
													<Text fontSize="12px" color="fg.muted" fontStyle="italic">
														no roles
													</Text>
												) : (
													<HStack gap="1.5" wrap="wrap" maxW="280px">
														{invitation.assignments.slice(0, ROLES_PREVIEW_COUNT).map((assignment) => (
															<AppRoleChip
																key={`${assignment.app_code}:${assignment.role_code}`}
																appCode={assignment.app_code}
																role={assignment.role_code}
																title={`${assignment.app_name} · ${assignment.role_name}`}
															/>
														))}
														{invitation.assignments.length > ROLES_PREVIEW_COUNT && (
															<Box
																{...PILL_BASE}
																px="9px"
																py="3px"
																color="fg.subtle"
																title={invitation.assignments
																	.slice(ROLES_PREVIEW_COUNT)
																	.map((assignment) => `${assignment.app_code}:${assignment.role_code}`)
																	.join(", ")}
															>
																+{invitation.assignments.length - ROLES_PREVIEW_COUNT} more
															</Box>
														)}
													</HStack>
												)}
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<InvitationStatusPill status={displayStatus} />
											</Table.Cell>
											<Table.Cell px="3.5" py="13px" {...NUM_PROPS} color={displayStatus === "pending" ? "aurora.amber" : "fg.muted"}>
												{displayStatus === "pending"
													? formatExpiry(invitation.expires_at)
													: displayStatus === "expired"
														? formatDateOnly(invitation.expires_at)
														: "—"}
											</Table.Cell>
											<Table.Cell px="3.5" py="13px" {...NUM_PROPS} color="fg.muted">
												{formatDateOnly(invitation.created_at)}
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<HStack className="row-actions" gap="1.5" justify="flex-end">
													{displayStatus === "pending" && (
														<QuickActionButton label="Revoke invite" color="aurora.magenta" danger onClick={() => handleRevokeInvitation(invitation)}>
															<LuBan size={14} />
														</QuickActionButton>
													)}
													{muted && (
														<QuickActionButton label="New invite" color="aurora.cyan" onClick={() => handleReissueInvite(invitation)}>
															<LuRefreshCw size={14} />
														</QuickActionButton>
													)}
												</HStack>
											</Table.Cell>
										</Table.Row>
									);
								})}
							</Table.Body>
						</Table.Root>

						<Flex
							align="center"
							justify="space-between"
							px="4"
							py="3"
							borderTopWidth="1px"
							borderColor="border"
							fontSize="12px"
							color="fg.muted"
							css={{ fontVariantNumeric: "tabular-nums" }}
						>
							<Text>
								{visibleInvitations.length} of {invitations.length} invitation{invitations.length === 1 ? "" : "s"} shown
							</Text>
						</Flex>
					</>
				)}
			</Box>
			)}

			<InviteUserDialog
				open={inviteOpen}
				onOpenChange={(open) => {
					setInviteOpen(open);
					if (!open) setInvitePrefill(null);
				}}
				onInvited={() => setReloadKey((previous) => previous + 1)}
				prefill={invitePrefill}
			/>

			<VerifyUserDialog
				open={verifyOpen}
				targets={verifyTargets}
				skippedCount={verifySkipped}
				onOpenChange={setVerifyOpen}
				onVerified={handleVerified}
			/>
		</AppShell>
	);
};
