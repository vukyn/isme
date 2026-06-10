/**
 * Per-resource icon registry for the permission catalog/matrix.
 *
 * A resource's icon is persisted as a STRING KEY on its permissions (see
 * internal/domains/role/entity/permission.go + the backend allowlist in
 * internal/domains/role/constants/constants.go) — never an SVG. This registry
 * maps each allowed key to a react-icons node so the UI can render it. Keep the
 * key set in sync with the backend allowlist (roleConstants.PERMISSION_ICON_KEYS).
 */
import type { ReactNode } from "react";
import {
	LuBell,
	LuBox,
	LuCloud,
	LuCloudRain,
	LuCode,
	LuDatabase,
	LuFile,
	LuFolder,
	LuGlobe,
	LuImage,
	LuKey,
	LuLayers,
	LuLock,
	LuMusic,
	LuServer,
	LuSettings,
	LuShapes,
	LuShield,
	LuStar,
	LuTag,
	LuUser,
	LuUsers,
	LuVideo,
} from "react-icons/lu";

/** Each entry renders the key's icon at the requested pixel size. */
type IconRenderer = (size: number) => ReactNode;

/** The allowlist of icon keys — must match roleConstants.PERMISSION_ICON_KEYS. */
export const PERMISSION_ICON_KEYS = [
	"file",
	"folder",
	"database",
	"box",
	"image",
	"music",
	"video",
	"shield",
	"key",
	"user",
	"users",
	"globe",
	"tag",
	"lock",
	"bell",
	"star",
	"settings",
	"server",
	"cloud",
	"code",
	"layers",
	"cloud-rain",
	"isme",
] as const;

export type PermissionIconKey = (typeof PERMISSION_ICON_KEYS)[number];

const REGISTRY: Record<PermissionIconKey, IconRenderer> = {
	file: (size) => <LuFile size={size} />,
	folder: (size) => <LuFolder size={size} />,
	database: (size) => <LuDatabase size={size} />,
	box: (size) => <LuBox size={size} />,
	image: (size) => <LuImage size={size} />,
	music: (size) => <LuMusic size={size} />,
	video: (size) => <LuVideo size={size} />,
	shield: (size) => <LuShield size={size} />,
	key: (size) => <LuKey size={size} />,
	user: (size) => <LuUser size={size} />,
	users: (size) => <LuUsers size={size} />,
	globe: (size) => <LuGlobe size={size} />,
	tag: (size) => <LuTag size={size} />,
	lock: (size) => <LuLock size={size} />,
	bell: (size) => <LuBell size={size} />,
	star: (size) => <LuStar size={size} />,
	settings: (size) => <LuSettings size={size} />,
	server: (size) => <LuServer size={size} />,
	cloud: (size) => <LuCloud size={size} />,
	code: (size) => <LuCode size={size} />,
	layers: (size) => <LuLayers size={size} />,
	"cloud-rain": (size) => <LuCloudRain size={size} />,
	// isme brand "i" monogram (the favicon mark, demo/aurora-favicon.svg) drawn
	// monochrome in currentColor so it tints like every other registry icon.
	isme: (size) => (
		<svg width={size} height={size} viewBox="0 0 32 32" fill="currentColor" aria-hidden="true">
			<circle cx="16" cy="10" r="3.6" />
			<rect x="12.4" y="16.4" width="7.2" height="9.6" rx="3.6" />
		</svg>
	),
};

/** Neutral fallback for an empty or unknown icon key. */
export const renderDefaultPermissionIcon = (size = 14): ReactNode => <LuShapes size={size} />;

/**
 * Renders the icon for a resource's stored key. Empty or unknown keys fall back
 * to the neutral default so brand-new/custom resources never show a misleading
 * icon.
 */
export const renderPermissionIcon = (key: string | undefined, size = 14): ReactNode => {
	if (key && key in REGISTRY) {
		return REGISTRY[key as PermissionIconKey](size);
	}
	return renderDefaultPermissionIcon(size);
};

/**
 * Maps known isme built-in resources to registry keys so the seeded catalog
 * (whose rows have an empty stored icon) keeps meaningful icons. Custom
 * resources rely on their persisted icon instead.
 */
export const ISME_RESOURCE_ICON_KEYS: Record<string, PermissionIconKey> = {
	user: "user",
	user_session: "settings",
	session: "settings",
	app_service: "box",
	role: "key",
};

/**
 * Resolves the icon key to render for a resource: the resource's persisted key
 * wins; otherwise an isme built-in mapping is used; otherwise empty (neutral).
 */
export const resolveResourceIconKey = (resource: string, storedKey: string | undefined): string => {
	if (storedKey) return storedKey;
	return ISME_RESOURCE_ICON_KEYS[resource] ?? "";
};

/**
 * Generic aliases — the icon registry is shared by the permission catalog AND
 * the app_service appearance tile (same backend allowlist, roleConstants.ICON_KEYS).
 * Prefer these names outside the permission domain.
 */
export const ICON_KEYS = PERMISSION_ICON_KEYS;
export type IconKey = PermissionIconKey;
export const renderIcon = renderPermissionIcon;
export const renderDefaultIcon = renderDefaultPermissionIcon;
