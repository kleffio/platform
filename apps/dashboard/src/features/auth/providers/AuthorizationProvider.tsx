import type { ReactNode } from "react";
import { AuthorizationContext } from "../context/AuthorizationContext";
import type { AuthorizationContextValue } from "../types";

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
