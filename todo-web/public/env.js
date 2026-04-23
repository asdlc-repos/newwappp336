// Default runtime config used when the app is served outside of the production
// container. The nginx entrypoint script replaces this file at container start
// with values derived from real environment variables.
window.__ENV__ = {
  VITE_API_URL: ""
};
