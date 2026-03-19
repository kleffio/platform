# src/app/interceptors/

Axios interceptors registered once at application boot via side-effect imports.

## Interceptors

### `authInterceptor.ts`
Listens on the shared `apiClient` response pipeline for HTTP `401` and `403` errors.

- **401 Unauthenticated** → calls the registered `onUnauthenticated` handler (triggers OIDC `signinRedirect`).
- **403 Forbidden** → calls the registered `onForbidden` handler (can show a "no access" UI).

Handlers are injected at runtime by `AuthProvider` once OIDC is ready:

```ts
registerAuthHandlers({
  onUnauthenticated: () => auth.signinRedirect(),
});
```

This runs *after* the error-normalisation interceptor in `shared/api/client.ts`, so by the time this interceptor fires the error is already a typed `ApiError` with a `.status` field.

## Interceptor execution order

1. `shared/api/client.ts` — attaches `Bearer` token to every request; normalises error shapes on response.
2. `app/interceptors/authInterceptor.ts` — reacts to `401`/`403` with OIDC callbacks.

## Adding a new interceptor

Create a new `.ts` file here, register it on `apiClient`, and import it as a side-effect in the appropriate provider or in `main.tsx`.
