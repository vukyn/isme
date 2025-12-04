import { Input as ChakraInput, InputGroup, type InputProps } from "@chakra-ui/react";
import { forwardRef } from "react";

export interface InputComponentProps extends InputProps {
	startElement?: React.ReactNode;
	endElement?: React.ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputComponentProps>(function Input(props, ref) {
	const { startElement, endElement, ...rest } = props;

	if (startElement || endElement) {
		return (
			<InputGroup startElement={startElement} endElement={endElement}>
				<ChakraInput ref={ref} {...rest} />
			</InputGroup>
		);
	}

	return <ChakraInput ref={ref} {...rest} />;
});
