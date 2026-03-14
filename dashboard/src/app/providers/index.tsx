import type { ReactNode } from "react";
import { AuthProvider } from "@/app/providers/auth-provider";
import { QueryProvider } from "@/app/providers/query-provider";
import { RouterProvider } from "@/app/providers/router-provider";
import { ThemeProvider } from "@/app/providers/theme-provider";
import { ToastProvider } from "@/app/providers/toast-provider";

export function AppProvider({ children }: { children: ReactNode }) {
    return (
        <ThemeProvider defaultTheme="dark" storageKey="kleff-ui-theme">
            <QueryProvider>
                <AuthProvider>
                    <RouterProvider>
                        <ToastProvider>{children}</ToastProvider>
                    </RouterProvider>
                </AuthProvider>
            </QueryProvider>
        </ThemeProvider>
    );
}
