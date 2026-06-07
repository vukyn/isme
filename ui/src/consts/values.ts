import type { PageSizeOption } from "@/types/base";

export const PAGE_SIZE_OPTIONS: PageSizeOption[] = [
	{ value: 10, label: "10" },
	{ value: 20, label: "20" },
	{ value: 50, label: "50" },
	{ value: 100, label: "100" },
];

/** user.status mapping — 3 (pending) is forward-looking (invite flow). */
export const USER_STATUS = {
	ACTIVE: 1,
	INACTIVE: 2,
	PENDING: 3,
} as const;

/** Forward-looking RBAC role options (no role domain yet). */
export const USER_ROLE_OPTIONS = ["admin", "member", "viewer", "content-ops"] as const;

export const USER_STATUS_FILTER_OPTIONS = ["all", "active", "inactive", "pending"] as const;

export const USER_LAST_LOGIN_FILTER_OPTIONS = [
	{ value: "any", label: "any time" },
	{ value: "24h", label: "last 24h" },
	{ value: "7d", label: "last 7 days" },
	{ value: "30d", label: "last 30 days" },
	{ value: "never", label: "never" },
] as const;

export const USERS_PAGE_SIZE_OPTIONS: PageSizeOption[] = [
	{ value: 10, label: "10" },
	{ value: 25, label: "25" },
	{ value: 50, label: "50" },
];

/** app_service.status mapping — terminated (3) is terminal. */
export const APP_SERVICE_STATUS = {
	ACTIVE: 1,
	INACTIVE: 2,
	TERMINATED: 3,
} as const;

/** ctx_info fixed set — mirrors constants.AllowedCtxInfos. */
export const APP_SERVICE_CTX_INFO_OPTIONS = [
	{ value: "authen", label: "authen — browser SSO flow" },
	{ value: "app_service", label: "app_service — service-to-service" },
] as const;

export const APP_SERVICE_STATUS_FILTER_OPTIONS = ["all", "active", "inactive", "terminated"] as const;

export const APP_SERVICES_PAGE_SIZE_OPTIONS: PageSizeOption[] = [
	{ value: 10, label: "10" },
	{ value: 25, label: "25" },
	{ value: 50, label: "50" },
];
