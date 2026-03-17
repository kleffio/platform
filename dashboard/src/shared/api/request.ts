import type { AxiosRequestConfig } from "axios";
import { apiClient } from "./client";

export function get<TResponse>(url: string, config?: AxiosRequestConfig) {
  return apiClient.get<TResponse>(url, config).then((response) => response.data);
}

export function post<TResponse, TBody = unknown>(
  url: string,
  data?: TBody,
  config?: AxiosRequestConfig<TBody>
) {
  return apiClient.post<TResponse>(url, data, config).then((response) => response.data);
}

export function put<TResponse, TBody = unknown>(
  url: string,
  data?: TBody,
  config?: AxiosRequestConfig<TBody>
) {
  return apiClient.put<TResponse>(url, data, config).then((response) => response.data);
}

export function patch<TResponse, TBody = unknown>(
  url: string,
  data?: TBody,
  config?: AxiosRequestConfig<TBody>
) {
  return apiClient.patch<TResponse>(url, data, config).then((response) => response.data);
}

export function del<TResponse>(url: string, config?: AxiosRequestConfig) {
  return apiClient.delete<TResponse>(url, config).then((response) => response.data);
}
