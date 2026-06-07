"use client";

import { useEffect, useRef, useState } from "react";
import { Box, Button, Center, Flex, Heading, HStack, Input, NativeSelect, Spinner, Table, Text } from "@chakra-ui/react";
import {
	LuAppWindow,
	LuBadgeCheck,
	LuChevronLeft,
	LuChevronRight,
	LuCircleCheck,
	LuCircleSlash,
	LuCircleX,
	LuCopy,
	LuPlus,
	LuRefreshCw,
	LuSearch,
} from "react-icons/lu";
import { listAppServices, updateAppServiceStatus } from "@/apis";
import { AppServiceSecretDialog } from "@/components/AppServiceSecretDialog";
import { RegisterAppServiceDialog } from "@/components/RegisterAppServiceDialog";
import { RotateAppServiceSecretDialog } from "@/components/RotateAppServiceSecretDialog";
import { TerminateAppServiceDialog } from "@/components/TerminateAppServiceDialog";
import { VerifyAppServiceDialog } from "@/components/VerifyAppServiceDialog";
import { toaster } from "@/components/ui/toaster";
import { APP_SERVICE_CTX_INFO_OPTIONS, APP_SERVICE_STATUS, APP_SERVICE_STATUS_FILTER_OPTIONS, APP_SERVICES_PAGE_SIZE_OPTIONS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { useCurrentUser } from "@/hooks/useCurrentUser";
import { AppShell } from "@/layouts/AppShell";
import type { AppService, AppServiceCtxInfo, AppServiceStatus } from "@/types";
import { formatDateOnly } from "@/utils";

type StatusFilter = (typeof APP_SERVICE_STATUS_FILTER_OPTIONS)[number];

const STATUS_FILTER_TO_VALUE: Record<Exclude<StatusFilter, "all">, AppServiceStatus> = {
	active: 1,
	inactive: 2,
	terminated: 3,
};

const STATUS_FILTER_ICONS: Partial<Record<StatusFilter, React.ReactNode>> = {
	active: <LuCircleCheck size={13} />,
	inactive: <LuCircleSlash size={13} />,
	terminated: <LuCircleX size={13} />,
};

const TILE_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
];

/** Uppercase micro-label (mock thead th). */
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

const tileGradient = (id: string) => {
	let hash = 0;
	for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) % 997;
	return TILE_GRADIENTS[hash % TILE_GRADIENTS.length];
};

const StatusPill = ({ status }: { status: AppServiceStatus }) => {
	if (status === APP_SERVICE_STATUS.ACTIVE) {
		return (
			<Box {...PILL_BASE} color="aurora.mint" borderColor="rgba(52,211,153,0.35)" bg="rgba(52,211,153,0.10)">
				<Box w="6px" h="6px" borderRadius="full" bg="aurora.mint" css={{ boxShadow: "0 0 10px #34D399" }} />
				active
			</Box>
		);
	}
	if (status === APP_SERVICE_STATUS.INACTIVE) {
		return (
			<Box {...PILL_BASE} color="fg.muted">
				<Box w="6px" h="6px" borderRadius="full" bg="fg.muted" />
				inactive
			</Box>
		);
	}
	return (
		<Box {...PILL_BASE} color="aurora.magenta" borderColor="rgba(236,72,153,0.35)" bg="rgba(236,72,153,0.10)">
			<Box w="6px" h="6px" borderRadius="full" bg="aurora.magenta" css={{ boxShadow: "0 0 10px #EC4899" }} />
			terminated
		</Box>
	);
};

const CtxPill = ({ ctxInfo }: { ctxInfo: AppService["ctx_info"] }) => {
	if (ctxInfo === "authen") {
		return (
			<Box {...PILL_BASE} color="aurora.violet" borderColor="rgba(139,92,246,0.40)" bg="rgba(139,92,246,0.12)">
				authen
			</Box>
		);
	}
	return (
		<Box {...PILL_BASE} color="aurora.cyan" borderColor="rgba(34,211,238,0.40)" bg="rgba(34,211,238,0.10)">
			app_service
		</Box>
	);
};

const AppTile = ({ service }: { service: AppService }) => {
	if (service.status === APP_SERVICE_STATUS.TERMINATED) {
		return (
			<Center
				w="34px"
				h="34px"
				flex="none"
				borderRadius="10px"
				bg="bg.glass"
				borderWidth="1.5px"
				borderStyle="dashed"
				borderColor="border.strong"
				color="fg.muted"
			>
				<LuAppWindow size={16} />
			</Center>
		);
	}
	const inactive = service.status === APP_SERVICE_STATUS.INACTIVE;
	return (
		<Center
			w="34px"
			h="34px"
			flex="none"
			borderRadius="10px"
			color={inactive ? "fg.subtle" : "white"}
			css={{
				background: inactive ? "linear-gradient(135deg, #3A3F6E, #23264A)" : tileGradient(service.id),
				boxShadow: inactive ? "none" : "0 0 14px rgba(99,102,241,0.30)",
			}}
		>
			<LuAppWindow size={16} />
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

/**
 * Secret column sub-line: secrets are AES-encrypted at rest and irrecoverable
 * by design — the mask is permanent, rotation is the only recovery path.
 */
const secretSubLabel = (service: AppService) => {
	if (service.status === APP_SERVICE_STATUS.TERMINATED) return "revoked";
	if (!service.updated_at || service.updated_at === service.created_at) return "never updated";
	return `updated ${formatDateOnly(service.updated_at)}`;
};

export const AppServices = () => {
	const { user: currentUser } = useCurrentUser();

	const [services, setServices] = useState<AppService[]>([]);
	const [total, setTotal] = useState(0);
	const [loading, setLoading] = useState(true);
	// Bumped to force a server refetch after register/rotate/status mutations.
	const [reloadKey, setReloadKey] = useState(0);

	const [search, setSearch] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const searchRef = useRef<HTMLInputElement>(null);
	const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
	const [ctxFilter, setCtxFilter] = useState("all");

	const [page, setPage] = useState(1);
	const [pageSize, setPageSize] = useState(APP_SERVICES_PAGE_SIZE_OPTIONS[0].value);

	const [registerOpen, setRegisterOpen] = useState(false);
	const [verifyOpen, setVerifyOpen] = useState(false);
	const [rotateTarget, setRotateTarget] = useState<AppService | null>(null);
	const [terminateTarget, setTerminateTarget] = useState<AppService | null>(null);
	// One-time secret hand-off from register/rotate — cleared only via Done.
	const [secretInfo, setSecretInfo] = useState<{ title: string; appCode: string; secret: string } | null>(null);

	// Debounce search before hitting the server.
	useEffect(() => {
		const timer = setTimeout(() => {
			setDebouncedSearch(search.trim());
			setPage(1);
		}, 300);
		return () => clearTimeout(timer);
	}, [search]);

	// Server-side list: search/status/ctx/page/size are pushed to GET /api/v1/app-service.
	// Refetches keep the previous page visible (no spinner) — `loading` only
	// covers the initial load.
	useEffect(() => {
		let active = true;
		listAppServices({
			page,
			page_size: pageSize,
			search: debouncedSearch || undefined,
			status: statusFilter !== "all" ? STATUS_FILTER_TO_VALUE[statusFilter] : undefined,
			ctx_info: ctxFilter !== "all" ? (ctxFilter as AppServiceCtxInfo) : undefined,
		})
			.then((response) => {
				if (!active) return;
				setServices(response.items);
				setTotal(response.total);
			})
			.catch(() => {
				if (active) toaster.create({ title: "Failed to load app services", type: "error", meta: { closable: true } });
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, [page, pageSize, debouncedSearch, statusFilter, ctxFilter, reloadKey]);

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

	// Search/status/ctx/pagination are server-side — `services` is the current page.
	const paged = services;
	const totalPages = Math.max(1, Math.ceil(total / pageSize));
	const safePage = Math.min(page, totalPages);

	const hasActiveFilters = debouncedSearch !== "" || statusFilter !== "all" || ctxFilter !== "all";

	// Page-head counts from the list response: `total` is server-side, the
	// active/terminated splits cover the fetched page only.
	const activeCount = services.filter((service) => service.status === APP_SERVICE_STATUS.ACTIVE).length;
	const terminatedCount = services.filter((service) => service.status === APP_SERVICE_STATUS.TERMINATED).length;

	const handleCopyCode = async (service: AppService) => {
		try {
			await navigator.clipboard.writeText(service.app_code);
			toaster.create({ title: `Copied "${service.app_code}"`, type: "success", meta: { closable: true } });
		} catch {
			toaster.create({ title: "Failed to copy app code", type: "error", meta: { closable: true } });
		}
	};

	// Activate ⇄ deactivate (PATCH 1/2). Terminate (3) goes through the
	// type-to-confirm dialog instead — it is terminal server-side.
	const handleToggleStatus = async (service: AppService) => {
		const nextStatus: AppServiceStatus = service.status === APP_SERVICE_STATUS.ACTIVE ? APP_SERVICE_STATUS.INACTIVE : APP_SERVICE_STATUS.ACTIVE;
		const verb = nextStatus === APP_SERVICE_STATUS.ACTIVE ? "Activate" : "Deactivate";
		if (!window.confirm(`${verb} "${service.app_name}" (${service.app_code})?`)) return;
		try {
			await updateAppServiceStatus(service.id, nextStatus);
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({
				title: err?.response?.data?.message || `Failed to ${verb.toLowerCase()} ${service.app_code}`,
				type: "error",
				meta: { closable: true },
			});
			return;
		}
		setReloadKey((previous) => previous + 1);
		toaster.create({
			title: `${service.app_name} ${nextStatus === APP_SERVICE_STATUS.ACTIVE ? "activated" : "deactivated"}`,
			type: "success",
			meta: { closable: true },
		});
	};

	const handleTerminated = (service: AppService) => {
		setTerminateTarget(null);
		setReloadKey((previous) => previous + 1);
		toaster.create({ title: `${service.app_name} terminated`, type: "success", meta: { closable: true } });
	};

	const handleRegistered = (appCode: string, appSecret: string) => {
		setRegisterOpen(false);
		setSecretInfo({ title: "App registered", appCode, secret: appSecret });
	};

	const handleRotated = (appCode: string, appSecret: string) => {
		setRotateTarget(null);
		setSecretInfo({ title: "Secret rotated", appCode, secret: appSecret });
	};

	const handleSecretDone = () => {
		setSecretInfo(null);
		setReloadKey((previous) => previous + 1);
	};

	const registerButton = (
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
			onClick={() => setRegisterOpen(true)}
		>
			<LuPlus size={16} /> Register app
		</Button>
	);

	return (
		<AppShell active="appServices" user={{ name: currentUser?.name || "User", email: currentUser?.email || "" }}>
			{/* Page head */}
			<Flex align="flex-end" justify="space-between" gap="4" wrap="wrap">
				<Box>
					<Heading fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1" color="fg">
						App Services
					</Heading>
					<Text fontSize="sm" color="fg.muted" mt="1.5">
						<Text as="span" color="aurora.cyan" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
							{total.toLocaleString()}
						</Text>{" "}
						registered ·{" "}
						<Text as="span" color="aurora.mint" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
							{activeCount.toLocaleString()}
						</Text>{" "}
						active ·{" "}
						<Text as="span" color="aurora.magenta" fontWeight="semibold" css={{ fontVariantNumeric: "tabular-nums" }}>
							{terminatedCount.toLocaleString()}
						</Text>{" "}
						terminated on this page
					</Text>
				</Box>
				<HStack gap="2.5">
					<Button h="11" px="4.5" fontSize="sm" {...GHOST_BUTTON_PROPS} onClick={() => setVerifyOpen(true)}>
						<LuBadgeCheck size={16} /> Verify credentials
					</Button>
					{registerButton}
				</HStack>
			</Flex>

			{/* App services glass panel */}
			<Box
				bg="bg.glass"
				borderWidth="1px"
				borderColor="border"
				borderRadius="20px"
				boxShadow="glassSoft"
				overflow="hidden"
				css={{ backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" }}
			>
				{loading ? (
					<Center py="16">
						<Spinner size="lg" color="accent" />
					</Center>
				) : total === 0 && !hasActiveFilters ? (
					/* Empty state — only when no apps exist at all; filtered-empty keeps the toolbar. */
					<Center px="6" py="16">
						<Box textAlign="center" maxW="380px">
							<Center
								w="14"
								h="14"
								mx="auto"
								mb="4.5"
								borderRadius="16px"
								bg="bg.glass"
								borderWidth="1px"
								borderColor="border.strong"
								color="fg.muted"
								boxShadow="0 0 30px rgba(99,102,241,0.18)"
							>
								<LuAppWindow size={24} />
							</Center>
							<Heading fontSize="md" fontWeight="semibold" color="fg" mb="1.5">
								No app services yet
							</Heading>
							<Text fontSize="13px" color="fg.muted" mb="4.5">
								Register an application to issue it SSO credentials. The secret is shown once at creation — store it safely.
							</Text>
							{registerButton}
						</Box>
					</Center>
				) : (
					<>
						{/* Toolbar: search + status segmented + ctx select */}
						<Flex align="center" gap="2.5" p="3.5" wrap="wrap" borderBottomWidth="1px" borderColor="border">
							<Box flex="1" minW="240px" position="relative" css={{ "&:focus-within .search-icon": { color: "#22D3EE" } }}>
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
									placeholder="Search by app name or code…"
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
							<HStack
								gap="0"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border.strong"
								overflow="hidden"
								w="fit-content"
								bg="bg.glass"
								css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
								role="group"
								aria-label="Status filter"
							>
								{APP_SERVICE_STATUS_FILTER_OPTIONS.map((option, index) => (
									<HStack
										key={option}
										as="button"
										h="11"
										px="3.5"
										gap="1.5"
										fontSize="13px"
										fontWeight="medium"
										cursor="pointer"
										bg={statusFilter === option ? "linear-gradient(135deg, rgba(99,102,241,0.30), rgba(139,92,246,0.25))" : "transparent"}
										boxShadow={statusFilter === option ? "0 0 16px rgba(139,92,246,0.25) inset" : "none"}
										color={statusFilter === option ? "fg" : "fg.muted"}
										borderRightWidth={index < APP_SERVICE_STATUS_FILTER_OPTIONS.length - 1 ? "1px" : "0"}
										borderColor="border"
										_hover={{ color: "fg" }}
										onClick={() => {
											setStatusFilter(option);
											setPage(1);
										}}
									>
										{STATUS_FILTER_ICONS[option]}
										{option}
									</HStack>
								))}
							</HStack>
							<NativeSelect.Root size="sm" w="36">
								<NativeSelect.Field
									h="11"
									borderRadius="glassSm"
									bg="bg.glass"
									borderColor="border.strong"
									fontSize="13px"
									color="fg"
									aria-label="Context filter"
									css={{
										backdropFilter: "blur(12px)",
										WebkitBackdropFilter: "blur(12px)",
										"& option": { background: "#12122E", color: "#F4F5FF" },
									}}
									_focus={{ borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none" }}
									value={ctxFilter}
									onChange={(event) => {
										setCtxFilter(event.target.value);
										setPage(1);
									}}
								>
									<option value="all">ctx: all</option>
									{APP_SERVICE_CTX_INFO_OPTIONS.map((option) => (
										<option key={option.value} value={option.value}>
											{option.value}
										</option>
									))}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Flex>

						{/* Table */}
						<Table.Root size="sm" css={{ "& td, & th": { borderColor: "rgba(255,255,255,0.10)" } }}>
							<Table.Header>
								<Table.Row bg="transparent">
									{["Application", "Ctx", "Redirect URL", "Secret", "Status", "Created"].map((label) => (
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
										<Table.Cell colSpan={7} px="3.5" py="8" textAlign="center" fontSize="sm" color="fg.muted">
											No app services match the current search / filters.
										</Table.Cell>
									</Table.Row>
								)}
								{paged.map((service) => {
									const terminated = service.status === APP_SERVICE_STATUS.TERMINATED;
									const inactive = service.status === APP_SERVICE_STATUS.INACTIVE;
									return (
										<Table.Row
											key={service.id}
											transition="background .15s ease-out"
											opacity={terminated ? 0.45 : inactive ? 0.65 : 1}
											bg="transparent"
											_hover={{ bg: "rgba(255,255,255,0.04)" }}
											css={{
												"& .row-actions": { opacity: 0, transition: "opacity .15s ease-out" },
												"&:hover .row-actions": { opacity: 1 },
											}}
										>
											<Table.Cell px="3.5" py="13px">
												<HStack gap="3" minW="0" maxW="280px">
													<AppTile service={service} />
													<Box minW="0" lineHeight="1.25">
														<Text
															fontSize="sm"
															fontWeight="semibold"
															truncate
															color={terminated ? "fg.subtle" : "fg"}
															textDecoration={terminated ? "line-through" : "none"}
															title={service.app_name}
														>
															{service.app_name}
														</Text>
														<Text fontSize="12px" color="fg.muted" truncate css={{ fontVariantNumeric: "tabular-nums" }}>
															{service.app_code}
														</Text>
													</Box>
												</HStack>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<CtxPill ctxInfo={service.ctx_info} />
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<Text display="block" maxW="240px" fontSize="13px" color="fg.subtle" truncate title={service.redirect_url}>
													{service.redirect_url}
												</Text>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px" lineHeight="1.3">
												{/* Permanent mask — AES-encrypted at rest, no reveal affordance by design. */}
												<Text
													fontSize="13px"
													color="fg.muted"
													letterSpacing="3px"
													css={{ userSelect: "none" }}
													title="Secret is encrypted at rest and cannot be revealed. Rotate to issue a new one."
												>
													••••••••
												</Text>
												<Text fontSize="11px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
													{secretSubLabel(service)}
												</Text>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<StatusPill status={service.status} />
											</Table.Cell>
											<Table.Cell px="3.5" py="13px" {...NUM_PROPS} color="fg.muted">
												{formatDateOnly(service.created_at)}
												<Text display="block" fontSize="11px" color="fg.muted">
													by {service.created_by_email || service.created_by || "—"}
												</Text>
											</Table.Cell>
											<Table.Cell px="3.5" py="13px">
												<HStack className="row-actions" gap="1.5" justify="flex-end">
													<QuickActionButton
														label={terminated ? "Terminated apps cannot rotate secrets" : "Rotate secret"}
														color="aurora.cyan"
														disabled={terminated}
														onClick={() => setRotateTarget(service)}
													>
														<LuRefreshCw size={14} />
													</QuickActionButton>
													<QuickActionButton label="Copy app code" color="fg.subtle" onClick={() => handleCopyCode(service)}>
														<LuCopy size={14} />
													</QuickActionButton>
													{service.status === APP_SERVICE_STATUS.ACTIVE ? (
														<QuickActionButton label="Deactivate" color="aurora.amber" onClick={() => handleToggleStatus(service)}>
															<LuCircleSlash size={14} />
														</QuickActionButton>
													) : (
														<QuickActionButton
															label={terminated ? "Termination is permanent" : "Activate"}
															color="aurora.mint"
															disabled={terminated}
															onClick={() => handleToggleStatus(service)}
														>
															<LuCircleCheck size={14} />
														</QuickActionButton>
													)}
													{!terminated && (
														<QuickActionButton label="Terminate" color="aurora.magenta" danger onClick={() => setTerminateTarget(service)}>
															<LuCircleX size={14} />
														</QuickActionButton>
													)}
												</HStack>
											</Table.Cell>
										</Table.Row>
									);
								})}
							</Table.Body>
						</Table.Root>

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
								Page {safePage} / {totalPages} · {paged.length} of {total.toLocaleString()} shown
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
										aria-label="Rows per page"
										css={{ "& option": { background: "#12122E", color: "#F4F5FF" } }}
										value={String(pageSize)}
										onChange={(event) => {
											setPageSize(Number(event.target.value));
											setPage(1);
										}}
									>
										{APP_SERVICES_PAGE_SIZE_OPTIONS.map((option) => (
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
					</>
				)}
			</Box>

			<RegisterAppServiceDialog open={registerOpen} onOpenChange={setRegisterOpen} onRegistered={handleRegistered} />
			<RotateAppServiceSecretDialog
				open={rotateTarget !== null}
				appService={rotateTarget}
				onOpenChange={(next) => {
					if (!next) setRotateTarget(null);
				}}
				onRotated={handleRotated}
			/>
			<TerminateAppServiceDialog
				open={terminateTarget !== null}
				appService={terminateTarget}
				onOpenChange={(next) => {
					if (!next) setTerminateTarget(null);
				}}
				onTerminated={handleTerminated}
			/>
			<VerifyAppServiceDialog open={verifyOpen} onOpenChange={setVerifyOpen} />
			<AppServiceSecretDialog
				open={secretInfo !== null}
				title={secretInfo?.title ?? ""}
				appCode={secretInfo?.appCode ?? ""}
				secret={secretInfo?.secret ?? ""}
				onDone={handleSecretDone}
			/>
		</AppShell>
	);
};
