import { Toaster } from "@kleff/ui";
import type { ReactNode } from "react";

export function ToastProvider({ children }: { children: ReactNode }) {
    return (
        <>
            {children}
            <Toaster />
        </>
    );
}
