# src/app/

App-level wiring. Everything here runs once at startup and is not tied to any specific domain feature.

## Subfolders

| Folder | Purpose |
|---|---|
| `guards/` | Route-level access control components (`AuthGuard`, `GuestGuard`) |
| `interceptors/` | Axios request/response interceptors registered at app boot |
| `providers/` | React context providers wrapping the whole app (query client, theme, toasts) |
| `router/` | Route definitions and the `ROUTES` constants map |
| `styles/` | Global CSS entry point — imports design tokens, Tailwind directives |

## What belongs here

- Things that wrap the entire application tree.
- Infrastructure that every feature depends on (routing, HTTP, auth state sync).
- Nothing domain-specific — no server lists, no billing logic, no user profile UI.
