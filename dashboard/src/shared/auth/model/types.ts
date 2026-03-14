export interface AuthorizationContextValue {
  shadowMode: boolean;
  enforceMode: boolean;
  isLoading: boolean;
  error: Error | null;
}
