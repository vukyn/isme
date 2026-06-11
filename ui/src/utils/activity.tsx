import type { ReactNode } from "react";
import { LuCheck, LuKey, LuLogOut, LuUserPlus } from "react-icons/lu";
import type { ActivityTone } from "@/components/ui/activity-row";
import type { ActivityItem } from "@/types";
import { formatRelative } from "@/utils/time";

export interface ActivityRowData {
	tone: ActivityTone;
	icon: ReactNode;
	body: ReactNode;
	time: string;
}

/** Maps a structured activity record to its display row. The server sends only
 *  {type, meta}; the copy/icon/tone are composed here from the type. Shared by
 *  the Welcome feed and the full Activity page. */
export const activityToRow = (item: ActivityItem): ActivityRowData => {
	const time = item.created_at ? formatRelative(item.created_at) : "";
	switch (item.type) {
		case "sign_in": {
			const device = typeof item.meta.device === "string" ? item.meta.device : "this device";
			const clientIp = typeof item.meta.client_ip === "string" ? item.meta.client_ip : "";
			return {
				tone: "ok",
				icon: <LuCheck />,
				body: (
					<>
						Sign-in from <b>{device}</b>
						{clientIp ? ` · ${clientIp}` : ""}
					</>
				),
				time,
			};
		}
		case "invitation_sent": {
			const email = typeof item.meta.email === "string" ? item.meta.email : "someone";
			const roles = Array.isArray(item.meta.roles) ? (item.meta.roles as unknown[]).filter((r): r is string => typeof r === "string") : [];
			return {
				tone: "magenta",
				icon: <LuUserPlus />,
				body: (
					<>
						Invited <b>{email}</b>
						{roles.length > 0 ? ` as ${roles.join(", ")}` : ""}
					</>
				),
				time,
			};
		}
		case "password_changed":
			return { tone: "violet", icon: <LuKey />, body: <>Password changed</>, time };
		case "sign_out":
			return { tone: "violet", icon: <LuLogOut />, body: <>Signed out</>, time };
		default:
			return { tone: "violet", icon: <LuCheck />, body: <>{item.type}</>, time };
	}
};
