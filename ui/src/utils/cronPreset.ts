/**
 * Bidirectional mapping between a human-friendly schedule preset and a 5-field
 * standard cron string (minute hour day-of-month month day-of-week). Pure
 * functions — mirrors the preset logic in demo/aurora-settings-design.html.
 */

export type CronPreset = "hourly" | "every6h" | "daily" | "weekly" | "custom";

export interface ParsedPreset {
	preset: CronPreset;
	/** "HH:MM" — meaningful for daily / weekly. */
	time: string;
	/** Day-of-week 0-6 (Sunday=0) as a string — meaningful for weekly. */
	weekday: string;
}

const pad2 = (value: number): string => String(value).padStart(2, "0");

const DAILY_RE = /^(\d{1,2}) (\d{1,2}) \* \* \*$/;
const WEEKLY_RE = /^(\d{1,2}) (\d{1,2}) \* \* ([0-6])$/;

/** Builds a 5-field cron string from a preset + time + weekday. */
export const presetToCron = (preset: CronPreset, time: string, weekday: string): string => {
	const [hourPart, minutePart] = (time || "03:00").split(":");
	const minute = String(parseInt(minutePart ?? "0", 10) || 0);
	const hour = String(parseInt(hourPart ?? "0", 10) || 0);
	switch (preset) {
		case "hourly":
			return "0 * * * *";
		case "every6h":
			return "0 */6 * * *";
		case "daily":
			return `${minute} ${hour} * * *`;
		case "weekly":
			return `${minute} ${hour} * * ${weekday}`;
		default:
			// "custom" carries no derivable cron — caller keeps the raw input
			return "";
	}
};

/** Maps a cron string back to a preset (falls back to "custom"). */
export const cronToPreset = (cron: string): ParsedPreset => {
	const trimmed = (cron || "").trim();
	const fallback: ParsedPreset = { preset: "custom", time: "03:00", weekday: "1" };

	if (trimmed === "0 * * * *") {
		return { preset: "hourly", time: "03:00", weekday: "1" };
	}
	if (trimmed === "0 */6 * * *") {
		return { preset: "every6h", time: "03:00", weekday: "1" };
	}

	const weekly = WEEKLY_RE.exec(trimmed);
	if (weekly) {
		const minute = parseInt(weekly[1], 10);
		const hour = parseInt(weekly[2], 10);
		return { preset: "weekly", time: `${pad2(hour)}:${pad2(minute)}`, weekday: weekly[3] };
	}

	const daily = DAILY_RE.exec(trimmed);
	if (daily) {
		const minute = parseInt(daily[1], 10);
		const hour = parseInt(daily[2], 10);
		return { preset: "daily", time: `${pad2(hour)}:${pad2(minute)}`, weekday: "1" };
	}

	return fallback;
};

export const WEEKDAY_NAMES: Record<string, string> = {
	"0": "Sunday",
	"1": "Monday",
	"2": "Tuesday",
	"3": "Wednesday",
	"4": "Thursday",
	"5": "Friday",
	"6": "Saturday",
};

/** Human-readable summary of a preset selection (the readout line). */
export const presetSummary = (preset: CronPreset, time: string, weekday: string): string => {
	switch (preset) {
		case "hourly":
			return "Every hour, on the hour";
		case "every6h":
			return "Every 6 hours";
		case "daily":
			return `Daily at ${time}`;
		case "weekly":
			return `Weekly · ${WEEKDAY_NAMES[weekday] ?? "Monday"} at ${time}`;
		default:
			return "Custom schedule";
	}
};
