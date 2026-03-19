import type { ReactNode } from "react";

interface AppLayoutProps {
  children: ReactNode;
}

// Placeholder — add sidebar, topbar, and shell chrome here as the app grows.
export function AppLayout({ children }: AppLayoutProps) {
  return <>{children}</>;
}
