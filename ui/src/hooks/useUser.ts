import { useContext } from "react";
import { UserContext, type UserContextValue } from "@/contexts/userContext";

export const useUser = (): UserContextValue => {
	const userContext = useContext(UserContext);
	if (!userContext) {
		throw new Error("useUser must be used within a UserProvider");
	}
	return userContext;
};
