const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "::1"]);
const DEFAULT_BACKEND_PORT = "8686";

export interface RuntimeConfig {
  apiUrl: string;
  assetBaseUrl: string;
  openapiDocsUrl: string;
  remotePatterns: string;
  backendPort: string;
}

let runtimeConfig: RuntimeConfig | null = null;
let runtimeConfigPromise: Promise<RuntimeConfig> | null = null;

function isLoopbackHost(hostname: string): boolean {
  return LOOPBACK_HOSTS.has(hostname);
}

function normalizeUrl(url: string): string {
  return url.replace(/\/$/, "");
}

function normalizeForCurrentBrowser(apiUrl: string): string {
  if (typeof window === "undefined") {
    return normalizeUrl(apiUrl);
  }

  if (!apiUrl) {
    return `${window.location.protocol}//${window.location.hostname}:${DEFAULT_BACKEND_PORT}`;
  }

  try {
    const parsed = new URL(apiUrl);
    if (isLoopbackHost(parsed.hostname) && !isLoopbackHost(window.location.hostname)) {
      parsed.hostname = window.location.hostname;
      if (!parsed.port) parsed.port = DEFAULT_BACKEND_PORT;
      return normalizeUrl(parsed.toString());
    }
    return normalizeUrl(parsed.toString());
  } catch {
    return normalizeUrl(apiUrl);
  }
}

function fallbackRuntimeConfig(): RuntimeConfig {
  const apiUrl = normalizeForCurrentBrowser("");
  return {
    apiUrl,
    assetBaseUrl: apiUrl,
    openapiDocsUrl: `${apiUrl}/openapi/docs`,
    remotePatterns: "",
    backendPort: DEFAULT_BACKEND_PORT,
  };
}

function normalizeRuntimeConfig(config: Partial<RuntimeConfig>): RuntimeConfig {
  const apiUrl = normalizeForCurrentBrowser(config.apiUrl || "");
  const assetBaseUrl = normalizeForCurrentBrowser(config.assetBaseUrl || apiUrl);
  return {
    apiUrl,
    assetBaseUrl,
    openapiDocsUrl: config.openapiDocsUrl || `${apiUrl}/openapi/docs`,
    remotePatterns: config.remotePatterns || "",
    backendPort: config.backendPort || DEFAULT_BACKEND_PORT,
  };
}

export function setRuntimeConfig(config: Partial<RuntimeConfig>): RuntimeConfig {
  runtimeConfig = normalizeRuntimeConfig(config);
  return runtimeConfig;
}

export async function loadRuntimeConfig(): Promise<RuntimeConfig> {
  if (runtimeConfig) return runtimeConfig;
  if (runtimeConfigPromise) return runtimeConfigPromise;

  runtimeConfigPromise = fetch("/api/config", { cache: "no-store" })
    .then(async (response) => {
      if (!response.ok) throw new Error(`Failed to load runtime config: ${response.status}`);
      return response.json();
    })
    .then((config) => setRuntimeConfig(config))
    .catch((error) => {
      console.error("Failed to fetch runtime config, using browser fallback:", error);
      runtimeConfig = fallbackRuntimeConfig();
      return runtimeConfig;
    })
    .finally(() => {
      runtimeConfigPromise = null;
    });

  return runtimeConfigPromise;
}

export function getRuntimeConfigSync(): RuntimeConfig {
  if (runtimeConfig) return runtimeConfig;
  runtimeConfig = fallbackRuntimeConfig();
  return runtimeConfig;
}

export async function resolveApiUrl(endpoint: string): Promise<URL> {
  const config = await loadRuntimeConfig();
  return new URL(endpoint, config.apiUrl || window.location.origin);
}

export function resolveApiUrlSync(endpoint: string): URL {
  const config = getRuntimeConfigSync();
  return new URL(endpoint, config.apiUrl || window.location.origin);
}

export function getFullUrl(url: string): string {
  if (!url) return "";
  if (url.startsWith("http://") || url.startsWith("https://")) return url;

  try {
    const config = getRuntimeConfigSync();
    return new URL(url, config.assetBaseUrl || config.apiUrl || window.location.origin).toString();
  } catch (error) {
    console.error("URL格式错误:", error);
    return url;
  }
}
