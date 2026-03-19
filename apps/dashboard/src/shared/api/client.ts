import axios from "axios";
import { apiConfig } from "./config";
import { normalizeApiError } from "./error";
import { getApiAccessToken } from "./token";

export const apiClient = axios.create({
  baseURL: apiConfig.baseURL || undefined,
  headers: {
    "Cache-Control": "no-cache, no-store, must-revalidate",
    Pragma: "no-cache",
    Expires: "0",
  },
});

apiClient.interceptors.request.use((config) => {
  const accessToken = getApiAccessToken();

  if (accessToken) {
    config.headers.set("Authorization", `Bearer ${accessToken}`);
  }

  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error: unknown) => Promise.reject(normalizeApiError(error))
);
