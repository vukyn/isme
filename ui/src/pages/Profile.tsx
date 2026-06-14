"use client";

import { useMemo, useRef, useState } from "react";
import { Box, Center, HStack, Input, Spinner, Stack, Text } from "@chakra-ui/react";
import {
	LuCalendar,
	LuImage,
	LuInfo,
	LuKeyRound,
	LuLink,
	LuLock,
	LuSave,
	LuTriangleAlert,
	LuUpload,
	LuUser,
	LuX,
} from "react-icons/lu";
import { changePassword, updateProfile, uploadMedia } from "@/apis";
import { Button } from "@/components/ui/button";
import { ImageSourceToggle, type ImageSource } from "@/components/ui/image-source-toggle";
import { PasswordField } from "@/components/ui/password-field";
import { PasswordStrength } from "@/components/ui/password-strength";
import { toaster } from "@/components/ui/toaster";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { AppShell } from "@/layouts/AppShell";
import { useUser } from "@/hooks/useUser";
import { avatarGradient, avatarInitials } from "@/utils";

// === shared aurora chrome (mirrors the Settings cards / aurora-profile mock) ===

const PANEL_PROPS = {
	bg: "bg.glass",
	borderWidth: "1px",
	borderColor: "border",
	borderRadius: "20px",
	overflow: "hidden",
	css: { backdropFilter: "blur(20px) saturate(1.15)", WebkitBackdropFilter: "blur(20px) saturate(1.15)" },
} as const;

const FIELD_LABEL_PROPS = {
	display: "block",
	fontSize: "13px",
	fontWeight: "medium",
	color: "fg.subtle",
	mb: "8px",
} as const;

const INPUT_WRAP_PROPS = {
	gap: "10px",
	h: "44px",
	px: "14px",
	borderRadius: "glassSm",
	borderWidth: "1px",
	borderColor: "border.strong",
	bg: "bg.glass",
	css: { backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" },
	_focusWithin: { borderColor: "aurora.violet", boxShadow: "focusRing", background: "rgba(255,255,255,0.08)" },
} as const;

const INPUT_INNER_PROPS = {
	flex: "1",
	h: "full",
	border: "0",
	bg: "transparent",
	px: "0",
	fontSize: "14px",
	color: "fg",
	_focusVisible: { boxShadow: "none", outline: "none" },
} as const;

const MAX_AVATAR_BYTES = 2 * 1024 * 1024;
const ALLOWED_AVATAR_TYPES = ["image/png", "image/jpeg", "image/webp"];

// Reusable card header (the mock's .sect-head).
const SectionHead = ({ icon, title, subtitle }: { icon: React.ReactNode; title: string; subtitle: string }) => (
	<HStack gap="13px" px="20px" py="18px" borderBottomWidth="1px" borderColor="border">
		<Center
			w="40px"
			h="40px"
			borderRadius="12px"
			flex="none"
			color="aurora.cyan"
			borderWidth="1px"
			borderColor="border.strong"
			css={{
				background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))",
				boxShadow: "0 0 18px rgba(99,102,241,0.18)",
			}}
		>
			{icon}
		</Center>
		<Box lineHeight="1.3" minW="0">
			<Text fontSize="16px" fontWeight="semibold" letterSpacing="-0.01em">
				{title}
			</Text>
			<Text fontSize="12px" color="fg.muted">
				{subtitle}
			</Text>
		</Box>
	</HStack>
);

// Footer action bar (the mock's .form-foot).
const FormFooter = ({ children }: { children: React.ReactNode }) => (
	<HStack gap="10px" px="20px" py="16px" borderTopWidth="1px" borderColor="border">
		<Box flex="1" />
		{children}
	</HStack>
);

const PrimaryButton = (props: React.ComponentProps<typeof Button>) => (
	<Button
		h="11"
		px="4.5"
		borderRadius="glassSm"
		fontSize="sm"
		fontWeight="semibold"
		color="white"
		css={AURORA_CTA_STYLE}
		boxShadow="ctaGlow"
		_hover={{ boxShadow: "ctaGlowHi", backgroundPosition: "100% 100%" }}
		{...props}
	/>
);

const GhostButton = (props: React.ComponentProps<typeof Button>) => (
	<Button
		h="11"
		px="4.5"
		fontSize="sm"
		variant="outline"
		borderRadius="glassSm"
		borderColor="border.strong"
		bg="bg.glass"
		color="fg"
		fontWeight="semibold"
		css={{ backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" }}
		_hover={{ bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" }}
		{...props}
	/>
);

// Avatar circle — image cover when a URL is set, else the gradient + initials
// fallback (matches the mock's .avatar-xl / .avatar-preview).
const AvatarCircle = ({
	id,
	name,
	url,
	size,
	fontSize,
}: {
	id: string;
	name: string;
	url: string;
	size: string;
	fontSize: string;
}) => (
	<Center
		w={size}
		h={size}
		flex="none"
		borderRadius="full"
		color="white"
		fontWeight="bold"
		fontSize={fontSize}
		letterSpacing="-0.01em"
		borderWidth="2px"
		borderColor="rgba(255,255,255,0.14)"
		boxShadow="0 0 28px rgba(139,92,246,0.45), 0 8px 24px rgba(0,0,0,0.45)"
		css={{
			background: url ? `center / cover no-repeat url(${JSON.stringify(url)})` : avatarGradient(id),
		}}
		aria-label={`${name} avatar`}
	>
		{url ? "" : avatarInitials(name)}
	</Center>
);

const formatBytes = (bytes: number): string => {
	if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	return `${Math.max(1, Math.round(bytes / 1024))} KB`;
};

const memberSince = (createdAt: string): string => {
	if (!createdAt) return "";
	const date = new Date(createdAt);
	if (Number.isNaN(date.getTime())) return "";
	return date.toLocaleDateString(undefined, { month: "short", year: "numeric" });
};

export const Profile = () => {
	const { user, loading, refetch } = useUser();

	if (loading || !user) {
		return (
			<AppShell active="overview" user={{ name: user?.name || "User", email: user?.email || "" }}>
				<Center py="20">
					<Spinner size="xl" color="accent" />
				</Center>
			</AppShell>
		);
	}

	return (
		<AppShell active="overview" user={{ name: user.name, email: user.email }}>
			<ProfileContent
				id={user.id}
				name={user.name}
				email={user.email}
				avatarURL={user.avatar_url}
				createdAt={user.created_at}
				onSaved={refetch}
			/>
		</AppShell>
	);
};

interface ProfileContentProps {
	id: string;
	name: string;
	email: string;
	avatarURL: string;
	createdAt: string;
	onSaved: () => Promise<unknown>;
}

const ProfileContent = ({ id, name, email, avatarURL, createdAt, onSaved }: ProfileContentProps) => {
	// === Avatar card state ===
	const fileInputRef = useRef<HTMLInputElement>(null);
	const [mode, setMode] = useState<ImageSource>("file");
	const [selectedFile, setSelectedFile] = useState<File | null>(null);
	const [filePreview, setFilePreview] = useState<string>("");
	const [fileError, setFileError] = useState<string>("");
	const [linkURL, setLinkURL] = useState("");
	const [savingAvatar, setSavingAvatar] = useState(false);

	const linkValid = /^https?:\/\/.+/.test(linkURL.trim());
	const previewURL = mode === "file" ? filePreview : linkValid ? linkURL.trim() : "";
	const avatarDirty = mode === "file" ? !!selectedFile : linkValid;

	const pickFile = (file: File | null) => {
		if (!file) return;
		if (!ALLOWED_AVATAR_TYPES.includes(file.type)) {
			setFileError("That file must be a PNG, JPEG, or WebP image.");
			return;
		}
		if (file.size > MAX_AVATAR_BYTES) {
			setFileError("That image is over 2MB. Please choose a smaller file.");
			return;
		}
		setFileError("");
		setSelectedFile(file);
		setFilePreview(URL.createObjectURL(file));
	};

	const clearFile = () => {
		setSelectedFile(null);
		setFilePreview("");
		setFileError("");
		if (fileInputRef.current) fileInputRef.current.value = "";
	};

	const cancelAvatar = () => {
		clearFile();
		setLinkURL("");
		setMode("file");
	};

	const handleSaveAvatar = async () => {
		setSavingAvatar(true);
		try {
			let nextURL = "";
			if (mode === "file" && selectedFile) {
				const result = await uploadMedia(selectedFile);
				nextURL = result.url;
			} else if (mode === "link" && linkValid) {
				nextURL = linkURL.trim();
			}
			await updateProfile({ name, avatar_url: nextURL });
			await onSaved();
			toaster.create({ title: "Profile photo updated", type: "success", meta: { closable: true } });
			cancelAvatar();
		} catch (err) {
			const message =
				(err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
				"Failed to update photo";
			toaster.create({ title: message, type: "error", meta: { closable: true } });
		} finally {
			setSavingAvatar(false);
		}
	};

	// === Basic info card state ===
	const [displayName, setDisplayName] = useState(name);
	const [savingName, setSavingName] = useState(false);
	const nameDirty = displayName.trim() !== name && displayName.trim().length > 0;

	const handleSaveName = async () => {
		setSavingName(true);
		try {
			await updateProfile({ name: displayName.trim(), avatar_url: avatarURL });
			await onSaved();
			toaster.create({ title: "Display name updated", type: "success", meta: { closable: true } });
		} catch (err) {
			const message =
				(err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
				"Failed to save changes";
			toaster.create({ title: message, type: "error", meta: { closable: true } });
		} finally {
			setSavingName(false);
		}
	};

	// === Change password card state ===
	const [currentPw, setCurrentPw] = useState("");
	const [newPw, setNewPw] = useState("");
	const [confirmPw, setConfirmPw] = useState("");
	const [savingPw, setSavingPw] = useState(false);

	const pwMismatch = confirmPw.length > 0 && confirmPw !== newPw;
	const pwValid = currentPw.length > 0 && newPw.length >= 8 && newPw === confirmPw;

	const handleChangePassword = async () => {
		setSavingPw(true);
		try {
			await changePassword({ old_password: currentPw, new_password: newPw });
			toaster.create({ title: "Password updated", type: "success", meta: { closable: true } });
			setCurrentPw("");
			setNewPw("");
			setConfirmPw("");
		} catch (err) {
			const message =
				(err as { response?: { data?: { message?: string } } })?.response?.data?.message ||
				"Failed to update password";
			toaster.create({ title: message, type: "error", meta: { closable: true } });
		} finally {
			setSavingPw(false);
		}
	};

	const since = useMemo(() => memberSince(createdAt), [createdAt]);

	return (
		<Stack gap="5" maxW="880px" w="full" mx="auto">
			<Box>
				<Text as="h1" fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1" color="fg">
					Profile
				</Text>
				<Text mt="6px" color="fg.muted" fontSize="14px">
					Manage your photo, display name, and password.
				</Text>
			</Box>

			{/* ===== CARD 1 · Profile header (read-only identity) ===== */}
			<Box {...PANEL_PROPS}>
				<HStack gap="22px" p="24px" flexWrap="wrap">
					<AvatarCircle id={id} name={name} url={avatarURL} size="96px" fontSize="34px" />
					<Box minW="0">
						<Text fontSize="26px" fontWeight="bold" letterSpacing="-0.02em" lineHeight="1.15">
							{name}
						</Text>
						<Text mt="4px" color="fg.subtle" fontSize="14px">
							{email}
						</Text>
						{since && (
							<HStack mt="12px" gap="8px" fontSize="12.5px" color="fg.muted">
								<Box color="aurora.mint" flex="none">
									<LuCalendar size={14} />
								</Box>
								<Text>Member since {since}</Text>
							</HStack>
						)}
					</Box>
				</HStack>
			</Box>

			{/* ===== CARD 2 · Avatar (update photo) ===== */}
			<Box {...PANEL_PROPS}>
				<SectionHead
					icon={<LuImage size={19} />}
					title="Profile photo"
					subtitle="Upload an image or paste a link. Used across all your apps."
				/>
				<Box p="20px">
					<HStack align="flex-start" gap="24px" flexWrap="wrap">
						<AvatarCircle id={id} name={name} url={previewURL || avatarURL} size="80px" fontSize="28px" />

						<Stack flex="1" minW="260px" gap="16px">
							<ImageSourceToggle value={mode} onChange={setMode} />

							{/* Mode A · Upload file */}
							{mode === "file" && (
								<Box>
									<input
										ref={fileInputRef}
										type="file"
										accept="image/png,image/jpeg,image/webp"
										style={{ display: "none" }}
										onChange={(event) => pickFile(event.target.files?.[0] ?? null)}
									/>
									{!selectedFile ? (
										<Stack
											gap="12px"
											justifyItems="center"
											textAlign="center"
											p="22px"
											borderWidth="1.5px"
											borderStyle="dashed"
											borderColor="border.strong"
											borderRadius="glassSm"
											cursor="pointer"
											css={{ background: "rgba(7,7,26,0.30)" }}
											_hover={{ borderColor: "aurora.violet", background: "rgba(139,92,246,0.06)" }}
											onClick={() => fileInputRef.current?.click()}
										>
											<Center
												w="44px"
												h="44px"
												mx="auto"
												borderRadius="12px"
												color="aurora.cyan"
												borderWidth="1px"
												borderColor="border.strong"
												css={{ background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(34,211,238,0.18))" }}
											>
												<LuUpload size={20} />
											</Center>
											<Text fontSize="13px" color="fg.subtle">
												<Text as="b" color="fg">
													Choose image
												</Text>{" "}
												or click to browse
											</Text>
											<Text fontSize="12px" color="fg.muted">
												PNG, JPEG or WebP · max 2MB
											</Text>
										</Stack>
									) : (
										<HStack
											gap="12px"
											p="10px 12px"
											borderWidth="1px"
											borderColor="border"
											borderRadius="glassSm"
											bg="bg.glass"
										>
											<Box
												w="38px"
												h="38px"
												flex="none"
												borderRadius="9px"
												borderWidth="1px"
												borderColor="border.strong"
												css={{ background: `center / cover no-repeat url(${JSON.stringify(filePreview)})` }}
											/>
											<Box flex="1" minW="0" lineHeight="1.3">
												<Text fontSize="13px" fontWeight="semibold" color="fg" truncate>
													{selectedFile.name}
												</Text>
												<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
													{formatBytes(selectedFile.size)}
												</Text>
											</Box>
											<Center
												as="button"
												w="30px"
												h="30px"
												flex="none"
												borderRadius="8px"
												color="fg.muted"
												cursor="pointer"
												aria-label="Remove file"
												_hover={{ color: "aurora.magenta", bg: "rgba(236,72,153,0.12)" }}
												onClick={clearFile}
											>
												<LuX size={16} />
											</Center>
										</HStack>
									)}
									{fileError && (
										<HStack gap="8px" mt="8px" fontSize="12.5px" color="#FCA5A5">
											<Box flex="none">
												<LuTriangleAlert size={15} />
											</Box>
											<Text>{fileError}</Text>
										</HStack>
									)}
								</Box>
							)}

							{/* Mode B · Paste link */}
							{mode === "link" && (
								<Box>
									<Text {...FIELD_LABEL_PROPS}>Image URL</Text>
									<HStack {...INPUT_WRAP_PROPS}>
										<Box color="fg.muted" flex="none">
											<LuLink size={16} />
										</Box>
										<Input
											type="url"
											placeholder="https://…/photo.jpg"
											value={linkURL}
											spellCheck={false}
											onChange={(event) => setLinkURL(event.target.value)}
											aria-label="Image URL"
											{...INPUT_INNER_PROPS}
										/>
									</HStack>
									<Text mt="12px" fontSize="12.5px" color="fg.muted">
										{linkValid ? "Looks good — this image will be used." : "Preview updates as you type a valid image URL."}
									</Text>
								</Box>
							)}

							<HStack
								align="flex-start"
								gap="9px"
								px="14px"
								py="11px"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border"
								bg="bg.glass"
								fontSize="12px"
								color="fg.muted"
							>
								<Box color="aurora.cyan" flex="none" mt="1px">
									<LuInfo size={15} />
								</Box>
								<Text>
									Uploaded photos are stored on{" "}
									<Text as="b" color="fg.subtle" fontWeight="semibold">
										medioa
									</Text>
									; we keep the returned link on your profile.
								</Text>
							</HStack>
						</Stack>
					</HStack>
				</Box>
				<FormFooter>
					<GhostButton disabled={!avatarDirty || savingAvatar} onClick={cancelAvatar}>
						Cancel
					</GhostButton>
					<PrimaryButton disabled={!avatarDirty} loading={savingAvatar} onClick={handleSaveAvatar}>
						<LuSave size={15} /> Save photo
					</PrimaryButton>
				</FormFooter>
			</Box>

			{/* ===== CARD 3 · Basic info (display name + read-only email) ===== */}
			<Box {...PANEL_PROPS}>
				<SectionHead
					icon={<LuUser size={19} />}
					title="Basic info"
					subtitle="Your display name across the platform."
				/>
				<Box p="20px">
					<Stack direction={{ base: "column", md: "row" }} gap="18px">
						<Box flex="1" minW="0">
							<Text {...FIELD_LABEL_PROPS}>Display name</Text>
							<HStack {...INPUT_WRAP_PROPS}>
								<Box color="fg.muted" flex="none">
									<LuUser size={16} />
								</Box>
								<Input
									type="text"
									value={displayName}
									onChange={(event) => setDisplayName(event.target.value)}
									aria-label="Display name"
									{...INPUT_INNER_PROPS}
								/>
							</HStack>
							<Text mt="8px" fontSize="12px" color="fg.muted">
								Shown on your avatar chip and in audit logs.
							</Text>
						</Box>

						<Box flex="1" minW="0">
							<Text {...FIELD_LABEL_PROPS}>
								Email{" "}
								<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
									· can't be changed
								</Text>
							</Text>
							<HStack
								{...INPUT_WRAP_PROPS}
								borderStyle="dashed"
								css={{ ...INPUT_WRAP_PROPS.css, background: "rgba(7,7,26,0.35)" }}
							>
								<Box color="fg.muted" flex="none">
									<LuLock size={16} />
								</Box>
								<Input
									type="email"
									value={email}
									disabled
									aria-label="Email (read-only)"
									{...INPUT_INNER_PROPS}
									color="fg.muted"
								/>
							</HStack>
							<Text mt="8px" fontSize="12px" color="fg.muted">
								Email is your sign-in identity and can't be changed here.
							</Text>
						</Box>
					</Stack>
				</Box>
				<FormFooter>
					<GhostButton disabled={!nameDirty || savingName} onClick={() => setDisplayName(name)}>
						Cancel
					</GhostButton>
					<PrimaryButton disabled={!nameDirty} loading={savingName} onClick={handleSaveName}>
						<LuSave size={15} /> Save changes
					</PrimaryButton>
				</FormFooter>
			</Box>

			{/* ===== CARD 4 · Change password ===== */}
			<Box {...PANEL_PROPS}>
				<SectionHead
					icon={<LuKeyRound size={19} />}
					title="Change password"
					subtitle="Use at least 8 characters. You'll stay signed in on this device."
				/>
				<Stack gap="22px" p="20px">
					<PasswordField
						label="Current password"
						value={currentPw}
						onChange={(event) => setCurrentPw(event.target.value)}
						autoComplete="current-password"
						name="current-password"
						placeholder="Enter current password"
					/>
					<Box>
						<PasswordField
							label="New password"
							value={newPw}
							onChange={(event) => setNewPw(event.target.value)}
							autoComplete="new-password"
							name="new-password"
							placeholder="Enter new password"
						/>
						{newPw.length > 0 && <PasswordStrength value={newPw} />}
					</Box>
					<Box>
						<PasswordField
							label="Confirm new password"
							value={confirmPw}
							onChange={(event) => setConfirmPw(event.target.value)}
							error={pwMismatch ? "Passwords don't match." : undefined}
							autoComplete="new-password"
							name="confirm-password"
							placeholder="Re-enter new password"
						/>
					</Box>
				</Stack>
				<FormFooter>
					<PrimaryButton disabled={!pwValid} loading={savingPw} onClick={handleChangePassword}>
						<LuKeyRound size={15} /> Update password
					</PrimaryButton>
				</FormFooter>
			</Box>
		</Stack>
	);
};
