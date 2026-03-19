# src/

Application source root. Every file in here is compiled by Vite/TypeScript.

## Folder map

| Folder | Purpose |
|---|---|
| `app/` | App-level wiring — routing, guards, interceptors, providers, global styles |
| `features/` | Self-contained domain slices (auth, servers, billing, …) |
| `layouts/` | Page shell components — sidebar, topbar, nav chrome |
| `pages/` | One top-level component per route, composed from features + layouts |
| `shared/` | Cross-cutting utilities used by more than one feature (API client, hooks, utils) |

## Rules

- **Pages are thin.** A page imports from `features/` and `layouts/`; it does not contain its own business logic.
- **Features are self-contained.** A feature may import from `shared/` but never from another feature or from `pages/`.
- **`shared/` has no domain knowledge.** Nothing in `shared/` should import from `features/` or `pages/`.
- **`app/` bootstraps the app.** Routing, providers, guards, and interceptors live here — not in features.
