export const oidcConfig = {
  authority: "https://auth.kleff.io/realms/kleffio",
  client_id: "platform-dashboard",
  redirect_uri: window.location.origin + "/",
  post_logout_redirect_uri: window.location.origin + "/",
  onSigninCallback: () => {
    window.history.replaceState({}, document.title, window.location.pathname);
  },
};
