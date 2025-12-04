"use client";

import { Button as ChakraButton, type ButtonProps } from "@chakra-ui/react";
import { forwardRef } from "react";

export interface ButtonComponentProps extends ButtonProps {}

export const Button = forwardRef<HTMLButtonElement, ButtonComponentProps>(function Button(props, ref) {
	const { variant = "primary", ...rest } = props;

	// Map custom variants to Chakra variants
	const chakraVariant = variant === "primary" ? "solid" : variant;

	// Apply color palette for primary/secondary
	const colorPalette = variant === "primary" ? "brand" : undefined;

	return <ChakraButton ref={ref} variant={chakraVariant} colorPalette={colorPalette} {...rest} />;
});
