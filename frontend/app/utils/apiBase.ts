let runtimeApiUrl = "";
const configuredApiUrl = process.env.NEXT_PUBLIC_API_URL || "";

function isLoopbackHost(hostname: string): boolean {
  return hostname === "localhost" || hostname === "127.0.0.1" || hostname === "::1";
}

function normalizeForCurrentBrowser(apiUrl: string): string {
  if (typeof window === "undefined") {
    return apiUrl.replace(/\/$/, "");
  }

  if (!apiUrl) {
    return `${window.location.protocol}//${window.location.hostname}:8686`;
  }

  try {
    const parsed = new URL(apiUrl);
    if (isLoopbackHost(parsed.hostname) && !isLoopbackHost(window.location.hostname)) {
      parsed.hostname = window.location.hostname;
      return parsed.toString().replace(/\/$/, "");
    }
    return parsed.toString().replace(/\/$/, "");
  } catch {
    return apiUrl.replace(/\/$/, "");
  }
}

export function setRuntimeApiBaseUrl(apiUrl: string): void {
  runtimeApiUrl = apiUrl;
}

/**
 * Resolve the backend API base URL for the current browser.
 *
 * Docker/local deployments often build NEXT_PUBLIC_API_URL as http://localhost:8686.
 * That works only on the host machine. When another LAN/WAN client opens the UI,
 * its browser would otherwise call its own localhost:8686. In browsers, rewrite
 * loopback API hosts to the current page hostname while preserving protocol/port.
 */
export function getApiBaseUrl(): string {
  return normalizeForCurrentBrowser(runtimeApiUrl || configuredApiUrl);
}

export function resolveApiUrl(endpoint: string): URL {
  const baseUrl = getApiBaseUrl();
  return new URL(endpoint, baseUrl || window.location.origin);
}
