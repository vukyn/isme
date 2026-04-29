import { Field, IconButton, Input, InputGroup } from "@chakra-ui/react";
import { useState } from "react";
import { LuEye, LuEyeOff, LuLock } from "react-icons/lu";

interface PasswordFieldProps {
	label: string;
	value: string;
	onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
	error?: string;
	autoComplete?: "current-password" | "new-password";
	name?: string;
	placeholder?: string;
}

export const PasswordField = ({
	label,
	value,
	onChange,
	error,
	autoComplete = "current-password",
	name = "password",
	placeholder = "Enter password",
}: PasswordFieldProps) => {
	const [show, setShow] = useState(false);

	return (
		<Field.Root invalid={!!error}>
			<Field.Label>{label}</Field.Label>
			<InputGroup
				startElement={<LuLock />}
				endElement={
					<IconButton
						aria-label={show ? "Hide password" : "Show password"}
						variant="ghost"
						size="xs"
						onClick={() => setShow((s) => !s)}
						color="fg.muted"
						_hover={{ color: "fg" }}
					>
						{show ? <LuEyeOff /> : <LuEye />}
					</IconButton>
				}
			>
				<Input
					name={name}
					type={show ? "text" : "password"}
					autoComplete={autoComplete}
					placeholder={placeholder}
					value={value}
					onChange={onChange}
				/>
			</InputGroup>
			{error && <Field.ErrorText>{error}</Field.ErrorText>}
		</Field.Root>
	);
};
