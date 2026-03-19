# src/pages/

One top-level component per route. Pages are thin assemblers — they import from `features/`, `layouts/`, and `shared/`, then compose them into a full screen.

## Current pages

| Folder | Route | Guard | Description |
|---|---|---|---|
| `dashboard/` | `/dashboard` | `AuthGuard` | Main app dashboard — server overview, metrics, billing summary |
| `components/` | `/components` | `GuestGuard` | Component showcase — demonstrates every `@kleff/ui` component variant |

## Rules

- **No business logic in pages.** Data fetching, state, and side-effects belong in feature hooks.
- **One folder per route.** If a route has sub-routes, nest them inside the same folder.
- **Name the file after the route.** `pages/dashboard/DashboardPage.tsx`, not `pages/dashboard/index.tsx` — named files are easier to find in editor tabs.
- Pages must not import from other pages.

## Adding a new page

1. Create `pages/<name>/<Name>Page.tsx`.
2. Add the route path to `ROUTES` in `app/router/index.tsx`.
3. Add the route entry (with appropriate guard) to the router config.
