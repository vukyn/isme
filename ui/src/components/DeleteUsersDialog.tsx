"use client";

import { useCallback, useRef, useState } from "react";
import { Box, Button, Collapsible, HStack, Progress, Spinner, Stack, Text } from "@chakra-ui/react";
import {
	LuArchive,
	LuCheck,
	LuChevronRight,
	LuCircleX,
	LuClock,
	LuInfo,
	LuLogIn,
	LuMonitor,
	LuRotateCw,
	LuShield,
	LuTrash2,
	LuTriangleAlert,
} from "react-icons/lu";
import { softDeleteUser } from "@/apis";
import { DestructiveDialog, Fact, FactList, Lead, Note } from "@/components/DestructiveDialog";
import { UserAvatar } from "@/components/UserAvatar";
import { toaster } from "@/components/ui/toaster";
import { DANGER_DIALOG_BUTTON_PROPS, GHOST_DIALOG_BUTTON_PROPS } from "@/consts/styles";
import { type RequestFailure, toRequestFailure } from "@/utils";
import { USER_STATUS } from "@/consts";
import type { UserListItem, UserStatus } from "@/types";

/** Spinner shows only when a request outlasts this, so fast deletes never flash one. */
const SPINNER_DELAY_MS = 150;
/** Lists longer than this get a count header ("Scroll to review all"). */
const LIST_HEAD_THRESHOLD = 6;
const NETWORK_REASON = "Network error: the request did not reach the server.";
const NETWORK_SINGLE = "The request did not reach the server. Check your connection and try again.";

type RowState = "queued" | "running" | "deleted" | "gone" | "failed";

interface RowResult {
	state: RowState;
	failure?: RequestFailure;
}

type Phase = "confirm" | "busy" | "error" | "gone" | "result";

export interface DeleteUsersOutcome {
	/** Ids that no longer exist (deleted now, or already gone). */
	deletedIds: string[];
	/** Ids whose delete failed — they still exist. */
	failedIds: string[];
}

interface DeleteUsersDialogProps {
	open: boolean;
	/** Full rows (not ids) so names, emails and avatars render without a fetch. */
	users: UserListItem[];
	/** The signed-in admin — skipped, the server rejects deleting yourself. */
	currentUserId?: string;
	onOpenChange: (open: boolean) => void;
	/** Called after every run so the page can refetch and fix its selection. */
	onSettled: (outcome: DeleteUsersOutcome, bulk: boolean) => void;
	/** Reversible alternative for a single active user; omitted = not offered. */
	onDeactivate?: (user: UserListItem) => void;
}

const displayName = (user: UserListItem) => user.name || user.email;

const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? "" : "s"}`;

const describeFailure = (failure: RequestFailure | undefined) => {
	if (!failure || failure.status === null) return NETWORK_REASON;
	if (failure.status >= 500) return `Server error (${failure.status}). Retry to check whether the account is gone.`;
	return failure.message || `Request failed (${failure.status}).`;
};

const isRetryable = (result: RowResult | undefined) => result?.state === "failed" && result.failure?.status !== 403;

const STATUS_PILL: Record<UserStatus, { label: string; color: string; border: string; bg: string }> = {
	[USER_STATUS.ACTIVE]: { label: "Active", color: "aurora.mint", border: "rgba(52,211,153,0.35)", bg: "rgba(52,211,153,0.10)" },
	[USER_STATUS.INACTIVE]: { label: "Inactive", color: "fg.muted", border: "border.strong", bg: "bg.glass" },
	[USER_STATUS.PENDING]: { label: "Pending", color: "aurora.amber", border: "rgba(245,158,11,0.35)", bg: "rgba(245,158,11,0.10)" },
};

const Pill = ({ color, border, bg, dot, children }: { color: string; border: string; bg: string; dot?: boolean; children: React.ReactNode }) => (
	<HStack
		as="span"
		flex="none"
		gap="6px"
		px="10px"
		py="5px"
		borderRadius="full"
		borderWidth="1px"
		borderColor={border}
		bg={bg}
		color={color}
		fontSize="12px"
		fontWeight="medium"
		lineHeight="1"
		whiteSpace="nowrap"
	>
		{dot && <Box w="6px" h="6px" borderRadius="full" bg="currentColor" />}
		{children}
	</HStack>
);

const StatusPill = ({ status }: { status: UserStatus }) => {
	const style = STATUS_PILL[status] ?? STATUS_PILL[USER_STATUS.INACTIVE];
	return (
		<Pill color={style.color} border={style.border} bg={style.bg} dot>
			{style.label}
		</Pill>
	);
};

const YouPill = ({ children }: { children: React.ReactNode }) => (
	<Pill color="aurora.violet" border="rgba(139,92,246,0.40)" bg="rgba(139,92,246,0.12)">
		{children}
	</Pill>
);

/** Name line + email line; a nameless user shows the email as the title, wrapped. */
const UserMeta = ({ user, compact, struck, reason, reasonMuted }: { user: UserListItem; compact?: boolean; struck?: boolean; reason?: string; reasonMuted?: boolean }) => (
	<Box minW="0" flex="1" lineHeight="1.3">
		<Text
			fontSize={compact ? "13px" : "14px"}
			fontWeight={user.name ? "semibold" : "normal"}
			color="fg"
			overflowWrap="anywhere"
			textDecoration={struck ? "line-through" : undefined}
		>
			{displayName(user)}
		</Text>
		{user.name && (
			<Text fontSize="12px" color="fg.muted" overflowWrap="anywhere">
				{user.email}
			</Text>
		)}
		{reason && (
			<Text fontSize="12px" color={reasonMuted ? "fg.muted" : "#FBC6DE"} mt="2px">
				{reason}
			</Text>
		)}
	</Box>
);

const RowStatusIcon = ({ state }: { state: RowState }) => {
	const labels: Record<RowState, string> = { queued: "Waiting", running: "Deleting", deleted: "Deleted", gone: "Deleted", failed: "Failed" };
	return (
		<Box w="20px" h="20px" flex="none" display="grid" placeItems="center" aria-label={labels[state]} role="img">
			{state === "queued" && <Box w="6px" h="6px" borderRadius="full" bg="fg.muted" />}
			{state === "running" && <Spinner size="xs" color="aurora.cyan" />}
			{(state === "deleted" || state === "gone") && (
				<Box color="aurora.mint">
					<LuCheck size={14} />
				</Box>
			)}
			{state === "failed" && (
				<Box color="aurora.magenta">
					<LuCircleX size={14} />
				</Box>
			)}
		</Box>
	);
};

/** Mock .ulist — bordered, scrolls inside the body, keyboard-scrollable. */
const UserList = ({ label, children }: { label: string; children: React.ReactNode }) => (
	<Box
		as="ul"
		aria-label={label}
		tabIndex={0}
		m="0"
		p="4px"
		listStyleType="none"
		borderRadius="glassSm"
		borderWidth="1px"
		borderColor="border"
		bg="rgba(7,7,26,0.45)"
		maxH="264px"
		overflowY="auto"
		outline="none"
		_focusVisible={{ boxShadow: "0 0 0 3px rgba(139,92,246,0.45)" }}
		css={{
			overscrollBehavior: "contain",
			scrollbarWidth: "thin",
			scrollbarColor: "rgba(255,255,255,0.18) transparent",
			"& > li + li": { borderTop: "1px solid rgba(255,255,255,0.05)" },
		}}
	>
		{children}
	</Box>
);

const UserRow = ({ dimmed, children }: { dimmed?: boolean; children: React.ReactNode }) => (
	<HStack as="li" gap="10px" minW="0" px="10px" py="8px" borderRadius="9px" opacity={dimmed ? 0.6 : 1}>
		{children}
	</HStack>
);

const ListHead = ({ left, right }: { left: React.ReactNode; right: React.ReactNode }) => (
	<HStack justify="space-between" gap="10px" fontSize="12px" color="fg.muted" lineHeight="1" css={{ "& b": { color: "var(--chakra-colors-fg-subtle)", fontWeight: 600 } }}>
		<Text as="span">{left}</Text>
		<Text as="span" css={{ fontVariantNumeric: "tabular-nums" }}>
			{right}
		</Text>
	</HStack>
);

/**
 * Delete confirmation for one or many users (aurora mock A1-A6, B1-B7).
 * Nothing is sent until the admin confirms; the signed-in admin is never
 * sent (the server rejects self-delete) and is shown as skipped instead.
 * Mount it with a fresh `key` per open so every run starts from confirm.
 */
export const DeleteUsersDialog = ({ open, users, currentUserId, onOpenChange, onSettled, onDeactivate }: DeleteUsersDialogProps) => {
	const cancelRef = useRef<HTMLButtonElement>(null);
	const closeRef = useRef<HTMLButtonElement>(null);
	const [phase, setPhase] = useState<Phase>("confirm");
	const [showSpinner, setShowSpinner] = useState(false);
	const [results, setResults] = useState<Record<string, RowResult>>({});
	const [singleFailure, setSingleFailure] = useState<RequestFailure | null>(null);

	const self = users.find((user) => user.id === currentUserId);
	const targets = users.filter((user) => user.id !== currentUserId);
	const bulk = users.length > 1;
	const busy = phase === "busy";
	const sessions = targets.reduce((sum, user) => sum + user.sessions_count, 0);

	// focus the error the moment it renders, so a screen reader announces it
	const focusAlert = useCallback((node: HTMLDivElement | null) => node?.focus(), []);

	const close = () => {
		if (!busy) onOpenChange(false);
	};

	const runSingle = async (user: UserListItem) => {
		setPhase("busy");
		setSingleFailure(null);
		const timer = setTimeout(() => setShowSpinner(true), SPINNER_DELAY_MS);
		try {
			await softDeleteUser(user.id);
			onSettled({ deletedIds: [user.id], failedIds: [] }, false);
			toaster.create({ title: `${displayName(user)} deleted`, type: "success", meta: { closable: true } });
			onOpenChange(false);
		} catch (error: unknown) {
			const failure = toRequestFailure(error);
			if (failure.status === 404) {
				// someone else deleted them (or an earlier try half-succeeded): the outcome is already true
				onSettled({ deletedIds: [user.id], failedIds: [] }, false);
				setPhase("gone");
			} else {
				setSingleFailure(failure);
				setPhase("error");
			}
		} finally {
			clearTimeout(timer);
			setShowSpinner(false);
		}
	};

	const runBulk = async (ids: string[]) => {
		setPhase("busy");
		setResults((previous) => {
			const next = { ...previous };
			for (const id of ids) next[id] = { state: "running" };
			return next;
		});
		const settled = await Promise.allSettled(
			ids.map((id) =>
				softDeleteUser(id).then(
					() => {
						setResults((previous) => ({ ...previous, [id]: { state: "deleted" } }));
						return "deleted" as const;
					},
					(error: unknown) => {
						const failure = toRequestFailure(error);
						const result: RowResult = failure.status === 404 ? { state: "gone" } : { state: "failed", failure };
						setResults((previous) => ({ ...previous, [id]: result }));
						if (result.state === "failed") throw failure;
						return "gone" as const;
					},
				),
			),
		);

		const runDeleted = ids.filter((_, index) => settled[index].status === "fulfilled");
		const runFailed = ids.filter((_, index) => settled[index].status === "rejected");
		// earlier runs' successes stay deleted; only this run's failures remain
		const allDeleted = targets
			.map((user) => user.id)
			.filter((id) => runDeleted.includes(id) || (!ids.includes(id) && results[id] && results[id].state !== "failed"));
		onSettled({ deletedIds: runDeleted, failedIds: runFailed }, true);

		if (runFailed.length === 0) {
			const total = allDeleted.length;
			toaster.create({
				title: `${plural(total, "user")} deleted`,
				description: sessions > 0 ? "Their sessions have ended." : undefined,
				type: "success",
				meta: { closable: true },
			});
			onOpenChange(false);
			return;
		}
		setPhase("result");
	};

	const handleConfirm = () => {
		if (targets.length === 0) return;
		if (bulk) {
			setResults(Object.fromEntries(targets.map((user) => [user.id, { state: "queued" as RowState }])));
			void runBulk(targets.map((user) => user.id));
		} else {
			void runSingle(targets[0]);
		}
	};

	const failedUsers = targets.filter((user) => results[user.id]?.state === "failed");
	const retryIds = failedUsers.filter((user) => isRetryable(results[user.id])).map((user) => user.id);
	const doneUsers = targets.filter((user) => results[user.id]?.state === "deleted" || results[user.id]?.state === "gone");
	const goneCount = targets.filter((user) => results[user.id]?.state === "gone").length;
	const settledCount = targets.filter((user) => {
		const state = results[user.id]?.state;
		return state === "deleted" || state === "gone" || state === "failed";
	}).length;

	const cancelButton = (label = "Cancel", focusRef = cancelRef) => (
		<Button ref={focusRef} {...GHOST_DIALOG_BUTTON_PROPS} disabled={busy} onClick={close}>
			{label}
		</Button>
	);

	// ---------- single user ----------
	if (!bulk && targets.length > 0) {
		const user = targets[0];

		if (phase === "gone") {
			return (
				<DestructiveDialog
					open={open}
					onClose={close}
					busy={false}
					tone="neutral"
					icon={<LuCheck size={18} />}
					iconTone="mint"
					title="Already deleted"
					initialFocusRef={closeRef}
					footer={cancelButton("Close", closeRef)}
				>
					<Lead>
						<b>{displayName(user)}</b> had already been deleted, possibly by another admin. The list has been refreshed.
					</Lead>
				</DestructiveDialog>
			);
		}

		const canRetry = phase === "error" && singleFailure?.status !== 403;
		return (
			<DestructiveDialog
				open={open}
				onClose={close}
				busy={busy}
				tone="danger"
				icon={<LuTrash2 size={18} />}
				iconTone="magenta"
				title="Delete user?"
				initialFocusRef={cancelRef}
				footer={
					<>
						{cancelButton()}
						{(phase !== "error" || canRetry) && (
							<Button
								{...DANGER_DIALOG_BUTTON_PROPS}
								disabled={busy}
								loading={busy && showSpinner}
								loadingText="Deleting…"
								onClick={handleConfirm}
							>
								{phase === "error" ? <LuRotateCw size={15} /> : <LuTrash2 size={15} />}
								{phase === "error" ? "Try again" : "Delete user"}
							</Button>
						)}
					</>
				}
			>
				<HStack gap="12px" minW="0" px="14px" py="12px" borderRadius="glassSm" bg="bg.glass" borderWidth="1px" borderColor="border">
					<UserAvatar user={user} size={40} />
					<UserMeta user={user} />
					<StatusPill status={user.status} />
				</HStack>

				{phase === "error" ? (
					<Note tone="red" icon={<LuCircleX size={14} />} title={`${displayName(user)} was not deleted`} alertRef={focusAlert}>
						{singleFailure?.status === null ? NETWORK_SINGLE : singleFailure?.message || describeFailure(singleFailure ?? undefined)}
					</Note>
				) : (
					<FactList>
						<Fact icon={<LuLogIn size={14} />}>They can no longer sign in to isme or to any app that signs in through it.</Fact>
						{user.sessions_count > 0 && (
							<>
								<Fact icon={<LuMonitor size={14} />}>
									Their <b>{user.sessions_count}</b> active session{user.sessions_count === 1 ? " ends" : "s end"} now.
								</Fact>
								<Fact icon={<LuClock size={14} />}>
									Apps that check tokens on their own may still accept one they already issued, until it expires (up to 15 min).
								</Fact>
							</>
						)}
						<Fact icon={<LuArchive size={14} />}>
							The account leaves this list. Its record is kept, but there is no way to restore it from the console.
						</Fact>
					</FactList>
				)}

				{phase === "confirm" && user.status === USER_STATUS.ACTIVE && onDeactivate && (
					<Note tone="amber" icon={<LuInfo size={14} />}>
						Might you need this account back?{" "}
						<Button
							variant="plain"
							h="auto"
							minW="auto"
							p="0"
							fontSize="12px"
							fontWeight="semibold"
							color="aurora.amber"
							textDecoration="underline"
							textUnderlineOffset="2px"
							verticalAlign="baseline"
							_hover={{ color: "#FBBF24" }}
							onClick={() => {
								onOpenChange(false);
								onDeactivate(user);
							}}
						>
							Deactivate instead
						</Button>
						. It blocks sign-in and can be undone.
					</Note>
				)}
			</DestructiveDialog>
		);
	}

	// ---------- only yourself selected (the row and bulk-bar actions are disabled for this; fallback) ----------
	if (targets.length === 0) {
		return (
			<DestructiveDialog
				open={open}
				onClose={close}
				busy={false}
				tone="neutral"
				icon={<LuShield size={18} />}
				iconTone="amber"
				title="Nothing to delete"
				initialFocusRef={closeRef}
				footer={
					<>
						{cancelButton("Close", closeRef)}
						<Button {...DANGER_DIALOG_BUTTON_PROPS} disabled>
							<LuTrash2 size={15} /> Delete
						</Button>
					</>
				}
			>
				<Lead>
					The only selected account is <b>yours</b>. You can't delete your own account; ask another admin if it has to go.
				</Lead>
			</DestructiveDialog>
		);
	}

	// ---------- bulk: result (partial or none deleted) ----------
	if (phase === "result") {
		const noneDeleted = doneUsers.length === 0;
		const reasons = new Set(failedUsers.map((user) => describeFailure(results[user.id]?.failure)));
		const sharedReason = reasons.size === 1 ? [...reasons][0] : null;
		return (
			<DestructiveDialog
				open={open}
				onClose={close}
				busy={false}
				tone="danger"
				wide
				icon={noneDeleted ? <LuCircleX size={18} /> : <LuTriangleAlert size={18} />}
				iconTone={noneDeleted ? "magenta" : "amber"}
				title={
					noneDeleted ? (
						"No users were deleted"
					) : (
						<>
							{failedUsers.length} of {targets.length} users were not deleted
						</>
					)
				}
				initialFocusRef={closeRef}
				footer={
					<>
						{cancelButton("Close", closeRef)}
						{retryIds.length > 0 && (
							<Button {...DANGER_DIALOG_BUTTON_PROPS} onClick={() => void runBulk(retryIds)}>
								<LuRotateCw size={15} /> Retry {retryIds.length}
							</Button>
						)}
					</>
				}
			>
				<Note tone="red" icon={<LuCircleX size={14} />} alertRef={focusAlert}>
					{noneDeleted ? (
						<>
							All <b>{failedUsers.length}</b> requests failed.{" "}
							{sharedReason ?? "See the reason on each row."} Your selection is unchanged.
						</>
					) : (
						<>
							<b>{doneUsers.length}</b> deleted. The <b>{failedUsers.length}</b> below still exist and are still selected. Retry them,
							or close and look into them.
						</>
					)}
				</Note>

				{!noneDeleted && <ListHead left={<b>Not deleted</b>} right={failedUsers.length} />}
				<UserList label="Users not deleted">
					{failedUsers.map((user) => (
						<UserRow key={user.id}>
							<RowStatusIcon state="failed" />
							<UserMeta
								user={user}
								compact
								reason={noneDeleted && sharedReason ? undefined : describeFailure(results[user.id]?.failure)}
							/>
						</UserRow>
					))}
				</UserList>

				{!noneDeleted && (
					<Collapsible.Root>
						<Collapsible.Trigger
							display="inline-flex"
							alignItems="center"
							gap="6px"
							fontSize="12px"
							color="fg.muted"
							cursor="pointer"
							lineHeight="1"
							borderRadius="4px"
							_focusVisible={{ outline: "2px solid", outlineColor: "aurora.violet", outlineOffset: "2px" }}
							css={{ "&[data-state=open] .chevron": { transform: "rotate(90deg)" } }}
						>
							<Box color="aurora.mint">
								<LuCheck size={13} />
							</Box>
							<Text as="span" css={{ fontVariantNumeric: "tabular-nums" }}>
								{doneUsers.length} deleted{goneCount > 0 ? ` (${goneCount} ${goneCount === 1 ? "was" : "were"} already gone)` : ""}.
							</Text>
							<Text as="span" color="fg.subtle">
								Show
							</Text>
							<Box className="chevron" transition="transform .2s ease">
								<LuChevronRight size={12} />
							</Box>
						</Collapsible.Trigger>
						<Collapsible.Content pt="8px">
							<UserList label="Users deleted">
								{doneUsers.map((user) => (
									<UserRow key={user.id}>
										<RowStatusIcon state="deleted" />
										<UserMeta
											user={user}
											compact
											reason={results[user.id]?.state === "gone" ? "Already deleted by someone else." : undefined}
											reasonMuted
										/>
									</UserRow>
								))}
							</UserList>
						</Collapsible.Content>
					</Collapsible.Root>
				)}
			</DestructiveDialog>
		);
	}

	// ---------- bulk: in progress ----------
	if (busy) {
		const percent = Math.round((settledCount / targets.length) * 100);
		return (
			<DestructiveDialog
				open={open}
				onClose={close}
				busy
				tone="danger"
				wide
				icon={<LuTrash2 size={18} />}
				iconTone="magenta"
				title={<>Deleting {targets.length} users…</>}
				initialFocusRef={cancelRef}
				footer={
					<>
						<Text mr="auto" fontSize="12px" color="fg.muted">
							Keep this open until it finishes.
						</Text>
						<Button {...DANGER_DIALOG_BUTTON_PROPS} loading loadingText="Deleting…" />
					</>
				}
			>
				<Stack gap="8px">
					<HStack justify="space-between" fontSize="12px" color="fg.subtle" lineHeight="1" css={{ fontVariantNumeric: "tabular-nums" }}>
						<Text as="span" aria-live="polite">
							{settledCount} of {targets.length} done
						</Text>
						<Text as="span">{percent}%</Text>
					</HStack>
					<Progress.Root value={settledCount} max={targets.length} size="xs" aria-label="Delete progress">
						<Progress.Track h="6px" borderRadius="full" bg="rgba(255,255,255,0.08)">
							<Progress.Range bg="aurora.cyan" borderRadius="full" />
						</Progress.Track>
					</Progress.Root>
				</Stack>
				<UserList label="Delete progress by user">
					{targets.map((user) => (
						<UserRow key={user.id}>
							<RowStatusIcon state={results[user.id]?.state ?? "queued"} />
							<UserMeta user={user} compact />
						</UserRow>
					))}
				</UserList>
			</DestructiveDialog>
		);
	}

	// ---------- bulk: confirm ----------
	return (
		<DestructiveDialog
			open={open}
			onClose={close}
			busy={false}
			tone="danger"
			wide
			icon={<LuTrash2 size={18} />}
			iconTone="magenta"
			title={
				<>
					Delete {targets.length} user{targets.length === 1 ? "" : "s"}?
				</>
			}
			initialFocusRef={cancelRef}
			footer={
				<>
					{cancelButton()}
					<Button {...DANGER_DIALOG_BUTTON_PROPS} onClick={handleConfirm}>
						<LuTrash2 size={15} /> Delete {plural(targets.length, "user")}
					</Button>
				</>
			}
		>
			{self && (
				<Note tone="violet" icon={<LuShield size={14} />}>
					Your own account is selected and will be skipped. You can't delete yourself.
				</Note>
			)}
			{users.length > LIST_HEAD_THRESHOLD && <ListHead left={<><b>{users.length}</b> users</>} right="Scroll to review all" />}
			<UserList label="Users to delete">
				{users.map((user) => {
					const skipped = user.id === currentUserId;
					return (
						<UserRow key={user.id} dimmed={skipped}>
							<UserAvatar user={user} size={28} />
							<UserMeta user={user} compact struck={skipped} />
							{skipped ? <YouPill>You, skipped</YouPill> : <StatusPill status={user.status} />}
						</UserRow>
					);
				})}
			</UserList>
			<FactList>
				{self ? (
					<Fact icon={<LuLogIn size={14} />}>
						None of the {targets.length} can sign in any more
						{sessions > 0 ? (
							<>
								, and their <b>{sessions}</b> active session{sessions === 1 ? " ends" : "s end"} now.
							</>
						) : (
							"."
						)}
					</Fact>
				) : (
					<>
						<Fact icon={<LuLogIn size={14} />}>None of them can sign in any more.</Fact>
						{sessions > 0 && (
							<Fact icon={<LuMonitor size={14} />}>
								<b>{sessions}</b> active session{sessions === 1 ? " ends" : "s end"} now. Tokens already issued may work in other apps for
								up to 15 min.
							</Fact>
						)}
					</>
				)}
				<Fact icon={<LuArchive size={14} />}>No way to restore them from the console.</Fact>
			</FactList>
		</DestructiveDialog>
	);
};
