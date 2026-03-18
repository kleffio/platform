import { useEffect, type ReactNode } from "react";
import { AuthProvider as OidcProvider } from "react-oidc-context";
import { useAuth as useOidcAuth } from "react-oidc-context";
import { clearApiAccessToken, setApiAccessToken } from "@/shared/api";
import { oidcConfig } from "@/shared/config/oidc";
import { AuthorizationProvider } from "./AuthorizationProvider";

interface Props {
  children: ReactNode;
}

function AuthTokenSync() {
  const auth = useOidcAuth();

  useEffect(() => {
    if (auth.isAuthenticated && auth.user?.access_token) {
      setApiAccessToken(auth.user.access_token);
      return;
    }

    clearApiAccessToken();
  }, [auth.isAuthenticated, auth.user?.access_token]);

  return null;
}

/**
 * Combined AuthProvider
 * Wraps the app in both OIDC and application-specific Authorization logic.
 */
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
