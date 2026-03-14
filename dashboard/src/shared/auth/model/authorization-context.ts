import { createContext } from "react";
import type { AuthorizationContextValue } from "./types";

export const AuthorizationContext =
  createContext<AuthorizationContextValue | undefined>(undefined);
