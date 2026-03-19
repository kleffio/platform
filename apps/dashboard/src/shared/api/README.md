# src/shared/api/

The HTTP layer. Everything that touches the network lives here or in a feature's `api/` subfolder.

## Files

| File | Purpose |
|---|---|
| `client.ts` | Axios instance with base URL, request interceptor (attaches `Bearer` token), response interceptor (normalises errors) |
| `config.ts` | API base URL read from `VITE_API_BASE_URL` env var |
| `error.ts` | `ApiError` class and `isApiError()` type guard |
| `request.ts` | Typed wrapper functions (`get`, `post`, `put`, `patch`, `del`) — use these instead of calling `apiClient` directly |
| `token.ts` | In-memory access token store (`setApiAccessToken`, `clearApiAccessToken`, `getApiAccessToken`) |
| `index.ts` | Re-exports everything above |

## Interceptor chain (on `apiClient`)

1. **Request** — reads the in-memory token from `token.ts` and sets `Authorization: Bearer <token>`.
2. **Response (success)** — passes through unchanged.
3. **Response (error)** — catches Axios errors and re-throws as a typed `ApiError` with `.status`, `.message`, and `.data` fields.
4. **Auth interceptor** (`app/interceptors/authInterceptor.ts`, registered separately) — handles `401` → sign-in redirect, `403` → forbidden handler.

## Making an API call

```ts
import { get, post } from "@/shared/api";

const server = await get<GameServer>("/gameservers/abc-123");
const created = await post<GameServer>("/gameservers", { name: "my-server", region: "eu-west" });
```

## Token lifecycle

`AuthTokenSync` (inside `AuthProvider`) calls `setApiAccessToken` when OIDC authenticates and `clearApiAccessToken` on sign-out. The token is never stored in `localStorage` or cookies — it lives only in memory for this tab's lifetime.
