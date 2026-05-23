import { useState, useEffect } from "react";
import type { RuntimeConfig } from "../utils/runtimeConfig";
import { setRuntimeConfig } from "../utils/runtimeConfig";

export function useConfig() {
  const [config, setConfig] = useState<RuntimeConfig>({
    apiUrl: "",
    assetBaseUrl: "",
    openapiDocsUrl: "",
    remotePatterns: "",
    backendPort: "8686",
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/config", { cache: "no-store" })
      .then((res) => res.json())
      .then((data) => {
        const normalized = setRuntimeConfig(data);
        setConfig(normalized);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Failed to load config:", err);
        setLoading(false);
      });
  }, []);

  return { config, loading };
}
