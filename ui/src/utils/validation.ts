import { z } from "zod";

/**
 * Password validation schema
 * Min 8 characters, at least 1 uppercase, 1 lowercase, 1 number, 1 special character
 */
const passwordSchema = z
	.string()
	.min(6, "Password must be at least 6 characters")
	.regex(/[A-Z]/, "Password must contain at least one uppercase letter")
	.regex(/[a-z]/, "Password must contain at least one lowercase letter")
	.regex(/[0-9]/, "Password must contain at least one number")
	.regex(/[^A-Za-z0-9]/, "Password must contain at least one special character");

/**
 * Simple password validation for login (only check not empty)
 */
const loginPasswordSchema = z.string().min(1, "Password is required");

/**
 * Login form schema
 */
export const loginSchema = z.object({
	email: z.email("Invalid email address"),
	password: loginPasswordSchema,
});

export type LoginFormData = z.infer<typeof loginSchema>;

/**
 * Signup form schema
 */
export const signupSchema = z.object({
	name: z.string().min(2, "Name must be at least 2 characters"),
	email: z.email("Invalid email address"),
	password: passwordSchema,
});

export type SignupFormData = z.infer<typeof signupSchema>;
