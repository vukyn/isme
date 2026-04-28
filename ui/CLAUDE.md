# UI scope — Chakra UI v3

Stack: React 19 + Vite 7 + Chakra UI **v3.30** + `@emotion/react` + `next-themes` + `react-icons` + `react-router-dom` v7 + `axios`.

Source rule: `.cursor/rules/chakra-v3.mdc`. This file is the canonical version Claude Code reads.

## Imports

| From | Symbols |
|------|---------|
| `@chakra-ui/react` | `Alert`, `Avatar`, `Button`, `Card`, `Field`, `Table`, `Input`, `NativeSelect`, `Tabs`, `Textarea`, `Separator`, `Box`, `Flex`, `Stack`, `HStack`, `VStack`, `Text`, `Heading`, `Icon`, `Dialog`, `Menu`, `Popover`, `useDisclosure`, `useChakra` |
| `@/components/ui/*` (relative) | `Provider`, `Toaster`, `toaster`, `ColorModeProvider`, `Tooltip`, `PasswordInput`, `Button`, `Card`, `Checkbox`, `Input`, `Link` |

Existing `components/ui/`: `button.tsx`, `card.tsx`, `checkbox.tsx`, `color-mode.tsx`, `input.tsx`, `link.tsx`, `provider.tsx`, `toaster.tsx`, `tooltip.tsx`. Reuse these — no fresh wrappers.

## Banned packages (v2 leftovers)

- `@emotion/styled` — gone. Only `@emotion/react`.
- `framer-motion` — gone.
- `@chakra-ui/icons` → use `react-icons` (project pick, not `lucide-react`).
- `@chakra-ui/hooks` → use `react-use` or `usehooks-ts`.
- `@chakra-ui/next-js` → use `asChild` prop.

## Renames (v2 → v3)

### Components
`Modal → Dialog`, `Divider → Separator`, `Collapse → Collapsible`, `Tags → Badge`, `useToast → toaster.create()` (from `components/ui/toaster`), `Select → NativeSelect`.

### Boolean props
`isOpen→open`, `isDisabled→disabled`, `isInvalid→invalid`, `isRequired→required`, `isLoading→loading`, `isChecked→checked`, `isIndeterminate→indeterminate`, `isActive→data-active`.

### Style props
`colorScheme→colorPalette`, `spacing→gap`, `noOfLines→lineClamp`, `truncated→truncate`, `thickness→borderWidth`, `speed→animationDuration`.

## Component shapes (compound API)

All complex components now namespaced: `X.Root`, `X.Trigger`, `X.Content`, etc.

### Toast
```tsx
import { toaster } from "@/components/ui/toaster"

toaster.create({
  title: "Title",
  type: "error",                 // was status
  meta: { closable: true },      // was isClosable
  placement: "top-end",          // was top-right
})
```

### Dialog (was Modal)
```tsx
<Dialog.Root open={isOpen} onOpenChange={onOpenChange} placement="center">
  <Dialog.Backdrop />
  <Dialog.Content>
    <Dialog.Header><Dialog.Title>Title</Dialog.Title></Dialog.Header>
    <Dialog.Body>Content</Dialog.Body>
  </Dialog.Content>
</Dialog.Root>
```

### Button icons — children, not props
```tsx
<Button><Mail /> Email <ChevronRight /></Button>
// NOT leftIcon / rightIcon
```

### Alert
```tsx
<Alert.Root borderStartWidth="4px" borderStartColor="colorPalette.solid">
  <Alert.Indicator />
  <Alert.Content>
    <Alert.Title>Title</Alert.Title>
    <Alert.Description>Description</Alert.Description>
  </Alert.Content>
</Alert.Root>
```

### Tooltip (project wrapper)
```tsx
import { Tooltip } from "@/components/ui/tooltip"

<Tooltip content="Content" showArrow positioning={{ placement: "top" }}>
  <Button>Hover me</Button>
</Tooltip>
```

### Field + Input validation
```tsx
<Field.Root invalid>
  <Field.Label>Email</Field.Label>
  <Input />
  <Field.ErrorText>This field is required</Field.ErrorText>
</Field.Root>
```

### Table
`Table.Root` / `Table.Header` / `Table.Body` / `Table.Row` / `Table.ColumnHeader` / `Table.Cell`. Variant `line` (was `simple`).

### Tabs
```tsx
<Tabs.Root defaultValue="one" colorPalette="orange">
  <Tabs.List><Tabs.Trigger value="one">One</Tabs.Trigger></Tabs.List>
  <Tabs.Content value="one">Content</Tabs.Content>
</Tabs.Root>
```

### Menu
```tsx
<Menu.Root>
  <Menu.Trigger asChild><Button>Actions</Button></Menu.Trigger>
  <Menu.Content><Menu.Item value="download">Download</Menu.Item></Menu.Content>
</Menu.Root>
```

### Popover
```tsx
<Popover.Root positioning={{ placement: "bottom-end" }}>
  <Popover.Trigger asChild><Button>Click</Button></Popover.Trigger>
  <Popover.Content>
    <Popover.Arrow />
    <Popover.Body>Content</Popover.Body>
  </Popover.Content>
</Popover.Root>
```

### NativeSelect (was Select)
```tsx
<NativeSelect.Root size="sm">
  <NativeSelect.Field placeholder="Select option">
    <option value="1">Option 1</option>
  </NativeSelect.Field>
  <NativeSelect.Indicator />
</NativeSelect.Root>
```

## Style system

### Nested selectors — `css` not `sx`, `&` required
```tsx
<Box css={{ "& svg": { color: "red.500" } }} />
```

### Gradients — split props
```tsx
<Box bgGradient="to-r" gradientFrom="red.200" gradientTo="pink.500" />
```

### Theme tokens — `useChakra().token()`
```tsx
const system = useChakra()
const gray400 = system.token("colors.gray.400")
```

## Quick checklist when writing/reviewing `.tsx`

1. No banned packages imported.
2. Boolean props use new names (`open`, `disabled`, `invalid`, `loading`, ...).
3. `colorPalette` not `colorScheme`; `gap` not `spacing`.
4. Compound API used (`Dialog.Root`, `Tabs.Root`, ...).
5. Button icons inline as children.
6. Toast via `toaster.create()` from `@/components/ui/toaster`.
7. `css={{ "& selector": ... }}` for nested styles, never `sx`.
8. Theme access via `useChakra().token(...)`.
