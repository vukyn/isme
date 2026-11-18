"use client";

import { useCallback, useRef, useState } from "react";
import { Box, Button, HStack, Text } from "@chakra-ui/react";
import { LuCircleX, LuLink2Off, LuMailX, LuRotateCw, LuTriangleAlert } from "react-icons/lu";
import { revokeInvitation } from "@/apis";
import { DestructiveDialog, Fact, FactList, Lead, Note } from "@/components/DestructiveDialog";
import { InvitationAvatar } from "@/components/UserAvatar";
import { toaster } from "@/components/ui/toaster";
import { DANGER_DIALOG_BUTTON_PROPS, GHOST_DIALOG_BUTTON_PROPS } from "@/consts/styles";
import { type RequestFailure, toRequestFailure } from "@/utils";
import type { InvitationListItem } from "@/types";

/** Spinner shows only when a request outlasts this, so a fast revoke never flashes one. */
const SPINNER_DELAY_MS = 150;
/** Role pills shown before collapsing the rest into "+N" (the row never wraps). */
const PILL_PREVIEW_COUNT = 2;

type Phase = "confirm" | "busy" | "error" | "stale";

/** The invitation changed state before the revoke landed; nothing left to do. */
type StaleReason = "accepted" | "revoked" | "expired";

const STALE_MESSAGES: Record<string, StaleReason> = {
	"invitation already used": "accepted",
	"invitation already revoked": "revoked",
	"invitation already expired": "expired",
};

const STALE_TITLES: Record<StaleReason, string> = {
	accepted: "Invite already accepted",
	revoked: "Already revoked",
	expired: "Invite already expired",
};

interface RevokeInvitationDialogProps {
	open: boolean;
	invitation: InvitationListItem | null;
	onOpenChange: (open: boolean) => void;
	/** Called once the invitation is no longer pending, so the page can refetch. */
	onSettled: () => void;
}

const DAY_MS = 24 * 60 * 60 * 1000;
const HOUR_MS = 60 * 60 * 1000;

/** "2 days" / "5 hours" / "a few minutes" — coarse on purpose. */
const spanLabel = (milliseconds: number) => {
	const days = Math.floor(milliseconds / DAY_MS);
	if (days >= 1) return `${days} day${days === 1 ? "" : "s"}`;
	const hours = Math.floor(milliseconds / HOUR_MS);
	if (hours >= 1) return `${hours} hour${hours === 1 ? "" : "s"}`;
	return "a few minutes";
};

const sentAndExpiry = (invitation: InvitationListItem, now: number) => {
	const sent = `Sent ${spanLabel(Math.max(0, now - new Date(invitation.created_at).getTime()))} ago`;
	const remaining = new Date(invitation.expires_at).getTime() - now;
	return remaining > 0 ? `${sent} · expires in ${spanLabel(remaining)}` : `${sent} · expired`;
};

const RolePill = ({ admin, children }: { admin: boolean; children: React.ReactNode }) => (
	<Box
		as="span"
		flex="none"
		px="10px"
		py="5px"
		borderRadius="full"
		borderWidth="1px"
		borderColor={admin ? "rgba(139,92,246,0.40)" : "rgba(34,211,238,0.40)"}
		bg={admin ? "rgba(139,92,246,0.12)" : "rgba(34,211,238,0.10)"}
		color={admin ? "aurora.violet" : "aurora.cyan"}
		fontSize="12px"
		fontWeight="medium"
		lineHeight="1"
		whiteSpace="nowrap"
	>
		{children}
	</Box>
);

/**
 * Revoke-invite confirmation (aurora mock C1-C4), replacing the browser's
 * own confirm box. A revoke that lost a race (accepted, already revoked or
 * expired) is an outcome, not an error: it shows a neutral note and the
 * page refetches. Mount it with a fresh `key` per open.
 */
export const RevokeInvitationDialog = ({ open, invitation, onOpenChange, onSettled }: RevokeInvitationDialogProps) => {
	const cancelRef = useRef<HTMLButtonElement>(null);
	const closeRef = useRef<HTMLButtonElement>(null);
	const [phase, setPhase] = useState<Phase>("confirm");
	const [showSpinner, setShowSpinner] = useState(false);
	const [failure, setFailure] = useState<RequestFailure | null>(null);
	const [staleReason, setStaleReason] = useState<StaleReason>("accepted");
	// captured once per mount: render must stay pure
	const [openedAt] = useState(() => Date.now());

	const focusAlert = useCallback((node: HTMLDivElement | null) => node?.focus(), []);

	if (!invitation) return null;

	const busy = phase === "busy";
	const close = () => {
		if (!busy) onOpenChange(false);
	};

	const handleRevoke = async () => {
		setPhase("busy");
		setFailure(null);
		const timer = setTimeout(() => setShowSpinner(true), SPINNER_DELAY_MS);
		try {
			await revokeInvitation(invitation.id);
			onSettled();
			toaster.create({ title: `Invite for ${invitation.email} revoked`, type: "success", meta: { closable: true } });
			onOpenChange(false);
		} catch (error: unknown) {
			const requestFailure = toRequestFailure(error);
			const stale = requestFailure.status === 400 ? STALE_MESSAGES[requestFailure.message] : undefined;
			if (stale) {
				setStaleReason(stale);
				setPhase("stale");
				onSettled();
			} else {
				setFailure(requestFailure);
				setPhase("error");
			}
		} finally {
			clearTimeout(timer);
			setShowSpinner(false);
		}
	};

	if (phase === "stale") {
		return (
			<DestructiveDialog
				open={open}
				onClose={close}
				busy={false}
				tone="neutral"
				icon={<LuTriangleAlert size={18} />}
				iconTone="amber"
				title={STALE_TITLES[staleReason]}
				initialFocusRef={closeRef}
				footer={
					<Button ref={closeRef} {...GHOST_DIALOG_BUTTON_PROPS} onClick={close}>
						Close
					</Button>
				}
			>
				<Lead>
					{staleReason === "accepted" && (
						<>
							<b>{invitation.email}</b> accepted this invite before it was revoked. They now have an account. To remove their access,
							delete or deactivate them in Users.
						</>
					)}
					{staleReason === "revoked" && (
						<>
							The invite for <b>{invitation.email}</b> had already been revoked, possibly by another admin. The list has been refreshed.
						</>
					)}
					{staleReason === "expired" && (
						<>
							The invite for <b>{invitation.email}</b> had already expired, so its link no longer works. The list has been refreshed.
						</>
					)}
				</Lead>
			</DestructiveDialog>
		);
	}

	const canRetry = phase === "error" && failure?.status !== 403;
	const previewAssignments = invitation.assignments.slice(0, PILL_PREVIEW_COUNT);
	const hiddenAssignments = invitation.assignments.length - previewAssignments.length;

	return (
		<DestructiveDialog
			open={open}
			onClose={close}
			busy={busy}
			tone="danger"
			icon={<LuLink2Off size={18} />}
			iconTone="magenta"
			title="Revoke invite?"
			initialFocusRef={cancelRef}
			footer={
				<>
					<Button ref={cancelRef} {...GHOST_DIALOG_BUTTON_PROPS} disabled={busy} onClick={close}>
						Cancel
					</Button>
					{(phase !== "error" || canRetry) && (
						<Button
							{...DANGER_DIALOG_BUTTON_PROPS}
							disabled={busy}
							loading={busy && showSpinner}
							loadingText="Revoking…"
							onClick={() => void handleRevoke()}
						>
							{phase === "error" ? <LuRotateCw size={15} /> : <LuLink2Off size={15} />}
							{phase === "error" ? "Try again" : "Revoke invite"}
						</Button>
					)}
				</>
			}
		>
			<HStack gap="12px" minW="0" px="14px" py="12px" borderRadius="glassSm" bg="bg.glass" borderWidth="1px" borderColor="border">
				<InvitationAvatar invitation={invitation} size={40} icon={<LuMailX size={16} />} />
				<Box minW="0" flex="1" lineHeight="1.3">
					<Text fontSize="14px" color="fg" overflowWrap="anywhere">
						{invitation.email}
					</Text>
					<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
						{sentAndExpiry(invitation, openedAt)}
					</Text>
				</Box>
			</HStack>

			{invitation.assignments.length > 0 && (
				<HStack gap="8px" flexWrap="nowrap" overflow="hidden" aria-label="Roles this invite grants">
					{previewAssignments.map((assignment) => (
						<RolePill key={`${assignment.app_service_id}-${assignment.role_id}`} admin={assignment.role_code.includes("admin")}>
							{assignment.app_code} · {assignment.role_code}
						</RolePill>
					))}
					{hiddenAssignments > 0 && (
						<Text as="span" flex="none" fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
							+{hiddenAssignments}
						</Text>
					)}
				</HStack>
			)}

			{phase === "error" ? (
				<Note tone="red" icon={<LuCircleX size={14} />} title="The invite is still active" alertRef={focusAlert}>
					{failure?.status === null ? "The request did not reach the server. Try again." : failure?.message || `Request failed (${failure?.status}).`}
				</Note>
			) : (
				<FactList>
					<Fact icon={<LuLink2Off size={14} />}>The invite link stops working now. Anyone who opens it sees that it was revoked.</Fact>
					<Fact icon={<LuRotateCw size={14} />}>To invite them again, use New invite. It creates a new link with the same roles.</Fact>
				</FactList>
			)}
		</DestructiveDialog>
	);
};
