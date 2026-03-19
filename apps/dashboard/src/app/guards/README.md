# src/app/guards/

Route guard components. Wrap route groups in the router to enforce access rules before rendering any page.

## Guards

### `AuthGuard`
Protects routes that require a signed-in user.

- While OIDC is initialising → shows a loading skeleton.
- If there is an auth error → shows an inline error message.
- If the user is not authenticated → calls `signinRedirect()` and shows the loading skeleton.
- Once authenticated → renders `children` (the route's `<Outlet />`).

### `GuestGuard`
Protects routes that should only be visible to unauthenticated users (e.g. sign-in page, component showcase).

- If the user is authenticated → redirects to `/dashboard`.
- While OIDC is loading, or while redirect is pending → shows a loading skeleton.
- Otherwise → renders `children`.

## Usage in the router

```tsx
// Authenticated group
{ element: <AuthGuard><Outlet /></AuthGuard>, children: [...] }

// Guest-only group
{ element: <GuestGuard><Outlet /></GuestGuard>, children: [...] }
```

## Rules

- Guards only read auth state — they never mutate it.
- The loading skeleton must match the full-screen background so there is no flash.
- Do not import guards from `features/` — guards are app-level infrastructure.
