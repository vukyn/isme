# Aurora UI Redesign — Design Spec

**Date**: 2026-04-29
**Status**: Approved (pending user review)
**Scope**: Frontend (`ui/`) only. No backend changes.
**Source mockup**: `design/auth-mockup.html`

## 1. Goal

Replace existing purple/light Chakra UI v3 frontend with Aurora UI: dark-only theme, animated mesh-gradient background, glass cards, iridescent text accents. Add a real welcome dashboard (mock data), forgot-password stub, and stubs for `/sessions`, `/team`, `/settings`. No backend or new APIs.

## 2. Non-Goals

- Light mode (dark-only).
- Real backend endpoints for sessions / activity / token rotations.
- Real forgot-password flow (email + token).
- SSOLogin layout redesign (token swap only).

## 3. Locked Decisions

| # | Question | Choice |
|---|----------|--------|
| Q1 | Welcome dashboard data | A — mock-only UI |
| Q2 | Aurora theme scope | A — global theme rewrite |
| Q3 | Forgot password | A — stub UI page |
| Q4 | SSOLogin redesign | C — token swap, keep single-card layout |
| Q5 | Topbar nav stubs | C — build `/sessions`, `/team`, `/settings` stub pages |
| Q6 | T&C checkbox on Signup | A — drop |
| Q7 | Eye toggle + strength meter | A — both |
| Q8 | Aurora background animation | A — global CSS, animated |
| Org | Code organization | Hybrid — shared layouts + reusable atoms |

## 4. Architecture

### 4.1 File tree

```
ui/src/
├── App.tsx                          [edit] mount AuroraBackground + new routes
├── index.css                        [edit] global aurora keyframes + reduced-motion guard
├── theme/
│   └── index.ts                     [rewrite] aurora tokens, dark only
├── layouts/                         [NEW]
│   ├── AuthLayout.tsx               split brand+form shell
│   └── AppShell.tsx                 topbar + main shell
├── components/ui/                   [NEW atoms]
│   ├── aurora-background.tsx
│   ├── brand-mark.tsx
│   ├── brand-panel.tsx
│   ├── glass-card.tsx
│   ├── feature-list.tsx
│   ├── password-field.tsx
│   ├── password-strength.tsx
│   ├── stat-card.tsx
│   ├── activity-row.tsx
│   ├── topbar.tsx
│   └── user-chip.tsx
├── pages/
│   ├── Login.tsx                    [rewrite]
│   ├── Signup.tsx                   [rewrite]
│   ├── Welcome.tsx                  [rewrite] → AppShell + dashboard
│   ├── ForgotPassword.tsx           [NEW]
│   ├── Sessions.tsx                 [NEW]
│   ├── Team.tsx                     [NEW]
│   ├── Settings.tsx                 [NEW]
│   ├── SSOLogin.tsx                 [edit] token swap only
│   ├── NotFound.tsx                 [edit minor]
│   └── Home.tsx                     [keep / minor]
├── consts/
│   └── mock.ts                      [NEW] mock dashboard data
└── validators/                      [edit] strip terms from signupSchema
```

### 4.2 Theme — `ui/src/theme/index.ts`

Full rewrite, no legacy `brand`/`darkGray`/`surface` tokens.

```ts
import { createSystem, defaultConfig, defineConfig } from "@chakra-ui/react";

const config = defineConfig({
  theme: {
    tokens: {
      colors: {
        canvas: {
          0: { value: "#07071A" },
          1: { value: "#0B0B23" },
          2: { value: "#12122E" },
          3: { value: "#04040E" },
        },
        aurora: {
          cyan:    { value: "#22D3EE" },
          blue:    { value: "#6366F1" },
          violet:  { value: "#8B5CF6" },
          magenta: { value: "#EC4899" },
          amber:   { value: "#F59E0B" },
          mint:    { value: "#34D399" },
        },
        ink: {
          DEFAULT: { value: "#F4F5FF" },
          soft:    { value: "#C7CAE8" },
          mute:    { value: "#8A8FB5" },
        },
        glass: {
          fill:    { value: "rgba(255,255,255,0.06)" },
          fillHi:  { value: "rgba(255,255,255,0.10)" },
          line:    { value: "rgba(255,255,255,0.10)" },
          lineHi:  { value: "rgba(255,255,255,0.18)" },
        },
        glow: {
          violet:     { value: "rgba(139,92,246,0.45)" },
          violetSoft: { value: "rgba(139,92,246,0.25)" },
          blue:       { value: "rgba(99,102,241,0.45)" },
          magenta:    { value: "rgba(236,72,153,0.45)" },
        },
      },
      gradients: {
        auroraPrimary: { value: "linear-gradient(135deg, {colors.aurora.blue} 0%, {colors.aurora.violet} 50%, {colors.aurora.magenta} 100%)" },
        auroraText:    { value: "linear-gradient(135deg, {colors.aurora.cyan}, {colors.aurora.blue} 40%, {colors.aurora.violet} 70%, {colors.aurora.magenta})" },
        auroraStat:    { value: "linear-gradient(135deg, {colors.aurora.cyan}, {colors.aurora.violet})" },
        conicLogo:     { value: "conic-gradient(from 180deg at 50% 50%, {colors.aurora.cyan}, {colors.aurora.blue}, {colors.aurora.violet}, {colors.aurora.magenta}, {colors.aurora.cyan})" },
      },
      shadows: {
        glassSoft: { value: "0 1px 0 rgba(255,255,255,0.06) inset, 0 8px 32px rgba(0,0,0,0.45)" },
        ctaGlow:   { value: "0 10px 30px {colors.glow.blue}, 0 0 0 1px rgba(255,255,255,0.10) inset" },
        ctaGlowHi: { value: "0 14px 40px {colors.glow.violet}" },
        focusRing: { value: "0 0 0 4px {colors.glow.violetSoft}, 0 0 24px {colors.glow.violetSoft}" },
      },
      radii:    { glass: { value: "16px" }, glassSm: { value: "12px" } },
      durations:{ ui: { value: "180ms" } },
      easings:  { ui: { value: "cubic-bezier(.2,.8,.2,1)" } },
    },
    semanticTokens: {
      colors: {
        bg:              { value: "{colors.canvas.0}" },
        "bg.subtle":     { value: "{colors.canvas.1}" },
        "bg.muted":      { value: "{colors.canvas.2}" },
        "bg.glass":      { value: "{colors.glass.fill}" },
        "bg.glassHi":    { value: "{colors.glass.fillHi}" },
        fg:              { value: "{colors.ink.DEFAULT}" },
        "fg.muted":      { value: "{colors.ink.mute}" },
        "fg.subtle":     { value: "{colors.ink.soft}" },
        border:          { value: "{colors.glass.line}" },
        "border.strong": { value: "{colors.glass.lineHi}" },
        success:         { value: "{colors.aurora.mint}" },
        warning:         { value: "{colors.aurora.amber}" },
        accent:          { value: "{colors.aurora.violet}" },
        accentAlt:       { value: "{colors.aurora.cyan}" },
        danger:          { value: "#F87171" },
      },
    },
  },
});

export const system = createSystem(defaultConfig, config);
```

`components/ui/provider.tsx` — pass `forcedTheme="dark"` to `next-themes` provider; drop any user-facing color-mode toggle.

### 4.3 Layouts

#### `layouts/AuthLayout.tsx`
Used by Login, Signup, ForgotPassword.
```tsx
type AuthLayoutProps = {
  topRight?: React.ReactNode;
  brand: React.ReactNode;
  children: React.ReactNode;
};
```
Renders `<Box>` outer frame (`bg.glass`, border, `glassSoft` shadow, `borderRadius="3xl"`, `overflow="hidden"`), inner `<Grid templateColumns={{base:"1fr", md:"1.05fr 1fr"}}>` with `brand` slot left, form panel right (padded, glass, blur). Mobile collapses to single column, brand panel `minH="320px"`.

#### `layouts/AppShell.tsx`
Used by Welcome, Sessions, Team, Settings.
```tsx
type AppShellProps = {
  active: "overview" | "sessions" | "team" | "settings";
  children: React.ReactNode;
};
```
Renders `<Topbar active={active}/>` + `<Box as="main" p="7" display="grid" gap="6">{children}</Box>`.

### 4.4 Atoms (props summary)

| Component | Props | Behavior |
|-----------|-------|----------|
| `AuroraBackground` | none | `<Box position="fixed" inset="0" zIndex="-2">` with mesh radial-gradient + `auroraFlow` 14s infinite alternate keyframes. Mounted once in `App.tsx` above `<BrowserRouter>`. |
| `BrandMark` | `size?: "sm" \| "md"` | Conic-gradient ring + inner dark box + 16px logo SVG. No animation. |
| `BrandPanel` | `pill, pillTone?, titleLead, titleGrad, sub, features` | Composes orbs + brand-row + hero text + feature list + footer. |
| `GlassCard` | extends `BoxProps` | Box with `bg="bg.glass"`, `borderWidth="1px"`, `borderColor="border"`, `borderRadius="2xl"`, `backdropFilter="blur(20px) saturate(1.15)"`. |
| `FeatureList` | `items: {icon, title, desc}[]` | Vertical stack of glass rows with icon + title + desc. |
| `PasswordField` | `label, value, onChange, error?, autoComplete, name?, placeholder?` | `Field.Root` wrapping `InputGroup` w/ lock startElement + eye `IconButton` endElement. Internal `useState` for show/hide. |
| `PasswordStrength` | `value: string` | Computes score 0–4 (length≥8 + uppercase + digit + symbol). Renders 4 `<Box>` bars; on=auroraStat gradient, off=`bg.glassHi`. `aria-hidden`. |
| `StatCard` | `icon, title, desc, stat, delta, tone?` | GlassCard variant. Stat uses `bgGradient="auroraStat" bgClip="text" color="transparent"`. Delta `color="success"`. |
| `ActivityRow` | `tone, icon, body, time` | Grid `38px 1fr auto` row with toned icon-box. |
| `Topbar` | `active` | `Flex h="16" px="7"` with logo + nav links + notif IconButton + UserChip. Active link highlighted via `aria-current="page"` + glow shadow. |
| `UserChip` | `name, email` | `Menu.Root` wrapping `HStack` trigger (avatar + meta + caret) + `Menu.Content` with `Logout` item that calls `useAuth().logout()`. |

### 4.5 Mock data — `consts/mock.ts`

```ts
import { LuMonitor, LuClock, LuShieldCheck, LuCheck, LuKey, LuUserPlus } from "react-icons/lu";

export const MOCK_STATS = [
  { tone: "cyan",    icon: LuMonitor,    title: "Active sessions", desc: "Devices currently signed in.", stat: "3",   delta: "+1 since yesterday" },
  { tone: "violet",  icon: LuClock,      title: "Token rotations", desc: "Refreshes in last 24h.",        stat: "128", delta: "+12% w/w" },
  { tone: "magenta", icon: LuShieldCheck,title: "Security score",  desc: "Account hardening level.",      stat: "A+",  delta: "2FA enabled" },
] as const;

export const MOCK_ACTIVITY = [
  { tone: "ok",      icon: LuCheck,    body: "Sign-in from MacBook · Safari · Hồ Chí Minh", time: "just now" },
  { tone: "violet",  icon: LuKey,      body: "API key rotated for billing-service",          time: "2h ago" },
  { tone: "magenta", icon: LuUserPlus, body: "Invited thanhlp3@hasaki.vn as Admin",          time: "yesterday" },
] as const;
```

Bold spans inside body: keep as plain string for simplicity; if richer formatting needed, switch to `body: () => JSX`.

### 4.6 Routing — `App.tsx`

```tsx
<>
  <AuroraBackground />
  <BrowserRouter>
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/sso/login" element={<SSOLogin />} />
      <Route path="/signup" element={<Signup />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/welcome"  element={<ProtectedRoute><Welcome /></ProtectedRoute>} />
      <Route path="/sessions" element={<ProtectedRoute><Sessions /></ProtectedRoute>} />
      <Route path="/team"     element={<ProtectedRoute><Team /></ProtectedRoute>} />
      <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
      <Route path="/" element={<Navigate to="/welcome" replace />} />
      <Route path="/404" element={<NotFound />} />
    </Routes>
  </BrowserRouter>
</>
```

### 4.7 Page rewrites

#### `Login.tsx`
- Drop outer gradient Box + center Card.
- `<AuthLayout topRight={<TopLink prompt="New here?" linkText="Create account" to="/signup" />} brand={<BrandPanel ...LOGIN_BRAND_PROPS />}>`.
- Email field: `Field.Root` + `InputGroup startElement={<LuMail/>}` + `Input`.
- Password: `<PasswordField name="password" autoComplete="current-password" />`.
- Row: Remember me Checkbox + `<Link as={RouterLink} to="/forgot-password">Forgot password?</Link>`.
- Submit Button: `bgGradient="auroraPrimary"`, `boxShadow="ctaGlow"`, `_hover={{ boxShadow:"ctaGlowHi" }}`, includes `<LuArrowRight/>` end icon.
- Drop "or" separator + signup Button.
- Keep: `useAuth().login`, `loginSchema`, `toaster.create`.

#### `Signup.tsx`
- Same `AuthLayout`. Brand `pill="Free for solo devs"` mint-tinted.
- `topRight`: `prompt="Already have an account?" linkText="Sign in" to="/login"`.
- Fields: Name (LuUser), Email (LuMail), `<PasswordField>` + `<PasswordStrength value={formData.password}/>` + helper text.
- Drop: T&C Checkbox + `agreeToTerms`/`termsError` state + that branch in `handleSubmit`. Drop bottom "Already have an account?" Link.
- Update `signupSchema` validator: remove any `agreeToTerms` field if present (verify `validators/`).

#### `Welcome.tsx`
- Wrap `<AppShell active="overview">`.
- Body sections:
  1. `<WelcomeHero name={user?.name}/>` — GlassCard with pill (success dot) + h1 (name in `auroraText` gradient) + p + CTA row.
  2. `<Grid templateColumns={{base:"1fr", lg:"repeat(3,1fr)"}} gap="4">` rendering `MOCK_STATS.map(s => <StatCard {...s}/>)`.
  3. Section heading "Recent activity" + `View all →` link + `<GlassCard p="2">` rendering `MOCK_ACTIVITY.map(a => <ActivityRow {...a}/>)`.
- Keep: `useEffect getCurrentUser`, loading Spinner state, error state (skinned with aurora colors).
- Remove: top-right Logout Button (moved into `UserChip` Menu).

#### `ForgotPassword.tsx` (NEW)
- `<AuthLayout topRight={<TopLink prompt="Remember it?" linkText="Sign in" to="/login"/>} brand={<BrandPanel pill="Account recovery" .../>}>`.
- Form: Email Field + Submit `Send reset link` Button.
- onSubmit: `e.preventDefault(); toaster.create({ type:"info", title:"Coming soon", description:"Reset flow not implemented yet." });`. No backend call.

#### `Sessions.tsx` / `Team.tsx` / `Settings.tsx` (NEW)
- `<AppShell active="sessions"|"team"|"settings">`.
- Body: `<GlassCard p="9" textAlign="center">` with icon (LuMonitor/LuUsers/LuSettings), `<Heading>{Page}</Heading>`, `<Text color="fg.muted">Coming soon.</Text>`.
- Wrapped in `<ProtectedRoute>`.

#### `SSOLogin.tsx`
- Keep current single-card layout.
- Replace `bgGradient brand.50/200/950/900` → `bg="bg"`.
- Replace Button `bgGradient brand.500/600` → `bgGradient="auroraPrimary" boxShadow="ctaGlow"`.
- Replace text `brand.*` → `accent`.

#### `NotFound.tsx` / `Home.tsx`
- NotFound: `bg="bg"`, `fg` headings, Button `bgGradient="auroraPrimary"`.
- Home: read; swap any `brand.*` if present.

### 4.8 `index.css`

```css
@keyframes auroraFlow {
  0%   { background-position:   0% 0%, 100% 0%, 100% 100%,   0% 100%; }
  50%  { background-position:  30% 20%, 60% 30%,  70% 70%,  30% 60%; }
  100% { background-position: 100% 100%, 0% 100%,   0% 0%, 100% 0%; }
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after { animation: none !important; transition: none !important; }
}
html, body { background: #07071A; color: #F4F5FF; }
```

## 5. Chakra v3 mapping cheatsheet

Mock HTML primitive → Chakra v3 equivalent.

| Mockup primitive | Chakra v3 |
|------------------|-----------|
| `body::before` aurora flow | `<AuroraBackground>` = `Box` `position="fixed" inset="0" zIndex="-2"` + `css={{ animation: "auroraFlow 14s ease-in-out infinite alternate" }}` |
| `.frame` outer border | `Box bg="bg.glass" borderColor="border" borderWidth="1px" borderRadius="3xl" boxShadow="glassSoft" overflow="hidden"` |
| `.split` 1.05fr 1fr grid | `<Grid templateColumns={{ base:"1fr", md:"1.05fr 1fr" }} minH="720px">` |
| `.brand-panel` w/ orbs | `<Box position="relative" p="11">` + 3 absolute `<Box>` orbs `filter="blur(40px)"` |
| `.brand-mark` conic | `Box w="10" h="10" rounded="xl" bgGradient="conicLogo"` |
| `.brand-h1 .grad` | `<Text as="span" bgGradient="auroraText" bgClip="text" color="transparent">` |
| `.pill` | `<HStack gap="2" px="3" py="1.5" rounded="full" bg="bg.glass" borderWidth="1px" borderColor="border.strong" backdropFilter="blur(20px)">` |
| `.feature` glass row | `<HStack p="3.5" bg="bg.glass" borderWidth="1px" borderColor="border" borderRadius="2xl" backdropFilter="blur(20px) saturate(1.2)">` |
| `.input-wrap` w/ icon | `<Field.Root>` + `<InputGroup startElement={<Icon/>} endElement={...}>` + `<Input bg="bg.glass" borderColor="border.strong" _focus={{ borderColor:"accent", boxShadow:"focusRing" }}>` |
| `.toggle-eye` | `<IconButton variant="ghost" size="xs" aria-label="Show password" onClick={toggle}>` |
| `.btn-primary` | `<Button bgGradient="auroraPrimary" color="white" boxShadow="ctaGlow" _hover={{ boxShadow:"ctaGlowHi" }}>` |
| `.btn-ghost` | `<Button variant="outline" bg="bg.glass" borderColor="border.strong" backdropFilter="blur(12px)">` |
| `.check input` checkbox | `<Checkbox.Root>` + `<Checkbox.Control _checked={{ bgGradient:"auroraPrimary", borderColor:"accent" }}>` |
| `.app-bar` | `<Flex h="16" px="7" borderBottomWidth="1px" borderColor="border" bg="rgba(7,7,26,0.55)" backdropFilter="blur(20px) saturate(1.2)">` |
| `.nav a.active` | Link with `aria-current="page"`, conditional `bg="bg.glass"` + glow `boxShadow` |
| `.icon-btn` w/ badge | `<IconButton position="relative" aria-label="...">` + absolute `<Box>` badge dot |
| `.user-chip` | `<Menu.Root>` + `<Menu.Trigger asChild><HStack as="button">…</HStack></Menu.Trigger>` + `<Menu.Content>` w/ Logout `Menu.Item` |
| `.welcome-hero` | `<GlassCard p="9" position="relative" overflow="hidden">` + 3 absolute orb Boxes |
| `.stat .grad` | `<Heading bgGradient="auroraStat" bgClip="text" color="transparent">` |
| `.activity` list | `<GlassCard p="2">` containing `<Stack divideY="1px" divideColor="border">` of Grid rows `gridTemplateColumns="38px 1fr auto"` |
| `.strength span.on/off` | `<HStack gap="1.5">` + four `<Box flex="1" h="1" rounded="full">` with `bgGradient` if on else `bg="bg.glassHi"` |

**v3 specifics**:
- `bgGradient="tokenName"` — reads gradient token directly; no `gradientFrom/to` props.
- `Field.Root invalid={!!error}` + `Field.Label` + `Field.ErrorText` (already in current Login/Signup).
- `Checkbox.Root` namespace; existing `components/ui/checkbox.tsx` wrapper compatible.
- `Menu.Root/Trigger/Content/Item` replaces v2 `Menu+MenuButton+MenuList+MenuItem`.
- `IconButton aria-label` required.
- Icons via `react-icons/lu` (Lucide). Used: `LuMail`, `LuLock`, `LuUser`, `LuEye`, `LuEyeOff`, `LuArrowRight`, `LuClock`, `LuShieldCheck`, `LuMonitor`, `LuKey`, `LuUserPlus`, `LuBell`, `LuChevronDown`, `LuPlus`, `LuFileText`, `LuCheck`, `LuUsers`, `LuSettings`.

## 6. Verification

### Build/test gate
- `cd ui && npm run build` — pass tsc + vite build.
- `cd ui && npm run lint` — pass.
- Manual smoke (`make run-ui`):
  1. `/login` — split layout renders, aurora bg flows, eye toggle works, "Forgot password?" navigates.
  2. `/signup` — strength meter updates per keystroke, no T&C, topRight `Sign in` link works.
  3. `/forgot-password` — submit shows `Coming soon` toast.
  4. Login w/ valid creds → `/welcome` — topbar nav + stats + activity render.
  5. Click `Sessions` / `Team` / `Settings` nav → stub pages render with shared topbar.
  6. UserChip menu → Logout → returns to `/login`.
  7. `/sso/login` — single card layout intact, aurora colors applied.
  8. `/404` — aurora bg, restyled.

### A11y
- IconButtons have `aria-label`.
- `prefers-reduced-motion` disables flow animation.
- Field errors via `Field.ErrorText`.
- Focus ring visible on Inputs + Buttons.
- Touch targets ≥44px (Buttons `h="12"`, IconButtons `boxSize="10"`).
- Foreground/background contrast ≥ 4.5:1.

## 7. Rollout sequence (commit per step)

1. `chore(ui): rewrite theme to aurora dark` — theme + index.css + provider forced dark.
2. `feat(ui): add aurora layouts and atoms` — AuthLayout, AppShell, all atoms, mock data.
3. `feat(ui): redesign Login + Signup + Welcome`.
4. `feat(ui): add forgot-password + sessions/team/settings stubs and routes`.
5. `chore(ui): retheme SSOLogin + NotFound + Home`.
6. `chore: build verify + manual smoke`.

## 8. Risks

- `forcedTheme="dark"` on `next-themes` may conflict with `ColorModeProvider`. Audit `components/ui/provider.tsx` + `components/ui/color-mode.tsx`; remove user-facing toggle if any.
- `bgGradient="conicLogo"` token — verify Chakra v3 supports `conic-gradient` value in gradient tokens. Fallback: use `css={{ background: "conic-gradient(...)" }}` directly.
- Safari `-webkit-backdrop-filter` — Chakra Box passes through; verify in build.
- `Menu.Root` API in v3 — confirm shape matches existing wrappers in `components/ui/`.

## 9. Out of scope

- Light mode toggle.
- Real `/sessions` `/activity` Go endpoints.
- Forgot-password backend (email + token).
- Backend domain code changes.
