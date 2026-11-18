"use client";

import type { ReactNode } from "react";
import { Center } from "@chakra-ui/react";
import { USER_STATUS } from "@/consts";
import type { InvitationListItem, UserListItem } from "@/types";

const AVATAR_GRADIENTS = [
	"conic-gradient(from 200deg, #22D3EE, #6366F1, #8B5CF6, #EC4899, #22D3EE)",
	"linear-gradient(135deg, #22D3EE, #34D399)",
	"linear-gradient(135deg, #F59E0B, #EC4899)",
	"linear-gradient(135deg, #34D399, #22D3EE)",
];

const initials = (name: string) =>
	name
		.split(" ")
		.filter(Boolean)
		.map((part) => part[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

const emailInitials = (email: string) =>
	email
		.split("@")[0]
		.split(/[._-]/)
		.filter(Boolean)
		.map((part) => part[0]?.toUpperCase())
		.slice(0, 2)
		.join("") || "?";

const avatarGradient = (id: string) => {
	let hash = 0;
	for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) % 997;
	return AVATAR_GRADIENTS[hash % AVATAR_GRADIENTS.length];
};

/** Pixel size → initials font size, so one avatar serves the table and the dialogs. */
const fontSizeFor = (size: number) => (size <= 28 ? "11px" : size >= 40 ? "13px" : "12px");

const DashedAvatar = ({ size, children }: { size: number; children: ReactNode }) => (
	<Center
		w={`${size}px`}
		h={`${size}px`}
		flex="none"
		borderRadius="full"
		bg="bg.glass"
		borderWidth="1.5px"
		borderStyle="dashed"
		borderColor="border.strong"
		fontSize={fontSizeFor(size)}
		fontWeight="bold"
		color="fg.muted"
		aria-hidden="true"
	>
		{children}
	</Center>
);

/** Dashed "?" = pending · slate = inactive · gradient keyed on the id = active. */
export const UserAvatar = ({ user, size = 32 }: { user: UserListItem; size?: number }) => {
	if (user.status === USER_STATUS.PENDING) {
		return <DashedAvatar size={size}>?</DashedAvatar>;
	}
	const inactive = user.status === USER_STATUS.INACTIVE;
	return (
		<Center
			w={`${size}px`}
			h={`${size}px`}
			flex="none"
			borderRadius="full"
			fontSize={fontSizeFor(size)}
			fontWeight="bold"
			color="white"
			aria-hidden="true"
			css={{
				background: inactive ? "linear-gradient(135deg, #3A3F6E, #23264A)" : avatarGradient(user.id),
				boxShadow: inactive ? "none" : "0 0 14px rgba(99,102,241,0.30)",
			}}
		>
			{initials(user.name)}
		</Center>
	);
};

/**
 * Dashed avatar = no account yet; gradient = accepted (account exists).
 * `icon` replaces the initials (the revoke dialog shows a mail glyph).
 */
export const InvitationAvatar = ({
	invitation,
	size = 32,
	icon,
}: {
	invitation: InvitationListItem;
	size?: number;
	icon?: ReactNode;
}) => {
	if (invitation.status === 2) {
		return (
			<Center
				w={`${size}px`}
				h={`${size}px`}
				flex="none"
				borderRadius="full"
				fontSize={fontSizeFor(size)}
				fontWeight="bold"
				color="white"
				aria-hidden="true"
				css={{ background: "linear-gradient(135deg, #22D3EE, #34D399)", boxShadow: "0 0 14px rgba(52,211,153,0.30)" }}
			>
				{icon ?? emailInitials(invitation.email)}
			</Center>
		);
	}
	return <DashedAvatar size={size}>{icon ?? emailInitials(invitation.email)}</DashedAvatar>;
};
