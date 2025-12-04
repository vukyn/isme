"use client";

import { Text, type TextProps } from "@chakra-ui/react";
import { Link as RouterLink } from "react-router-dom";
import { forwardRef } from "react";

export interface LinkComponentProps extends TextProps {
	to: string;
	children: React.ReactNode;
}

export const Link = forwardRef<HTMLAnchorElement, LinkComponentProps>(function Link(props, ref) {
	const { to, children, ...textProps } = props;

	return (
		<Text as="span" {...textProps}>
			<RouterLink
				to={to}
				ref={ref}
				style={{
					textDecoration: "none",
					color: "inherit",
				}}
			>
				{children}
			</RouterLink>
		</Text>
	);
});
