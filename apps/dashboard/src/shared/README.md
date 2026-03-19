# src/shared/

Cross-cutting utilities and infrastructure used by more than one feature. Nothing in `shared/` has domain knowledge — it does not know about servers, billing, or users specifically.

## Subfolders

| Folder | Purpose |
|---|---|
| `api/` | Axios client, token storage, error types, typed request helpers |
| `hooks/` | _(scaffolded)_ Generic reusable hooks (e.g. `useDebounce`, `useLocalStorage`, `useMediaQuery`) |
| `utils/` | _(scaffolded)_ Pure utility functions with no React dependency (e.g. formatters, validators) |

## Rules

- `shared/` may import from `packages/` (`@kleff/ui`, `@kleff/shared-types`) but never from `features/` or `pages/`.
- If a utility is only used by one feature, keep it inside that feature — don't promote it to `shared/` prematurely.
- No React components belong here — UI primitives live in `packages/ui`, feature-specific components live in their feature folder.
