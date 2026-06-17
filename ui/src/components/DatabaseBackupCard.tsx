"use client";

import { useEffect, useMemo, useState } from "react";
import {
	Box,
	Center,
	Collapsible,
	HStack,
	Input,
	NativeSelect,
	Spinner,
	Stack,
	Switch,
	Text,
} from "@chakra-ui/react";
import { LuDatabase, LuCheck, LuChevronRight, LuClock, LuCode, LuInfo, LuSave, LuSettings2 } from "react-icons/lu";
import { getDatabaseBackupConfig, updateDatabaseBackupConfig } from "@/apis";
import { Button } from "@/components/ui/button";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { usePermissions } from "@/hooks/usePermissions";
import {
	type CronPreset,
	cronToPreset,
	formatRelative,
	presetSummary,
	presetToCron,
} from "@/utils";

const PANEL_PROPS = {
	bg: "bg.glass",
	borderWidth: "1px",
	borderColor: "border",
	borderRadius: "20px",
	overflow: "hidden",
	css: { backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" },
} as const;

const FIELD_LABEL_PROPS = {
	display: "block",
	fontSize: "13px",
	fontWeight: "medium",
	color: "fg.subtle",
	mb: "8px",
} as const;

const SELECT_FIELD_PROPS = {
	h: "44px",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "14px",
	color: "fg",
	css: {
		backdropFilter: "blur(12px)",
		WebkitBackdropFilter: "blur(12px)",
		"& option": { background: "#12122E", color: "#F4F5FF" },
	},
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none" },
} as const;

const INPUT_WRAP_PROPS = {
	gap: "10px",
	h: "44px",
	px: "14px",
	borderRadius: "glassSm",
	borderWidth: "1px",
	borderColor: "border.strong",
	bg: "bg.glass",
	css: { backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" },
	_focusWithin: { borderColor: "aurora.violet", boxShadow: "focusRing", background: "rgba(255,255,255,0.08)" },
} as const;

const INPUT_INNER_PROPS = {
	flex: "1",
	h: "full",
	border: "0",
	bg: "transparent",
	px: "0",
	fontSize: "14px",
	color: "fg",
	_focusVisible: { boxShadow: "none", outline: "none" },
} as const;

// Default cron 0 3 * * * (03:00 daily) — avoids the rotation-cleanup 0 4 and
// activity-cleanup 0 5 slots.
const DEFAULT_CRON = "0 3 * * *";
const DEFAULT_RETAIN = 10;

// Retain presets offered in the NativeSelect. Values are a COUNT of backup files
// to keep (not a time window). Floor is 1 and ceiling is 90 (backend-enforced).
const RETAIN_OPTIONS = [
	{ value: 3, label: "3 backups" },
	{ value: 7, label: "7 backups" },
	{ value: 10, label: "10 backups" },
	{ value: 14, label: "14 backups" },
	{ value: 30, label: "30 backups" },
	{ value: 90, label: "90 backups" },
] as const;

interface SavedConfig {
	enabled: boolean;
	cron: string;
	retainCount: number;
}

// basename returns the trailing path segment of a backup path (POSIX-style).
const basename = (path: string): string => {
	const segments = path.split("/");
	return segments[segments.length - 1] || path;
};

export const DatabaseBackupCard = () => {
	const { can } = usePermissions();
	const canUpdate = can("settings:update");

	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);
	const [saved, setSaved] = useState<SavedConfig>({ enabled: false, cron: DEFAULT_CRON, retainCount: DEFAULT_RETAIN });
	const [lastRunAt, setLastRunAt] = useState<number | null>(null);
	const [lastBackupPath, setLastBackupPath] = useState<string | null>(null);
	const [lastKeptCount, setLastKeptCount] = useState<number | null>(null);

	// editable state
	const [enabled, setEnabled] = useState(false);
	const [cron, setCron] = useState(DEFAULT_CRON);
	const [retainCount, setRetainCount] = useState(DEFAULT_RETAIN);
	const [preset, setPreset] = useState<CronPreset>("daily");
	const [time, setTime] = useState("03:00");
	const [weekday, setWeekday] = useState("1");
	const [advancedOpen, setAdvancedOpen] = useState(false);
	// Whole-card collapse/expand (header toggles the body + footer). Default collapsed.
	const [cardOpen, setCardOpen] = useState(false);

	// hydrate editable state from a cron string
	const hydrateFromCron = (cronExpr: string) => {
		const parsed = cronToPreset(cronExpr);
		setPreset(parsed.preset);
		setTime(parsed.time);
		setWeekday(parsed.weekday);
		setCron(cronExpr);
	};

	useEffect(() => {
		let active = true;
		(async () => {
			try {
				const config = await getDatabaseBackupConfig();
				if (!active) return;
				setSaved({ enabled: config.enabled, cron: config.cron, retainCount: config.retainCount });
				setEnabled(config.enabled);
				setRetainCount(config.retainCount);
				setLastRunAt(config.lastRunAt);
				setLastBackupPath(config.lastBackupPath);
				setLastKeptCount(config.lastKeptCount);
				hydrateFromCron(config.cron);
			} catch {
				if (active) toaster.create({ title: "Failed to load settings", type: "error", meta: { closable: true } });
			} finally {
				if (active) setLoading(false);
			}
		})();
		return () => {
			active = false;
		};
	}, []);

	const dirty = enabled !== saved.enabled || cron !== saved.cron || retainCount !== saved.retainCount;

	// derived readout
	const summary = useMemo(() => presetSummary(preset, time, weekday), [preset, time, weekday]);

	const applyPreset = (next: CronPreset) => {
		setPreset(next);
		if (next === "custom") return;
		const nextCron = presetToCron(next, time, weekday);
		setCron(nextCron);
	};

	const applyTime = (next: string) => {
		setTime(next);
		if (preset !== "custom") setCron(presetToCron(preset, next, weekday));
	};

	const applyWeekday = (next: string) => {
		setWeekday(next);
		if (preset !== "custom") setCron(presetToCron(preset, time, next));
	};

	// editing the raw cron field re-maps back to a preset (or "custom")
	const applyRawCron = (next: string) => {
		setCron(next);
		const parsed = cronToPreset(next);
		setPreset(parsed.preset);
		if (parsed.preset !== "custom") {
			setTime(parsed.time);
			setWeekday(parsed.weekday);
		}
	};

	const discard = () => {
		setEnabled(saved.enabled);
		setRetainCount(saved.retainCount);
		hydrateFromCron(saved.cron);
	};

	const handleSave = async () => {
		setSaving(true);
		try {
			await updateDatabaseBackupConfig({ enabled, cron, retainCount });
			setSaved({ enabled, cron, retainCount });
			toaster.create({ title: "Settings saved", type: "success", meta: { closable: true } });
		} catch (err) {
			const message =
				(err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
				"Failed to save settings";
			toaster.create({ title: message, type: "error", meta: { closable: true } });
		} finally {
			setSaving(false);
		}
	};

	const needsTime = preset === "daily" || preset === "weekly";
	const needsWeekday = preset === "weekly";
	// Amber chip label in the readout — retain is a COUNT, so "10×".
	const retainLabel = `${retainCount}×`;
	// Relative "Last run 1d ago" label, matching the activity card's run strip.
	const lastRunLabel = lastRunAt ? formatRelative(new Date(lastRunAt * 1000).toISOString()) : null;

	if (loading) {
		return (
			<Box {...PANEL_PROPS}>
				<Center py="14">
					<Spinner size="lg" color="accent" />
				</Center>
			</Box>
		);
	}

	return (
		<Box {...PANEL_PROPS}>
			{/* section head — click to collapse/expand the card */}
			<HStack
				as="button"
				w="full"
				textAlign="left"
				gap="13px"
				px="20px"
				py="18px"
				borderBottomWidth={cardOpen ? "1px" : "0"}
				borderColor="border"
				cursor="pointer"
				aria-expanded={cardOpen}
				onClick={() => setCardOpen((open) => !open)}
				_hover={{ bg: "rgba(255,255,255,0.02)" }}
			>
				<Center
					w="40px"
					h="40px"
					borderRadius="12px"
					flex="none"
					color="aurora.cyan"
					borderWidth="1px"
					borderColor="border.strong"
					css={{
						background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))",
						boxShadow: "0 0 18px rgba(99,102,241,0.18)",
					}}
				>
					<LuDatabase size={19} />
				</Center>
				<Box lineHeight="1.3" minW="0">
					<Text fontSize="16px" fontWeight="semibold" letterSpacing="-0.01em">
						Database backup
					</Text>
					<Text fontSize="12px" color="fg.muted">
						Snapshot the SQLite database on a schedule and keep the N most recent backups.
					</Text>
				</Box>
				<Box
					ml="auto"
					flex="none"
					color="fg.muted"
					css={{ transform: cardOpen ? "rotate(90deg)" : "none", transition: "transform .2s ease" }}
				>
					<LuChevronRight size={18} />
				</Box>
			</HStack>

			{cardOpen && (
			<>
			<Stack gap="22px" p="20px">
				{/* enable/disable switch row */}
				<HStack
					gap="14px"
					px="16px"
					py="14px"
					borderRadius="glassSm"
					borderWidth="1px"
					borderColor="border"
					css={{ background: "rgba(7,7,26,0.35)" }}
				>
					<Box flex="1" minW="0" lineHeight="1.35">
						<Text fontSize="14px" fontWeight="semibold" color="fg">
							Backup cronjob
						</Text>
						<Text fontSize="12px" color="fg.muted">
							When on, the database is snapshotted on the schedule below.
						</Text>
					</Box>
					<Text fontSize="12px" fontWeight="semibold" color={enabled ? "success" : "fg.muted"}>
						{enabled ? "Enabled" : "Disabled"}
					</Text>
					<Switch.Root
						checked={enabled}
						onCheckedChange={(details) => setEnabled(details.checked)}
						disabled={!canUpdate}
						colorPalette="purple"
					>
						<Switch.HiddenInput aria-label="Enable database backup" />
						<Switch.Control>
							<Switch.Thumb />
						</Switch.Control>
					</Switch.Root>
				</HStack>

				{/* schedule + retain block — dims when disabled */}
				<Box
					opacity={enabled ? 1 : 0.42}
					pointerEvents={enabled ? "auto" : "none"}
					css={{ filter: enabled ? "none" : "grayscale(0.2)" }}
					transition="opacity 0.2s"
				>
					{/* schedule grid */}
					<Stack
						direction={{ base: "column", md: "row" }}
						gap="14px"
						align={{ base: "stretch", md: "end" }}
					>
						{/* Run preset */}
						<Box flex="1" minW="0">
							<Text {...FIELD_LABEL_PROPS}>Run</Text>
							<NativeSelect.Root>
								<NativeSelect.Field
									{...SELECT_FIELD_PROPS}
									value={preset}
									onChange={(event) => applyPreset(event.target.value as CronPreset)}
									aria-label="Run schedule preset"
								>
									<option value="hourly">Hourly</option>
									<option value="every6h">Every 6 hours</option>
									<option value="daily">Daily</option>
									<option value="weekly">Weekly</option>
									{preset === "custom" && <option value="custom">Custom (from cron)</option>}
								</NativeSelect.Field>
								<NativeSelect.Indicator color="fg.muted" />
							</NativeSelect.Root>
						</Box>

						{/* Time picker — daily / weekly */}
						{needsTime && (
							<Box w={{ base: "full", md: "168px" }}>
								<Text {...FIELD_LABEL_PROPS}>At</Text>
								<HStack {...INPUT_WRAP_PROPS}>
									<Box color="fg.muted" flex="none">
										<LuClock size={16} />
									</Box>
									<Input
										type="time"
										value={time}
										onChange={(event) => applyTime(event.target.value)}
										aria-label="Run time"
										{...INPUT_INNER_PROPS}
									/>
								</HStack>
							</Box>
						)}

						{/* Weekday picker — weekly only */}
						{needsWeekday && (
							<Box w={{ base: "full", md: "168px" }}>
								<Text {...FIELD_LABEL_PROPS}>On</Text>
								<NativeSelect.Root>
									<NativeSelect.Field
										{...SELECT_FIELD_PROPS}
										value={weekday}
										onChange={(event) => applyWeekday(event.target.value)}
										aria-label="Run weekday"
									>
										<option value="1">Monday</option>
										<option value="2">Tuesday</option>
										<option value="3">Wednesday</option>
										<option value="4">Thursday</option>
										<option value="5">Friday</option>
										<option value="6">Saturday</option>
										<option value="0">Sunday</option>
									</NativeSelect.Field>
									<NativeSelect.Indicator color="fg.muted" />
								</NativeSelect.Root>
							</Box>
						)}

					</Stack>

					{/* Retain count — COUNT-scale presets, on its own row under the Run
					    preset so it stays put when the time/weekday columns toggle. */}
					<Box maxW="220px" mt="14px">
						<Text {...FIELD_LABEL_PROPS}>
							Keep last N backups{" "}
							<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
								· retention
							</Text>
						</Text>
						<NativeSelect.Root>
							<NativeSelect.Field
								{...SELECT_FIELD_PROPS}
								value={String(retainCount)}
								onChange={(event) => setRetainCount(Number(event.target.value))}
								aria-label="Backups to keep"
							>
								{RETAIN_OPTIONS.map((option) => (
									<option key={option.value} value={String(option.value)}>
										{option.label}
									</option>
								))}
							</NativeSelect.Field>
							<NativeSelect.Indicator color="fg.muted" />
						</NativeSelect.Root>
					</Box>

					{/* resolved-cron + retain readout (cron chip + amber keep-N× chip) */}
					<HStack gap="8px" mt="10px" fontSize="12px" color="fg.muted">
						<Text>{summary}</Text>
						<LuChevronRight size={15} />
						<Text
							as="code"
							fontFamily="inherit"
							fontWeight="semibold"
							color="aurora.cyan"
							px="9px"
							py="3px"
							borderRadius="8px"
							css={{
								background: "rgba(34,211,238,0.10)",
								border: "1px solid rgba(34,211,238,0.30)",
								letterSpacing: "0.06em",
							}}
						>
							{cron || "—"}
						</Text>
						<Text as="span">·</Text>
						<HStack as="span" gap="5px">
							<Text as="span">keep</Text>
							<Text
								as="code"
								fontFamily="inherit"
								fontWeight="semibold"
								px="9px"
								py="3px"
								borderRadius="8px"
								css={{
									color: "#F59E0B",
									background: "rgba(245,158,11,0.10)",
									border: "1px solid rgba(245,158,11,0.30)",
									letterSpacing: "0.06em",
								}}
							>
								{retainLabel}
							</Text>
						</HStack>
					</HStack>

					{/* advanced (cron) disclosure */}
					<Collapsible.Root
						open={advancedOpen}
						onOpenChange={(details) => setAdvancedOpen(details.open)}
						mt="18px"
					>
						<Box borderWidth="1px" borderColor="border" borderRadius="glassSm" css={{ background: "rgba(7,7,26,0.30)" }} overflow="hidden">
							<Collapsible.Trigger
								width="100%"
								cursor="pointer"
								css={{ background: "transparent", border: "0", textAlign: "left" }}
							>
								<HStack gap="12px" px="16px" py="13px" color="fg.subtle">
									<Box color="aurora.violet" flex="none">
										<LuSettings2 size={16} />
									</Box>
									<Box flex="1" minW="0">
										<Text fontSize="13px" fontWeight="semibold" color="fg">
											Advanced (cron)
										</Text>
										<Text fontSize="12px" color="fg.muted">
											Set the schedule with a raw cron expression.
										</Text>
									</Box>
									<Box
										color="fg.muted"
										transform={advancedOpen ? "rotate(90deg)" : "rotate(0deg)"}
										transition="transform 0.2s"
									>
										<LuChevronRight size={16} />
									</Box>
								</HStack>
							</Collapsible.Trigger>
							<Collapsible.Content>
								<Box px="16px" pb="16px">
									<Text {...FIELD_LABEL_PROPS}>
										Cron expression{" "}
										<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
											· 5-field
										</Text>
									</Text>
									<HStack {...INPUT_WRAP_PROPS}>
										<Box color="fg.muted" flex="none">
											<LuCode size={16} />
										</Box>
										<Input
											type="text"
											value={cron}
											onChange={(event) => applyRawCron(event.target.value)}
											spellCheck={false}
											aria-label="Cron expression"
											{...INPUT_INNER_PROPS}
											css={{ letterSpacing: "0.06em" }}
										/>
									</HStack>
									<Text fontSize="12px" color="fg.muted" mt="8px">
										5-field cron ·{" "}
										<Text as="code" color="fg.subtle" fontFamily="inherit">
											minute hour day month weekday
										</Text>
										. Editing this sets the schedule directly — the preset above reads{" "}
										<Text as="code" color="fg.subtle" fontFamily="inherit">
											Custom
										</Text>{" "}
										if it no longer matches a preset. Retain count is independent of the cron.
									</Text>
								</Box>
							</Collapsible.Content>
						</Box>
					</Collapsible.Root>
				</Box>

				{/* last-run strip */}
				<HStack
					gap="11px"
					px="14px"
					py="12px"
					borderRadius="glassSm"
					borderWidth="1px"
					borderColor={lastRunLabel ? "border" : "border.strong"}
					bg="bg.glass"
					fontSize="13px"
					color={lastRunLabel ? "fg.subtle" : "fg.muted"}
				>
					<Center
						w="30px"
						h="30px"
						borderRadius="9px"
						flex="none"
						color={lastRunLabel ? "success" : "fg.muted"}
						borderWidth="1px"
						borderColor={lastRunLabel ? "rgba(52,211,153,0.30)" : "border.strong"}
						css={{ background: lastRunLabel ? "rgba(52,211,153,0.12)" : "rgba(255,255,255,0.06)" }}
					>
						{lastRunLabel ? <LuCheck size={16} /> : <LuClock size={16} />}
					</Center>
					{lastRunLabel ? (
						<Text>
							Last run <Text as="b" color="fg" fontWeight="semibold">{lastRunLabel}</Text>
							{lastBackupPath ? (
								<>
									{" "}·{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">{basename(lastBackupPath)}</Text>
								</>
							) : null}{" "}·{" "}
							<Text as="span" color="success" fontWeight="semibold">
								kept {(lastKeptCount ?? 0).toLocaleString()} backup{lastKeptCount === 1 ? "" : "s"}
							</Text>
						</Text>
					) : (
						<Text>Never run — the first backup happens on the next scheduled time.</Text>
					)}
				</HStack>

				{/* informational note */}
				<HStack
					align="flex-start"
					gap="9px"
					px="14px"
					py="11px"
					borderRadius="glassSm"
					borderWidth="1px"
					borderColor="border"
					bg="bg.glass"
					fontSize="12px"
					color="fg.muted"
				>
					<Box color="aurora.cyan" flex="none" mt="1px">
						<LuInfo size={15} />
					</Box>
					<Text>
						Each run writes a consistent snapshot to{" "}
						<Text as="b" color="fg.subtle" fontWeight="semibold">
							db/backups/
						</Text>{" "}
						via SQLite{" "}
						<Text as="b" color="fg.subtle" fontWeight="semibold">
							VACUUM INTO
						</Text>
						. Only the most recent backups are kept — older files past the retain count are pruned.
					</Text>
				</HStack>
			</Stack>

			{/* footer actions */}
			<HStack gap="10px" px="20px" py="16px" borderTopWidth="1px" borderColor="border">
				{dirty && (
					<HStack gap="8px" fontSize="13px" color="aurora.violet" fontWeight="semibold">
						<Box w="7px" h="7px" borderRadius="full" bg="aurora.violet" css={{ boxShadow: "0 0 10px #8B5CF6" }} />
						Unsaved changes
					</HStack>
				)}
				<Box flex="1" />
				<Button
					h="11"
					px="4.5"
					fontSize="sm"
					variant="outline"
					borderRadius="glassSm"
					borderColor="border.strong"
					bg="bg.glass"
					color="fg"
					fontWeight="semibold"
					css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
					_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
					disabled={!dirty}
					onClick={discard}
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
					disabled={!dirty || !canUpdate}
					loading={saving}
					onClick={handleSave}
				>
					<LuSave size={15} /> Save changes
				</Button>
			</HStack>
			</>
			)}
		</Box>
	);
};
