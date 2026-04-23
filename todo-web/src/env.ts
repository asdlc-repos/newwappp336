// Resolves the API base URL. Order of precedence:
// 1. window.__ENV__.VITE_API_URL  (runtime-injected via nginx entrypoint)
// 2. import.meta.env.VITE_API_URL (build-time Vite env var)
// 3. Empty string -> same-origin requests (useful for local `vite dev` with proxy)
declare global {
  interface Window {
    __ENV__?: { VITE_API_URL?: string };
  }
}

export function getApiUrl(): string {
  const runtime = typeof window !== 'undefined' ? window.__ENV__?.VITE_API_URL : '';
  if (runtime && runtime.trim() !== '' && !runtime.includes('${')) {
    return runtime.replace(/\/+$/, '');
  }
  const buildTime = import.meta.env?.VITE_API_URL as string | undefined;
  if (buildTime && buildTime.trim() !== '') {
    return buildTime.replace(/\/+$/, '');
  }
  return '';
}
