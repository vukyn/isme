import type { ReactNode } from "react";
import { LuMonitor, LuClock, LuCalendar, LuCheck, LuKey, LuUserPlus } from "react-icons/lu";
import type { StatTone, StatDeltaTone } from "@/components/ui/stat-card";
import type { ActivityTone } from "@/components/ui/activity-row";

export interface StatEntry {
	tone: StatTone;
	icon: ReactNode;
	title: string;
	desc: string;
	stat: string;
	delta?: string;
	deltaTone?: StatDeltaTone;
}

export interface ActivityEntry {
	tone: ActivityTone;
	icon: ReactNode;
	body: ReactNode;
	time: string;
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

export const MOCK_ACTIVITY: ActivityEntry[] = [
	{
		tone: "ok",
		icon: <LuCheck />,
		body: (
			<>
				Sign-in from <b>MacBook · Safari</b> · Hồ Chí Minh
			</>
		),
		time: "just now",
	},
	{
		tone: "violet",
		icon: <LuKey />,
		body: (
			<>
				API key rotated for <b>billing-service</b>
			</>
		),
		time: "2h ago",
	},
	{
		tone: "magenta",
		icon: <LuUserPlus />,
		body: (
			<>
				Invited <b>thanhlp3@example.com</b> as Admin
			</>
		),
		time: "yesterday",
	},
];
