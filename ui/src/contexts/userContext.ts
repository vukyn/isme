import { createContext } from "react";
import type { GetMeResponse } from "@/types";

export interface UserContextValue {
	user: GetMeResponse | null;
	loading: boolean;
	error: Error | null;
	refetch: () => Promise<GetMeResponse | null>;
	clear: () => void;
}

export const UserContext = createContext<UserContextValue | null>(null);
