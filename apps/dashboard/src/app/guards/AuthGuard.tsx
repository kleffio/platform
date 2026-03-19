import { useEffect, useState, type ReactNode } from "react";
import { useAuth as useOidcAuth } from "react-oidc-context";
import { Skeleton } from "@kleff/ui";

interface AuthGuardProps {
  children: ReactNode;
}

/**
 * Protects routes that require authentication.
 * - Shows a loading state while OIDC is initialising.
 * - Triggers an OIDC redirect if the user is not signed in.
 * - Renders children once the user is authenticated.
 */
export function AuthGuard({ children }: AuthGuardProps) {
  const auth = useOidcAuth();
  const [redirecting, setRedirecting] = useState(false);

  useEffect(() => {
    if (!auth.isLoading && !auth.error && !auth.isAuthenticated && !redirecting) {
      setRedirecting(true);
      auth.signinRedirect();
    }
  }, [auth, redirecting]);

  if (auth.isLoading || redirecting) {
    return <AuthLoadingScreen />;
  }

  if (auth.error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-center space-y-2">
          <p className="text-sm font-medium text-red-400">Authentication error</p>
          <p className="text-xs text-zinc-500">{auth.error.message}</p>
        </div>
      </div>
    );
  }

  if (!auth.isAuthenticated) {
    return <AuthLoadingScreen />;
  }

  return <>{children}</>;
}

function AuthLoadingScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-48 space-y-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
      </div>
    </div>
  );
}
