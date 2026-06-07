"use client";

import { Button as ChakraButton, type ButtonProps } from "@chakra-ui/react";
import { forwardRef } from "react";

export type ButtonComponentProps = ButtonProps;

export const Button = forwardRef<HTMLButtonElement, ButtonComponentProps>(function Button(props, ref) {
	return <ChakraButton ref={ref} {...props} />;
});
