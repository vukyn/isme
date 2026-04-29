import { Box, HStack } from "@chakra-ui/react";

interface PasswordStrengthProps {
	value: string;
}

const score = (pw: string): number => {
	let s = 0;
	if (pw.length >= 8) s++;
	if (/[A-Z]/.test(pw)) s++;
	if (/[0-9]/.test(pw)) s++;
	if (/[^A-Za-z0-9]/.test(pw)) s++;
	return s;
};

export const PasswordStrength = ({ value }: PasswordStrengthProps) => {
	const lit = score(value);

	return (
		<HStack gap="1.5" mt="2.5" aria-hidden="true">
			{[0, 1, 2, 3].map((i) => (
				<Box
					key={i}
					flex="1"
					h="1"
					borderRadius="full"
					bg={i < lit ? undefined : "bg.glassHi"}
					css={
						i < lit
							? {
									background: "linear-gradient(90deg, #22D3EE, #8B5CF6)",
									boxShadow: "0 0 8px rgba(139,92,246,0.5)",
							  }
							: undefined
					}
				/>
			))}
		</HStack>
	);
};
