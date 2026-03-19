# src/layouts/

Page shell components. A layout wraps one or more pages and provides persistent chrome — sidebar, topbar, breadcrumbs, navigation.

## Current layouts

| Folder | Description |
|---|---|
| `AppLayout/` | Main authenticated shell — sidebar nav, topbar, content area. Currently a passthrough placeholder. |

## Usage

Layouts are composed into route groups in the router, not inside individual pages:

```tsx
// router/index.tsx
{
  element: <AuthGuard><AppLayout><Outlet /></AppLayout></AuthGuard>,
  children: [
    { path: ROUTES.DASHBOARD, element: <DashboardPage /> },
    { path: ROUTES.SERVERS,   element: <ServersPage /> },
  ],
}
```

## Rules

- Layouts only handle **chrome** (nav, sidebars, headers). No business logic.
- Layouts receive `children` or render `<Outlet />` — they never know which page is inside them.
- Feature-specific panels (e.g. a server detail drawer) are not layouts — they belong in `features/`.
