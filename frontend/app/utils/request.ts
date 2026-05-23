import { getApiKey } from "./auth";
import { loadRuntimeConfig, resolveApiUrl } from "./runtimeConfig";

interface RequestOptions extends RequestInit {
  params?: Record<string, string>;
}

interface ConfigResponse {
  apiUrl: string;
  assetBaseUrl: string;
  openapiDocsUrl: string;
  remotePatterns: string;
  backendPort: string;
}

let hasInitialized = false;

async function initializeRuntimeConfig() {
  try {
    await loadRuntimeConfig();
  } catch (error) {
    console.error("Failed to initialize runtime config:", error);
  }
}

export async function request<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  if (!hasInitialized) {
    await initializeRuntimeConfig();
    hasInitialized = true;
  }

  const apiKey = getApiKey();
  const { params, ...restOptions } = options;

  const url: URL = await resolveApiUrl(endpoint);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.append(key, value);
    }
  }

  const headers = {
    Authorization: `Bearer ${apiKey}`,
    ...options.headers,
  };

  const response = await fetch(url.toString(), {
    ...restOptions,
    headers,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || "请求失败");
  }

  return response.json();
}

// 获取静态文件目录列表
export async function fetchDirectoryListing(
  path = "/images/"
): Promise<string[]> {
  const response = await api.get<{ files: string[] }>("/directory", { path });
  return response.files;
}

// 封装常用请求方法
export const api = {
  request,
  get: <T>(endpoint: string, params?: Record<string, string>) =>
    request<T>(endpoint, { method: "GET", params }),

  post: <T>(endpoint: string, data?: any) =>
    request<T>(endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    }),

  delete: <T>(endpoint: string) => request<T>(endpoint, { method: "DELETE" }),

  upload: <T>(endpoint: string, files: File[]) => {
    const formData = new FormData();
    for (const file of files) {
      formData.append("images[]", file);
    }
    return request<T>(endpoint, {
      method: "POST",
      body: formData,
    });
  },
};
