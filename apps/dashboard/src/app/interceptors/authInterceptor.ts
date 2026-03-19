import { apiClient, isApiError } from "@/shared/api";

type VoidFn = () => void;

// Handlers are registered at runtime by AuthProvider once OIDC is available.
const handlers: {
  onUnauthenticated?: VoidFn;
  onForbidden?: VoidFn;
} = {};

export function registerAuthHandlers(h: typeof handlers) {
  Object.assign(handlers, h);
}

// Registered once at module load — runs after the existing normalisation
// interceptor, so `error` is already an ApiError with a `.status` field.
apiClient.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (isApiError(error)) {
      if (error.status === 401) handlers.onUnauthenticated?.();
      if (error.status === 403) handlers.onForbidden?.();
    }
    return Promise.reject(error);
  }
);
