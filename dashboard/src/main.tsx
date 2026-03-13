import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import "@/app/styles/index.css"
import App from "./App.tsx"
import { ThemeProvider } from "@/app/providers/theme-provider.tsx"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider>
      <App />
    </ThemeProvider>
  </StrictMode>
)
