import { Checkbox as ChakraCheckbox, type CheckboxRootProps } from "@chakra-ui/react";

export interface CheckboxComponentProps extends CheckboxRootProps {
	label?: string;
}

export const Checkbox = (props: CheckboxComponentProps) => {
	const { label, children, ...rest } = props;

	return (
		<ChakraCheckbox.Root {...rest}>
			<ChakraCheckbox.HiddenInput />
			<ChakraCheckbox.Control />
			{(label || children) && <ChakraCheckbox.Label>{label || children}</ChakraCheckbox.Label>}
		</ChakraCheckbox.Root>
	);
};
