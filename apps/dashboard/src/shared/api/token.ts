let accessToken: string | null = null;

export function setApiAccessToken(token: string | null) {
  accessToken = token;
}

export function clearApiAccessToken() {
  accessToken = null;
}

export function getApiAccessToken() {
  return accessToken;
}
