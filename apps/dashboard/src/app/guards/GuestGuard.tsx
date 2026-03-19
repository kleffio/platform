import { useEffect, type ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth as useOidcAuth } from "react-oidc-context";
import { Skeleton } from "@kleff/ui";

interface GuestGuardProps {
  children: ReactNode;
}

/**
 * Protects guest-only routes (sign-in page, landing page, etc.).
 * - Redirects authenticated users to the dashboard.
 * - Renders children for unauthenticated users.
 */
export function GuestGuard({ children }: GuestGuardProps) {
  const auth = useOidcAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (auth.isAuthenticated) {
      navigate("/dashboard", { replace: true });
    }
  }, [auth.isAuthenticated, navigate]);

  if (auth.isLoading || auth.isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-950">
        <div className="w-48 space-y-3">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
