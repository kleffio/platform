import axios from "axios";

export interface ApiError extends Error {
  status: number | null;
  code: string | null;
  data: unknown;
  isAuthError: boolean;
  isNetworkError: boolean;
  isCanceled: boolean;
}

export function isApiError(error: unknown): error is ApiError {
  return (
    error instanceof Error &&
    "status" in error &&
    "isAuthError" in error &&
    "isNetworkError" in error &&
    "isCanceled" in error
  );
}

export function normalizeApiError(error: unknown): ApiError {
  if (isApiError(error)) {
    return error;
  }

  if (axios.isAxiosError(error)) {
    const status = error.response?.status ?? null;
    const apiError = new Error(error.message || "API request failed") as ApiError;

    apiError.name = "ApiError";
    apiError.status = status;
    apiError.code = error.code ?? null;
    apiError.data = error.response?.data ?? null;
    apiError.isAuthError = status === 401 || status === 403;
    apiError.isNetworkError = !error.response;
    apiError.isCanceled = error.code === "ERR_CANCELED";

    return apiError;
  }

  const fallbackError = new Error("Unexpected API error") as ApiError;

  fallbackError.name = "ApiError";
  fallbackError.status = null;
  fallbackError.code = null;
  fallbackError.data = null;
  fallbackError.isAuthError = false;
  fallbackError.isNetworkError = false;
  fallbackError.isCanceled = false;

  return fallbackError;
}
