import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import "@/app/styles/index.css"
import App from "./App.tsx"
import { AppProvider } from "@/app/providers";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppProvider>
      <App />
    </AppProvider>
  </StrictMode>
)
