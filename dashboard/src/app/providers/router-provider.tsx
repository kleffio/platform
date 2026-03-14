import { createBrowserRouter, RouterProvider as ReactRouterProvider, Navigate } from "react-router-dom";
import DashboardPage from "@/pages/dashboard/DashboardPage";

/**
 * ROUTES constant mirroring the prototype's structure.
 * This is flattened into the provider file as per the "no-folder" FSD request.
 */
export const ROUTES = {
  HOME: "/",
  DASHBOARD: "/dashboard",
  AUTH_CALLBACK: "/auth/callback",
  AUTH_SIGNIN: "/auth/signin",
  ERROR_DEACTIVATED: "/error/deactivated",
} as const;

const router = createBrowserRouter([
  {
    path: ROUTES.HOME,
    element: <Navigate to={ROUTES.DASHBOARD} replace />,
  },
  {
    path: ROUTES.DASHBOARD,
    element: <DashboardPage />,
  },
  // Add other mirrored routes from prototype as needed
]);

export function RouterProvider() {
  return <ReactRouterProvider router={router} />;
}
