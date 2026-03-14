import type { ReactNode } from "react";
import { AuthProvider as SharedAuthProvider } from "@/features/auth";

interface Props {
  children: ReactNode;
}

export function AuthProvider({ children }: Props) {
  return <SharedAuthProvider>{children}</SharedAuthProvider>;
}
