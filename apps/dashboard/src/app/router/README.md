# src/app/router/

Centralised routing configuration. All routes, route constants, and the router provider live here.

## Exports

| Export | Description |
|---|---|
| `ROUTES` | `as const` map of every route path in the app — import this instead of hardcoding strings |
| `RouterProvider` | Drop-in wrapper that provides the React Router context to the app |

## Route groups

```
/                    → redirects to /dashboard
/dashboard           → AuthGuard  → DashboardPage
/components          → GuestGuard → ComponentsPage  (showcase, visible to guests only)
/auth/signin         → (future) GuestGuard → SignInPage
/auth/callback       → (future) OIDC callback handler
/error/deactivated   → (future) account deactivated error page
```

## Rules

- Always reference paths via `ROUTES.*` — never hardcode `/dashboard` anywhere else.
- Lazy-load heavy pages with `React.lazy` + `Suspense` as the app grows.
- Guard each route group at the layout level, not inside individual page components.
