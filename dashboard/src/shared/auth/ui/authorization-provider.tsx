import type { ReactNode } from "react";
import { AuthorizationContext } from "../model/authorization-context";
import type { AuthorizationContextValue } from "@/shared/auth";

interface Props {
  children: ReactNode;
}

export function AuthorizationProvider({ children }: Props) {
  const value: AuthorizationContextValue = {
    shadowMode: false,
    enforceMode: true,
    isLoading: false,
    error: null,
  };

  return (
    <AuthorizationContext.Provider value={value}>
      {children}
    </AuthorizationContext.Provider>
  );
}
