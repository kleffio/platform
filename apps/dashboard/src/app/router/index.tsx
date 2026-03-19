import { createBrowserRouter, RouterProvider as ReactRouterProvider, Navigate, Outlet } from "react-router-dom";
import { AuthGuard } from "@/app/guards/AuthGuard";
import { GuestGuard } from "@/app/guards/GuestGuard";
import DashboardPage from "@/pages/dashboard/DashboardPage";
import ComponentsPage from "@/pages/components/ComponentsPage";
import { ROUTES } from "./routes";

export { ROUTES } from "./routes";

const router = createBrowserRouter([
  {
    path: ROUTES.HOME,
    element: <Navigate to={ROUTES.DASHBOARD} replace />,
  },
  {
    // Authenticated routes — requires a signed-in user.
    element: <AuthGuard><Outlet /></AuthGuard>,
    children: [
      {
        path: ROUTES.DASHBOARD,
        element: <DashboardPage />,
      },
    ],
  },
  {
    // Guest-only routes — redirects authenticated users to the dashboard.
    element: <GuestGuard><Outlet /></GuestGuard>,
    children: [
      {
        path: ROUTES.COMPONENTS,
        element: <ComponentsPage />,
      },
    ],
  },
]);

export function RouterProvider() {
  return <ReactRouterProvider router={router} />;
}
