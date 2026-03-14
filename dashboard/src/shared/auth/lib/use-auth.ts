import { useAuth as useOidcAuth } from "react-oidc-context";

export function useAuth() {
  return useOidcAuth();
}
