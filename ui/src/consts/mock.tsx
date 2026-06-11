import type { ReactNode } from "react";
import { LuMonitor, LuClock, LuCalendar } from "react-icons/lu";
import type { StatTone, StatDeltaTone } from "@/components/ui/stat-card";

export interface StatEntry {
	tone: StatTone;
	icon: ReactNode;
	title: string;
	desc: string;
	stat: string;
	delta?: string;
	deltaTone?: StatDeltaTone;
}

export const MOCK_STATS: StatEntry[] = [
	{
		tone: "cyan",
		icon: <LuMonitor />,
		title: "Active sessions",
		desc: "Devices currently signed in.",
		stat: "3",
		delta: "▲ +1 since yesterday",
	},
	{
		tone: "violet",
		icon: <LuClock />,
		title: "Token rotations",
		desc: "Refreshes in last 24h.",
		stat: "128",
		// Informational (relative last-refreshed time), not a positive trend → muted.
		delta: "↻ last refreshed 3m ago",
		deltaTone: "neutral",
	},
	{
		tone: "magenta",
		icon: <LuCalendar />,
		title: "Member since",
		desc: "When you joined.",
		stat: "—",
		delta: "● account age",
		deltaTone: "neutral",
	},
];
