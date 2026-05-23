"use client";

import { useEffect, useState } from "react";
import { api } from "../utils/request";

const ROLES = ["reader", "writer", "admin"];
const PERMISSIONS = [
  { key: "api:random", label: "随机图片" },
  { key: "api:images", label: "图片列表" },
  { key: "api:tags", label: "标签" },
  { key: "api:config", label: "配置" },
  { key: "api:upload", label: "上传" },
  { key: "api:delete", label: "删除" },
  { key: "api:cleanup", label: "清理" },
  { key: "api:debug", label: "调试" },
];

type AKSKEntry = {
  access_key: string;
  name: string;
  description: string;
  role: string;
  custom_permissions: string[];
  created_at: number;
  enabled: boolean;
};

type CreateResult = {
  access_key: string;
  secret_key: string;
};

function maskAK(ak: string) {
  if (!ak || ak.length <= 8) return ak;
  return `${ak.slice(0, 6)}****${ak.slice(-4)}`;
}

export default function AKSKManager() {
  const [entries, setEntries] = useState<AKSKEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [newSecret, setNewSecret] = useState<CreateResult | null>(null);
  const [form, setForm] = useState({
    name: "",
    description: "",
    role: "reader",
    custom_permissions: [] as string[],
  });

  const loadEntries = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.get<{ entries: AKSKEntry[]; count: number }>(
        "/api/admin/aksk/list"
      );
      setEntries(data.entries || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载 AK/SK 列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadEntries();
  }, []);

  const togglePermission = (permission: string) => {
    setForm((prev) => ({
      ...prev,
      custom_permissions: prev.custom_permissions.includes(permission)
        ? prev.custom_permissions.filter((p) => p !== permission)
        : [...prev.custom_permissions, permission],
    }));
  };

  const createAKSK = async () => {
    if (!form.name.trim()) {
      setError("名称不能为空");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const result = await api.post<CreateResult>("/api/admin/aksk/create", form);
      setNewSecret(result);
      setShowCreate(false);
      setForm({ name: "", description: "", role: "reader", custom_permissions: [] });
      setMessage("AK/SK 创建成功，请立即保存 Secret Key");
      await loadEntries();
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建 AK/SK 失败");
    } finally {
      setLoading(false);
    }
  };

  const updateEntry = async (entry: AKSKEntry, patch: Partial<AKSKEntry>) => {
    setLoading(true);
    setError(null);
    try {
      await api.request("/api/admin/aksk/update", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          access_key: entry.access_key,
          ...patch,
        }),
      });
      setMessage("更新成功");
      await loadEntries();
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新失败");
    } finally {
      setLoading(false);
    }
  };

  const deleteEntry = async (entry: AKSKEntry) => {
    if (!confirm(`确认删除 ${entry.name || entry.access_key}？`)) return;
    setLoading(true);
    setError(null);
    try {
      await api.request("/api/admin/aksk/delete", {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ access_key: entry.access_key }),
      });
      setMessage("删除成功");
      await loadEntries();
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败");
    } finally {
      setLoading(false);
    }
  };

  const rotateSK = async (entry: AKSKEntry) => {
    if (!confirm(`确认轮换 ${entry.name || entry.access_key} 的 Secret Key？旧 SK 将立即失效。`)) return;
    setLoading(true);
    setError(null);
    try {
      const result = await api.post<CreateResult>("/api/admin/aksk/rotate", {
        access_key: entry.access_key,
      });
      setNewSecret(result);
      setMessage("SK 已轮换，请立即保存新 Secret Key");
    } catch (err) {
      setError(err instanceof Error ? err.message : "轮换失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-gray-900 dark:text-gray-100">AK/SK 管理</h2>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理外部脚本访问 /openapi/* 的访问密钥、角色和权限。
          </p>
        </div>
        <div className="flex gap-2">
          <a
            href="/openapi/docs"
            target="_blank"
            rel="noreferrer"
            className="rounded-lg border border-gray-300 dark:border-gray-600 px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-slate-700"
          >
            Swagger 文档
          </a>
          <button
            onClick={() => setShowCreate(true)}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
          >
            创建 AK/SK
          </button>
        </div>
      </div>

      {(message || error) && (
        <div className={`rounded-lg p-3 text-sm ${error ? "bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-300" : "bg-green-50 text-green-700 dark:bg-green-900/20 dark:text-green-300"}`}>
          {error || message}
        </div>
      )}

      <div className="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-slate-800">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-slate-900/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Access Key</th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">名称</th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">角色</th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">自定义权限</th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">状态</th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
              {loading && entries.length === 0 ? (
                <tr><td className="px-4 py-8 text-center text-gray-500" colSpan={6}>加载中...</td></tr>
              ) : entries.length === 0 ? (
                <tr><td className="px-4 py-8 text-center text-gray-500" colSpan={6}>暂无 AK/SK</td></tr>
              ) : (
                entries.map((entry) => (
                  <tr key={entry.access_key} className="text-sm text-gray-700 dark:text-gray-200">
                    <td className="px-4 py-3 font-mono">{maskAK(entry.access_key)}</td>
                    <td className="px-4 py-3">
                      <div className="font-medium">{entry.name}</div>
                      {entry.description && <div className="text-xs text-gray-500">{entry.description}</div>}
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={entry.role}
                        onChange={(e) => updateEntry(entry, { role: e.target.value })}
                        className="rounded border border-gray-300 bg-white px-2 py-1 dark:border-gray-600 dark:bg-slate-900"
                      >
                        {ROLES.map((role) => <option key={role} value={role}>{role}</option>)}
                      </select>
                    </td>
                    <td className="px-4 py-3 max-w-xs">
                      <div className="flex flex-wrap gap-1">
                        {(entry.custom_permissions || []).length === 0 ? (
                          <span className="text-xs text-gray-400">无</span>
                        ) : (
                          entry.custom_permissions.map((p) => (
                            <span key={p} className="rounded bg-indigo-50 px-2 py-1 text-xs text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300">{p}</span>
                          ))
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => updateEntry(entry, { enabled: !entry.enabled })}
                        className={`rounded-full px-3 py-1 text-xs ${entry.enabled ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300" : "bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300"}`}
                      >
                        {entry.enabled ? "启用" : "停用"}
                      </button>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        <button onClick={() => rotateSK(entry)} className="text-indigo-600 hover:underline dark:text-indigo-300">轮换</button>
                        <button onClick={() => deleteEntry(entry)} className="text-red-600 hover:underline dark:text-red-300">删除</button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-xl rounded-xl bg-white p-6 shadow-xl dark:bg-slate-800">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">创建 AK/SK</h3>
            <div className="mt-4 space-y-4">
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="名称，如 batch-uploader"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 dark:border-gray-600 dark:bg-slate-900"
              />
              <input
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="描述（可选）"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 dark:border-gray-600 dark:bg-slate-900"
              />
              <select
                value={form.role}
                onChange={(e) => setForm({ ...form, role: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 dark:border-gray-600 dark:bg-slate-900"
              >
                {ROLES.map((role) => <option key={role} value={role}>{role}</option>)}
              </select>
              <div>
                <div className="mb-2 text-sm font-medium text-gray-700 dark:text-gray-200">额外自定义权限</div>
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
                  {PERMISSIONS.map((permission) => (
                    <label key={permission.key} className="flex items-center gap-2 rounded border border-gray-200 p-2 text-sm dark:border-gray-700">
                      <input
                        type="checkbox"
                        checked={form.custom_permissions.includes(permission.key)}
                        onChange={() => togglePermission(permission.key)}
                      />
                      {permission.label}
                    </label>
                  ))}
                </div>
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <button onClick={() => setShowCreate(false)} className="rounded-lg border border-gray-300 px-4 py-2 dark:border-gray-600">取消</button>
              <button onClick={createAKSK} className="rounded-lg bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700">创建</button>
            </div>
          </div>
        </div>
      )}

      {newSecret && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-lg rounded-xl bg-white p-6 shadow-xl dark:bg-slate-800">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">请保存 Secret Key</h3>
            <p className="mt-2 text-sm text-red-600 dark:text-red-300">Secret Key 只显示一次，关闭后无法再次查看，只能轮换。</p>
            <div className="mt-4 space-y-3 rounded-lg bg-gray-50 p-4 font-mono text-sm dark:bg-slate-900">
              <div><span className="text-gray-500">AK: </span>{newSecret.access_key}</div>
              <div><span className="text-gray-500">SK: </span>{newSecret.secret_key}</div>
            </div>
            <div className="mt-6 flex justify-end gap-2">
              <button
                onClick={() => navigator.clipboard?.writeText(`AK=${newSecret.access_key}\nSK=${newSecret.secret_key}`)}
                className="rounded-lg border border-gray-300 px-4 py-2 dark:border-gray-600"
              >
                复制
              </button>
              <button onClick={() => setNewSecret(null)} className="rounded-lg bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700">我已保存</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
