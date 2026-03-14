import type { ReactNode } from "react";
import { AuthProvider as OidcProvider } from "react-oidc-context";
import { oidcConfig } from "@/shared/config/oidc";
import { AuthorizationProvider } from "./AuthorizationProvider";

interface Props {
  children: ReactNode;
}

/**
 * Combined AuthProvider
 * Wraps the app in both OIDC and application-specific Authorization logic.
 */
export function AuthProvider({ children }: Props) {
  return (
    <OidcProvider {...oidcConfig}>
      <AuthorizationProvider>
        {children}
      </AuthorizationProvider>
    </OidcProvider>
  );
}
