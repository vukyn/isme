// Avatar helpers shared across the Users / Roles / Profile screens. The conic
// gradient is the aurora default (matches the topbar UserChip + the mock's
// .avatar-xl / .avatar-preview); the linear variants give per-id variety.
const AVATAR_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
	"linear-gradient(135deg, #F59E0B, #EC4899)",
	"linear-gradient(135deg, #34D399, #22D3EE)",
];

/** Deterministic gradient for a stable id (same id → same gradient). */
export const avatarGradient = (id: string): string => {
	let hash = 0;
	for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) % 997;
	return AVATAR_GRADIENTS[hash % AVATAR_GRADIENTS.length];
};

/** Up-to-two-letter uppercase initials from a display name (fallback "?"). */
export const avatarInitials = (name: string): string =>
	name
		.split(" ")
		.filter(Boolean)
		.map((part) => part[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";
