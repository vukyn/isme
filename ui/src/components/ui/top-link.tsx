import { Text } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";

interface TopLinkProps {
	prompt: string;
	linkText: string;
	to: string;
}

export const TopLink = ({ prompt, linkText, to }: TopLinkProps) => {
	return (
		<Text fontSize="sm" color="fg.muted">
			{prompt}{" "}
			<RouterLink
				to={to}
				style={{
					color: "var(--chakra-colors-fg)",
					fontWeight: 600,
					borderBottom: "1px solid var(--chakra-colors-aurora-violet)",
					paddingBottom: 1,
				}}
			>
				{linkText}
			</RouterLink>
		</Text>
	);
};
