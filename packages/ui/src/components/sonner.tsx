import * as React from "react";
import { Toaster as Sonner, type ToasterProps } from "sonner"
import { CircleCheckIcon, InfoIcon, TriangleAlertIcon, OctagonXIcon, Loader2Icon } from "lucide-react"

type Theme = "light" | "dark" | "system";

function getDocumentTheme(): Theme {
  if (typeof document === "undefined") return "system";
  const root = document.documentElement;
  if (root.classList.contains("dark")) return "dark";
  if (root.classList.contains("light")) return "light";
  return "system";
}

const Toaster = ({ ...props }: ToasterProps) => {
  const theme = getDocumentTheme();

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      icons={{
        success: <CircleCheckIcon className="size-4 text-amber-400" />,
        info: <InfoIcon className="size-4 text-amber-400/70" />,
        warning: <TriangleAlertIcon className="size-4 text-amber-400" />,
        error: <OctagonXIcon className="size-4 text-red-400" />,
        loading: <Loader2Icon className="size-4 animate-spin text-amber-400/60" />,
      }}
      style={
        {
          "--normal-bg": "oklch(0.16 0.008 55)",
          "--normal-text": "oklch(0.97 0 0)",
          "--normal-border": "oklch(0.8 0.17 90 / 0.18)",
          "--success-bg": "oklch(0.16 0.008 55)",
          "--success-text": "oklch(0.97 0 0)",
          "--success-border": "oklch(0.8 0.17 90 / 0.35)",
          "--error-bg": "oklch(0.16 0.008 55)",
          "--error-text": "oklch(0.97 0 0)",
          "--error-border": "oklch(0.63 0.24 27 / 0.4)",
          "--warning-bg": "oklch(0.16 0.008 55)",
          "--warning-text": "oklch(0.97 0 0)",
          "--warning-border": "oklch(0.8 0.17 90 / 0.35)",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
      {...props}
    />
  )
}

export { Toaster }
