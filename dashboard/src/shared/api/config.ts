export const apiConfig = {
  baseURL: (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, ""),
} as const;
