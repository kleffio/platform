import { Toaster } from "@/shared/ui/sonner";
import type { ReactNode } from "react";

export function ToastProvider({ children }: { children: ReactNode }) {
    return (
        <>
            {children}
            <Toaster />
        </>
    );
}
