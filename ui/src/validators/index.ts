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

/**
 * Invite user form schema (Users management)
 * Mirrors user_invitation models.CreateRequest — email + role only; the
 * invitee picks their own name and password on the accept page.
 */
export const inviteUserSchema = z.object({
	email: z.email("Invalid email address"),
	role_id: z.string().min(1, "Role is required"),
});

export type InviteUserFormData = z.infer<typeof inviteUserSchema>;

/**
 * Accept invite form schema (public /accept-invite page)
 * Mirrors user_invitation models.AcceptRequest — password min 6, no
 * complexity rules beyond the backend's.
 */
export const acceptInviteSchema = z.object({
	name: z.string().min(1, "Name is required"),
	password: z.string().min(6, "Password must be at least 6 characters"),
});

export type AcceptInviteFormData = z.infer<typeof acceptInviteSchema>;

/**
 * ctx_info fixed set — mirrors app_service constants.AllowedCtxInfos.
 */
const appServiceCtxInfoSchema = z.enum(["authen", "app_service"], "Context must be authen or app_service");

/**
 * Register app service form schema (App Services management)
 * Mirrors models.RegisterRequest.Validate() — all four fields required.
 */
export const registerAppServiceSchema = z.object({
	app_code: z.string().min(1, "App code is required"),
	app_name: z.string().min(1, "App name is required"),
	redirect_url: z.url("Redirect URL must be a valid URL"),
	ctx_info: appServiceCtxInfoSchema,
});

export type RegisterAppServiceFormData = z.infer<typeof registerAppServiceSchema>;

/**
 * Rotate secret form schema — the refresh API requires the CURRENT secret.
 */
export const rotateAppServiceSecretSchema = z.object({
	app_secret: z.string().min(1, "Current secret is required"),
});

export type RotateAppServiceSecretFormData = z.infer<typeof rotateAppServiceSecretSchema>;

/**
 * Verify credentials form schema — mirrors models.VerifyRequest.Validate().
 */
export const verifyAppServiceSchema = z.object({
	app_code: z.string().min(1, "App code is required"),
	ctx_info: appServiceCtxInfoSchema,
	app_secret: z.string().min(1, "App secret is required"),
});

export type VerifyAppServiceFormData = z.infer<typeof verifyAppServiceSchema>;
