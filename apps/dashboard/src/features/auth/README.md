# src/features/auth/

OIDC authentication, token management, and role-based authorization.

## Subfolders

| Folder | Contents |
|---|---|
| `config/` | `oidc.ts` — OIDC provider settings (authority, client ID, redirect URIs) read from `VITE_*` env vars |
| `context/` | `AuthorizationContext.ts` — React context type for the user's role and permissions |
| `hooks/` | `useAuth` — wraps `react-oidc-context`; `useAuthorization` — reads role/permission context |
| `providers/` | `AuthProvider` — composes OIDC + token sync + authorization; `AuthorizationProvider` — resolves roles |
| `types/` | `AuthorizationContextValue` and related TypeScript types |

## How it fits together

```
AuthProvider (features/auth)
  └─ OidcProvider          ← react-oidc-context, configured via config/oidc.ts
      └─ AuthTokenSync     ← keeps in-memory access token in sync with OIDC state
          └─ AuthorizationProvider  ← fetches user roles and exposes them via context
              └─ children
```

`AuthProvider` is mounted by `main.tsx` around the whole app.

## Token flow

1. OIDC sign-in completes → `AuthTokenSync` calls `setApiAccessToken(token)`.
2. `shared/api/client.ts` request interceptor reads the token and adds `Authorization: Bearer <token>` to every API call.
3. On sign-out or session expiry → `clearApiAccessToken()` is called and the `apiClient` stops sending the header.
4. If a `401` response arrives → `app/interceptors/authInterceptor.ts` triggers `signinRedirect()`.

## Public API (`index.ts`)

```ts
import { AuthProvider, useAuth, useAuthorization } from "@/features/auth";
import type { AuthorizationContextValue } from "@/features/auth";
```

Guards (`AuthGuard`, `GuestGuard`) live in `app/guards/` — not here — because they are routing infrastructure, not domain logic.
