import { NextRequest } from "next/server";

const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "::1"]);
const DEFAULT_BACKEND_PORT = process.env.IMAGEFLOW_BACKEND_PORT || "8686";

function stripPort(host: string): string {
  if (!host) return "";
  if (host.startsWith("[")) {
    const end = host.indexOf("]");
    return end >= 0 ? host.slice(1, end) : host;
  }
  return host.split(":")[0];
}

function normalizeUrl(url: string): string {
  return url.replace(/\/$/, "");
}

function configuredPublicBackendUrl(): string {
  return (
    process.env.IMAGEFLOW_PUBLIC_BACKEND_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    process.env.API_URL ||
    ""
  );
}

export interface RuntimeConfig {
  apiUrl: string;
  assetBaseUrl: string;
  openapiDocsUrl: string;
  remotePatterns: string;
  backendPort: string;
}

export function buildRuntimeConfig(request: NextRequest): RuntimeConfig {
  const requestHostHeader = request.headers.get("host") || "";
  const requestHostname = stripPort(requestHostHeader) || request.nextUrl.hostname;
  const configured = configuredPublicBackendUrl();

  let apiUrl = "";

  if (configured) {
    try {
      const backend = new URL(configured);
      if (LOOPBACK_HOSTS.has(backend.hostname) && !LOOPBACK_HOSTS.has(requestHostname)) {
        backend.hostname = requestHostname;
        if (!backend.port) backend.port = DEFAULT_BACKEND_PORT;
      }
      apiUrl = normalizeUrl(backend.toString());
    } catch {
      apiUrl = normalizeUrl(configured);
    }
  }

  if (!apiUrl) {
    const protocol = request.nextUrl.protocol || "http:";
    apiUrl = `${protocol}//${requestHostname}:${DEFAULT_BACKEND_PORT}`;
  }

  const remotePatterns = process.env.NEXT_PUBLIC_REMOTE_PATTERNS || new URL(apiUrl).host;

  return {
    apiUrl,
    assetBaseUrl: apiUrl,
    openapiDocsUrl: `${apiUrl}/openapi/docs`,
    remotePatterns,
    backendPort: DEFAULT_BACKEND_PORT,
  };
}
