"use client";

import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Box, Button, Center, Field, Flex, Grid, Heading, HStack, Input, Spinner, Stack, Text } from "@chakra-ui/react";
import { LuArrowLeft, LuArrowRight, LuInfo, LuLink2, LuLock, LuPencil, LuPlus, LuSave, LuX } from "react-icons/lu";
import { getAppService, updateAppServiceAppearance } from "@/apis";
import { AppTile } from "@/components/AppTile";
import { ColorSwatchPicker } from "@/components/ColorSwatchPicker";
import { IconPicker } from "@/components/IconPicker";
import { toaster } from "@/components/ui/toaster";
import { APP_COLORS } from "@/consts";
import { AURORA_CTA_STYLE } from "@/consts/styles";
import { useUser } from "@/hooks/useUser";
import { usePermissions } from "@/hooks/usePermissions";
import { AppShell } from "@/layouts/AppShell";
import type { AppService } from "@/types";

const GHOST_BUTTON_PROPS = {
	variant: "outline",
	borderRadius: "glassSm",
	borderColor: "border.strong",
	bg: "bg.glass",
	color: "fg",
	fontWeight: "semibold",
	css: { backdropFilter: "blur(12px)", WebkitBackdropFilter: "blur(12px)" },
	_hover: { bg: "bg.glassHi", borderColor: "rgba(255,255,255,0.28)" },
} as const;

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

// Mirrors RegisterAppServiceDialog INPUT_PROPS so the redirect URL field is
// visually identical to the create-time field.
const INPUT_PROPS = {
	h: "11",
	borderRadius: "glassSm",
	bg: "bg.glass",
	borderColor: "border.strong",
	fontSize: "sm",
	color: "fg",
	_placeholder: { color: "fg.muted" },
	_hover: { borderColor: "rgba(255,255,255,0.28)" },
	_focus: { borderColor: "aurora.violet", boxShadow: "focusRing", outline: "none", bg: "rgba(255,255,255,0.08)" },
} as const;

const PV_CAP_PROPS = {
	fontSize: "11px",
	fontWeight: "semibold",
	letterSpacing: "0.12em",
	textTransform: "uppercase",
	color: "fg.muted",
} as const;

/** Max ADDITIONAL redirect URLs (the allowlist) — mirrors models.MaxAdditionalRedirectURLs. */
const MAX_ADDITIONAL_REDIRECT_URLS = 3;

/** Compare two string lists for order-sensitive equality (drives the dirty check). */
const sameUrlList = (a: string[], b: string[]) => a.length === b.length && a.every((value, index) => value === b[index]);

export const EditAppService = () => {
	const { id = "" } = useParams();
	const navigate = useNavigate();
	const { user: currentUser } = useUser();
	const { can } = usePermissions();
	const canUpdate = can("app_service:update");

	const [app, setApp] = useState<AppService | null>(null);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);

	// saved (server) values vs. local edits
	const [savedName, setSavedName] = useState("");
	const [savedRedirect, setSavedRedirect] = useState("");
	const [savedRedirectUrls, setSavedRedirectUrls] = useState<string[]>([]);
	const [savedIcon, setSavedIcon] = useState("");
	const [savedColor, setSavedColor] = useState("");
	const [name, setName] = useState("");
	const [redirect, setRedirect] = useState("");
	const [redirectUrls, setRedirectUrls] = useState<string[]>([]);
	const [icon, setIcon] = useState("");
	const [color, setColor] = useState("");

	useEffect(() => {
		let active = true;
		setLoading(true);
		getAppService(id)
			.then((data) => {
				if (!active) return;
				// isme is the read-only platform app — block direct URL access
				// (the backend rejects the PATCH too, this just avoids a dead form).
				if (data.app_code === "isme") {
					toaster.create({ title: "The isme platform app is read-only", type: "error", meta: { closable: true } });
					navigate("/app-services", { replace: true });
					return;
				}
				setApp(data);
				const loadedUrls = data.redirect_urls ?? [];
				setSavedName(data.app_name);
				setSavedRedirect(data.redirect_url ?? "");
				setSavedRedirectUrls(loadedUrls);
				setSavedIcon(data.icon);
				setSavedColor(data.color);
				setName(data.app_name);
				setRedirect(data.redirect_url ?? "");
				setRedirectUrls(loadedUrls);
				setIcon(data.icon);
				setColor(data.color);
			})
			.catch(() => {
				if (active) toaster.create({ title: "Failed to load app service", type: "error", meta: { closable: true } });
			})
			.finally(() => {
				if (active) setLoading(false);
			});
		return () => {
			active = false;
		};
	}, [id, navigate]);

	// trimmed, blank-row-dropped view of the allowlist used for both the dirty
	// check and submit (so a trailing empty row doesn't falsely mark the form dirty).
	const cleanedRedirectUrls = useMemo(() => redirectUrls.map((url) => url.trim()).filter((url) => url !== ""), [redirectUrls]);
	const urlsDirty = !sameUrlList(cleanedRedirectUrls, savedRedirectUrls);
	const dirty = name !== savedName || redirect !== savedRedirect || urlsDirty || icon !== savedIcon || color !== savedColor;
	const colorHex = useMemo(() => (color && color in APP_COLORS ? APP_COLORS[color as keyof typeof APP_COLORS].hex : ""), [color]);

	const addRedirectUrl = () => {
		if (redirectUrls.length >= MAX_ADDITIONAL_REDIRECT_URLS) return;
		setRedirectUrls((prev) => [...prev, ""]);
	};

	const updateRedirectUrl = (index: number, value: string) => {
		setRedirectUrls((prev) => prev.map((url, i) => (i === index ? value : url)));
	};

	const removeRedirectUrl = (index: number) => {
		setRedirectUrls((prev) => prev.filter((_, i) => i !== index));
	};

	const discard = () => {
		setName(savedName);
		setRedirect(savedRedirect);
		setRedirectUrls(savedRedirectUrls);
		setIcon(savedIcon);
		setColor(savedColor);
	};

	const handleSave = async () => {
		if (!dirty || !app) return;
		setSaving(true);
		try {
			await updateAppServiceAppearance(app.id, {
				app_name: name !== savedName ? name : undefined,
				redirect_url: redirect !== savedRedirect ? redirect : undefined,
				redirect_urls: urlsDirty ? cleanedRedirectUrls : undefined,
				icon: icon !== savedIcon ? icon : undefined,
				color: color !== savedColor ? color : undefined,
			});
			toaster.create({ title: "Appearance saved", type: "success", meta: { closable: true } });
			navigate("/app-services");
		} catch (error: unknown) {
			const err = error as { response?: { data?: { message?: string } } };
			toaster.create({ title: err?.response?.data?.message || "Failed to save appearance", type: "error", meta: { closable: true } });
		} finally {
			setSaving(false);
		}
	};

	const userMeta = { name: currentUser?.name || "User", email: currentUser?.email || "" };
	const displayName = name || "—";

	return (
		<AppShell active="appServices" user={userMeta}>
			{/* breadcrumb */}
			<HStack gap="9px" fontSize="13px" color="fg.muted">
				<Text as="button" color="fg.muted" _hover={{ color: "fg" }} onClick={() => navigate("/app-services")}>
					App Services
				</Text>
				<Text opacity="0.5">/</Text>
				<Text color="fg.subtle">{app ? app.app_name : "…"}</Text>
				<Text opacity="0.5">/</Text>
				<Text color="fg.subtle">Edit settings</Text>
			</HStack>

			{/* page head */}
			<Flex align="flex-end" justify="space-between" gap="16px" flexWrap="wrap">
				<Box>
					<Heading as="h1" fontSize="32px" fontWeight="bold" letterSpacing="-0.025em" lineHeight="1.1">
						Edit App Service
					</Heading>
					<Text mt="6px" color="fg.muted" fontSize="14px">
						Update the display name, SSO redirect URL, and appearance (icon &amp; color) for{" "}
						<Text as="code" color="fg.subtle" fontFamily="inherit">
							{app?.app_code || "…"}
						</Text>
						.
					</Text>
				</Box>
				<Button h="11" px="4.5" fontSize="sm" {...GHOST_BUTTON_PROPS} onClick={() => navigate("/app-services")}>
					<LuArrowLeft size={16} /> Back
				</Button>
			</Flex>

			{loading ? (
				<Center py="20">
					<Spinner color="aurora.violet" />
				</Center>
			) : !app ? (
				<Center py="20">
					<Text color="fg.muted">App service not found.</Text>
				</Center>
			) : (
				<Grid gap="20px" alignItems="start" gridTemplateColumns={{ base: "1fr", lg: "1fr 340px" }}>
					{/* ===== LEFT: edit form ===== */}
					<Box {...PANEL_PROPS}>
						{/* section head */}
						<HStack gap="12px" px="20px" py="16px" borderBottomWidth="1px" borderColor="border">
							<Center
								w="36px"
								h="36px"
								borderRadius="11px"
								flex="none"
								color="aurora.cyan"
								borderWidth="1px"
								borderColor="border.strong"
								css={{ background: "linear-gradient(135deg, rgba(99,102,241,0.25), rgba(236,72,153,0.20))" }}
							>
								<LuPencil size={16} />
							</Center>
							<Box lineHeight="1.25">
								<Text fontSize="15px" fontWeight="semibold">
									App settings &amp; appearance
								</Text>
								<Text fontSize="12px" color="fg.muted">
									app_id ·{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										{app.id}
									</Text>
								</Text>
							</Box>
						</HStack>

						<Stack gap="20px" p="20px">
							{/* app_code — LOCKED identity key */}
							<Box>
								<Text {...FIELD_LABEL_PROPS}>
									App code{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										· identity key, immutable after register
									</Text>
								</Text>
								<HStack
									gap="10px"
									h="44px"
									px="14px"
									borderRadius="glassSm"
									borderWidth="1px"
									borderStyle="dashed"
									borderColor="border.strong"
									css={{ background: "rgba(7,7,26,0.40)" }}
								>
									<Input
										value={app.app_code}
										readOnly
										tabIndex={-1}
										flex="1"
										h="full"
										border="0"
										bg="transparent"
										px="0"
										fontSize="14px"
										color="fg.subtle"
										cursor="not-allowed"
										css={{ fontVariantNumeric: "tabular-nums", outline: "none", boxShadow: "none" }}
									/>
									<Box color="fg.muted" flex="none">
										<LuLock size={14} />
									</Box>
								</HStack>
								<Text mt="8px" fontSize="12px" color="fg.muted">
									The app code is part of the SSO contract (used in{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										resource_access
									</Text>{" "}
									token keys) — it can't be changed.
								</Text>
							</Box>

							{/* app_name — editable */}
							<Box>
								<Text {...FIELD_LABEL_PROPS}>Display name</Text>
								<HStack
									gap="10px"
									h="44px"
									px="14px"
									borderRadius="glassSm"
									borderWidth="1px"
									borderColor="border.strong"
									bg="bg.glass"
									_focusWithin={{ borderColor: "aurora.violet", boxShadow: "focusRing", bg: "rgba(255,255,255,0.08)" }}
								>
									<Input
										value={name}
										onChange={(event) => setName(event.target.value)}
										flex="1"
										h="full"
										border="0"
										bg="transparent"
										px="0"
										fontSize="14px"
										color="fg"
										css={{ outline: "none", boxShadow: "none" }}
									/>
								</HStack>
							</Box>

							{/* redirect_url — FUNCTIONAL/SECURITY field (SSO contract) */}
							<Field.Root>
								<Field.Label {...FIELD_LABEL_PROPS} mb="0">
									Redirect URL
									<HStack
										as="span"
										display="inline-flex"
										gap="5px"
										ml="8px"
										px="8px"
										py="1px"
										borderRadius="full"
										fontSize="11px"
										fontWeight="semibold"
										letterSpacing="0.02em"
										color="aurora.amber"
										bg="rgba(245,158,11,0.12)"
										borderWidth="1px"
										borderColor="rgba(245,158,11,0.32)"
										verticalAlign="middle"
									>
										<LuLock size={11} /> SSO contract
									</HStack>
								</Field.Label>
								<Box w="full" position="relative" mt="8px" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
									<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
										<LuLink2 size={16} />
									</Box>
									<Input
										{...INPUT_PROPS}
										pl="10"
										type="url"
										inputMode="url"
										placeholder="https://app.example.com/auth/callback"
										value={redirect}
										onChange={(event) => setRedirect(event.target.value)}
									/>
								</Box>
								<Field.HelperText mt="8px" fontSize="12px" color="fg.muted">
									OAuth callback the app redirects to after SSO login. Must be an{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										https://
									</Text>{" "}
									URL — a mismatch breaks the login round-trip. Leave empty to clear.
								</Field.HelperText>
							</Field.Root>

							{/* ADDITIONAL REDIRECT URLs — OAuth-style allowlist of EXTRA permitted callbacks.
							    The primary redirect_url above stays the default; these are extra URLs the
							    SSO client may also redirect to. Capped at 3; the Add button disables at the cap. */}
							<Box>
								<Text {...FIELD_LABEL_PROPS}>
									Additional redirect URLs{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										· allowlist, max {MAX_ADDITIONAL_REDIRECT_URLS}
									</Text>
								</Text>
								{redirectUrls.length === 0 ? (
									<Text fontSize="12px" color="fg.muted" fontStyle="italic">
										No additional URLs — the primary redirect URL is the only allowed callback.
									</Text>
								) : (
									<Stack gap="10px">
										{redirectUrls.map((url, index) => (
											<HStack key={index} gap="8px" alignItems="center">
												<Box flex="1" minW="0" position="relative" css={{ "&:focus-within .field-icon": { color: "#22D3EE" } }}>
													<Box className="field-icon" position="absolute" left="3.5" top="3.5" color="fg.muted" pointerEvents="none" zIndex="1">
														<LuLink2 size={16} />
													</Box>
													<Input
														{...INPUT_PROPS}
														pl="10"
														type="url"
														inputMode="url"
														placeholder="https://app.example.com/auth/callback/alt"
														value={url}
														onChange={(event) => updateRedirectUrl(index, event.target.value)}
													/>
												</Box>
												<Button
													variant="ghost"
													size="xs"
													p="1.5"
													minW="auto"
													h="11"
													borderRadius="9px"
													color="fg.muted"
													_hover={{ bg: "rgba(236,72,153,0.14)", color: "aurora.magenta" }}
													aria-label="Remove URL"
													onClick={() => removeRedirectUrl(index)}
												>
													<LuX size={16} />
												</Button>
											</HStack>
										))}
									</Stack>
								)}
								<Button
									variant="outline"
									mt="10px"
									h="40px"
									px="14px"
									borderRadius="glassSm"
									borderColor="border.strong"
									borderStyle="dashed"
									bg="transparent"
									fontSize="13px"
									fontWeight="medium"
									color="fg.muted"
									_hover={{ color: "fg", borderColor: "aurora.violet", bg: "rgba(139,92,246,0.10)" }}
									disabled={redirectUrls.length >= MAX_ADDITIONAL_REDIRECT_URLS}
									onClick={addRedirectUrl}
								>
									<LuPlus size={15} /> Add redirect URL
								</Button>
								<Text mt="8px" fontSize="12px" color="fg.muted">
									Extra callback URLs this app may redirect to after login. Each must be a valid URL.{" "}
									<Text as="b" color="fg.subtle" fontWeight="semibold">
										Max {MAX_ADDITIONAL_REDIRECT_URLS}.
									</Text>
								</Text>
							</Box>

							{/* ICON PICKER — 20-icon grid */}
							<Box>
								<Text {...FIELD_LABEL_PROPS}>
									Icon{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										· pick from the shared icon set
									</Text>
								</Text>
								<IconPicker value={icon} onChange={setIcon} ariaLabel="App icon" />
								<Text mt="8px" fontSize="12px" color="fg.muted">
									Stored as a string key (e.g.{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										{icon || "neutral"}
									</Text>
									). Same 20-icon allowlist used by the permission catalog.
								</Text>
							</Box>

							{/* COLOR PICKER — fixed aurora palette */}
							<Box>
								<Text {...FIELD_LABEL_PROPS}>
									Color{" "}
									<Text as="span" color="fg.muted" fontWeight="normal" fontSize="12px">
										· aurora palette
									</Text>
								</Text>
								<ColorSwatchPicker value={color} onChange={setColor} ariaLabel="App color" />
								<Text mt="10px" fontSize="12px" color="fg.muted">
									Stored as palette key{" "}
									<Text as="b" color="fg.subtle" fontWeight="semibold">
										{color || "neutral"}
									</Text>
									{colorHex && (
										<>
											{" "}
											· <Text as="span" css={{ fontVariantNumeric: "tabular-nums" }}>{colorHex}</Text>
										</>
									)}
								</Text>
							</Box>

							{/* note */}
							<HStack
								gap="9px"
								px="14px"
								py="11px"
								borderRadius="glassSm"
								borderWidth="1px"
								borderColor="border"
								bg="bg.glass"
								fontSize="12px"
								color="fg.muted"
								alignItems="flex-start"
							>
								<Box color="aurora.cyan" mt="1px" flex="none">
									<LuInfo size={15} />
								</Box>
								<Box>
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										redirect_url
									</Text>{" "}
									is part of the SSO contract (functional);{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										icon
									</Text>{" "}
									(string key) and{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										color
									</Text>{" "}
									(palette key) are appearance. All are saved on{" "}
									<Text as="code" color="fg.subtle" fontFamily="inherit">
										app_service
									</Text>{" "}
									in one update — appearance drives every rendered tile/chip, redirect URL drives the login round-trip.
								</Box>
							</HStack>
						</Stack>

						{/* footer actions */}
						<HStack gap="10px" px="20px" py="16px" borderTopWidth="1px" borderColor="border">
							{dirty && (
								<HStack gap="8px" fontSize="13px" color="aurora.violet" fontWeight="semibold">
									<Box w="7px" h="7px" borderRadius="full" bg="aurora.violet" css={{ boxShadow: "0 0 10px #8B5CF6" }} />
									Unsaved changes
								</HStack>
							)}
							<Box flex="1" />
							<Button h="11" px="4.5" fontSize="sm" {...GHOST_BUTTON_PROPS} disabled={!dirty} onClick={discard}>
								Cancel
							</Button>
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
								disabled={!dirty || !canUpdate}
								loading={saving}
								onClick={handleSave}
							>
								<LuSave size={15} /> Save changes
							</Button>
						</HStack>
					</Box>

					{/* ===== RIGHT: live preview rail ===== */}
					<Box {...PANEL_PROPS} position="sticky" top="84px">
						<Text px="18px" py="14px" borderBottomWidth="1px" borderColor="border" {...PV_CAP_PROPS} letterSpacing="0.14em">
							Live preview
						</Text>
						<Stack gap="22px" p="20px">
							{/* detail hero (64px) */}
							<Stack gap="10px">
								<Text {...PV_CAP_PROPS}>Detail header</Text>
								<HStack gap="14px">
									<AppTile iconKey={icon} colorKey={color} size="detail" fallbackSeed={app.id} appCode={app.app_code} />
									<Box lineHeight="1.3" minW="0">
										<Text fontSize="16px" fontWeight="bold" letterSpacing="-0.01em" truncate>
											{displayName}
										</Text>
										<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
											{app.app_code}
										</Text>
										{/* redirect URL as functional text — empty state called out in amber */}
										<HStack gap="7px" mt="4px" fontSize="11px" minW="0">
											<Box color="aurora.amber" flex="none">
												<LuLink2 size={12} />
											</Box>
											{redirect.trim() ? (
												<Text color="fg.subtle" truncate maxW="220px" css={{ fontVariantNumeric: "tabular-nums" }}>
													{redirect.trim()}
												</Text>
											) : (
												<Text color="aurora.amber" fontStyle="italic" truncate>
													No redirect URL set
												</Text>
											)}
										</HStack>
									</Box>
								</HStack>
							</Stack>

							{/* list row (34px) */}
							<Stack gap="10px">
								<Text {...PV_CAP_PROPS}>List row</Text>
								<HStack gap="12px" px="12px" py="10px" borderRadius="12px" borderWidth="1px" borderColor="border" bg="rgba(255,255,255,0.03)">
									<AppTile iconKey={icon} colorKey={color} size="list" fallbackSeed={app.id} appCode={app.app_code} />
									<Box minW="0" lineHeight="1.25">
										<Text fontSize="14px" fontWeight="semibold" truncate>
											{displayName}
										</Text>
										<Text fontSize="12px" color="fg.muted" css={{ fontVariantNumeric: "tabular-nums" }}>
											{app.app_code}
										</Text>
									</Box>
								</HStack>
							</Stack>

							{/* role chip (22px) */}
							<Stack gap="10px">
								<Text {...PV_CAP_PROPS}>Role chip (AppRoleChip)</Text>
								<HStack
									as="span"
									display="inline-flex"
									gap="8px"
									pl="5px"
									pr="12px"
									py="5px"
									borderRadius="full"
									bg="bg.glass"
									borderWidth="1px"
									borderColor="border.strong"
									fontSize="13px"
									fontWeight="medium"
									color="fg.subtle"
									w="fit-content"
								>
									<AppTile iconKey={icon} colorKey={color} size="chip" fallbackSeed={app.id} appCode={app.app_code} />
									<Text as="span" truncate maxW="200px">
										{displayName}
									</Text>
								</HStack>
							</Stack>

							{/* before → after */}
							<Stack gap="10px">
								<Text {...PV_CAP_PROPS}>Before → after</Text>
								<HStack gap="14px" align="center">
									<Stack gap="7px" align="center">
										{/* BEFORE: today's fallback (hashed gradient / APP_META) — no stored color */}
										<AppTile size="list" fallbackSeed={app.id} appCode={app.app_code} />
										<Text {...PV_CAP_PROPS}>before</Text>
									</Stack>
									<Box color="fg.muted">
										<LuArrowRight size={18} />
									</Box>
									<Stack gap="7px" align="center">
										<AppTile iconKey={icon} colorKey={color} size="list" fallbackSeed={app.id} appCode={app.app_code} />
										<Text {...PV_CAP_PROPS}>stored</Text>
									</Stack>
								</HStack>
							</Stack>
						</Stack>
					</Box>
				</Grid>
			)}
		</AppShell>
	);
};
