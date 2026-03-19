import { useEffect, type ReactNode } from "react";
import { AuthProvider as OidcProvider, useAuth as useOidcAuth } from "react-oidc-context";
import { clearApiAccessToken, setApiAccessToken } from "@/shared/api";
import { oidcConfig } from "../config/oidc";
import { registerAuthHandlers } from "@/app/interceptors/authInterceptor";
import { AuthorizationProvider } from "./AuthorizationProvider";

// Activate the auth interceptor — registered once when this module is imported.
import "@/app/interceptors/authInterceptor";

interface Props {
  children: ReactNode;
}

function AuthTokenSync() {
  const auth = useOidcAuth();

  // Wire OIDC callbacks into the HTTP interceptor so API 401s
  // trigger a sign-in redirect automatically.
  useEffect(() => {
    registerAuthHandlers({
      onUnauthenticated: () => auth.signinRedirect(),
    });
  }, [auth]);

  // Keep the in-memory access token in sync with OIDC state.
  useEffect(() => {
    if (auth.isAuthenticated && auth.user?.access_token) {
      setApiAccessToken(auth.user.access_token);
      return;
    }
    clearApiAccessToken();
  }, [auth.isAuthenticated, auth.user?.access_token]);

  return null;
}

export function AuthProvider({ children }: Props) {
  return (
    <OidcProvider {...oidcConfig}>
      <AuthTokenSync />
      <AuthorizationProvider>
        {children}
      </AuthorizationProvider>
    </OidcProvider>
  );
}
