import type { ReactNode } from "react";
import { AuthProvider } from "@/features/auth";
import { QueryProvider } from "@/app/providers/query-provider";
import { ThemeProvider } from "@/app/providers/theme-provider";
import { ToastProvider } from "@/app/providers/toast-provider";

export function AppProvider({ children }: { children: ReactNode }) {
    return (
        <ThemeProvider defaultTheme="dark" storageKey="kleff-ui-theme">
            <QueryProvider>
                <AuthProvider>
                    <ToastProvider>{children}</ToastProvider>
                </AuthProvider>
            </QueryProvider>
        </ThemeProvider>
    );
}
