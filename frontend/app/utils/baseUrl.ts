import { getApiBaseUrl } from "./apiBase";

/**
 * 为URL添加后端基础地址
 * @param url 相对路径或完整URL
 * @returns 完整URL
 */
export function getFullUrl(url: string): string {
  if (!url) return "";
  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }

  try {
    const baseUrl = getApiBaseUrl();
    if (typeof window !== "undefined") {
      return new URL(url, baseUrl || window.location.origin).toString();
    }
    return `${baseUrl}${url.startsWith("/") ? url : `/${url}`}`;
  } catch (error) {
    console.error("URL格式错误:", error);
    return url;
  }
}
