# src/app/providers/

React context providers that wrap the entire application tree. Rendered in `main.tsx` via `<AppProvider>`.

## Providers

| File | Provider | Purpose |
|---|---|---|
| `query-provider.tsx` | `QueryClientProvider` | TanStack Query client — server state, caching, refetching |
| `theme-provider.tsx` | `ThemeProvider` | Light/dark mode via `next-themes` or equivalent |
| `toast-provider.tsx` | `Toaster` | Sonner toast notifications |
| `index.tsx` | `AppProvider` | Composes all of the above into a single wrapper |

`AuthProvider` lives in `features/auth/providers/` rather than here because it owns OIDC state and is a domain concern.

## Adding a provider

1. Create `my-provider.tsx` in this folder.
2. Export a component that wraps `children`.
3. Import and nest it inside `AppProvider` in `index.tsx`.
