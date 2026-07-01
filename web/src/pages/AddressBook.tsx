import { useEffect, useState } from "react";
import { Pencil, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { api, type ListResponse } from "../api/client";

interface AddressBookEntry {
  id: string;
  name: string;
  device_id: string;
  alias: string;
  group: string;
  sync_status: string;
  last_sync: string;
}

function initials(name: string) {
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

export default function AddressBookPage() {
  const { t } = useTranslation("addressBook");
  const [entries, setEntries] = useState<AddressBookEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState(t("personal"));

  const categories = [t("personal"), t("shared")];
  const teamGroups = [t("engineering"), t("support"), t("sales"), t("devops")];

  useEffect(() => {
    let mounted = true;
    api.get<ListResponse<AddressBookEntry>>("/api/admin/address_book/list")
      .then((res) => { if (mounted) setEntries(res.data); })
      .catch(() => {})
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, []);

  return (
    <div className="flex h-full">
      <aside className="w-56 bg-white border-r border-slate-200 p-4">
        <div className="text-sm font-semibold text-slate-700 mb-3">{t("categories")}</div>
        <ul className="space-y-1 mb-4">
          {categories.map((c) => (
            <li key={c}>
              <button
                onClick={() => setSelectedCategory(c)}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                  selectedCategory === c
                    ? "bg-slate-100 text-slate-900 font-medium"
                    : "text-slate-600 hover:bg-slate-50"
                }`}
              >
                {c}
              </button>
            </li>
          ))}
        </ul>
        <div className="text-sm font-semibold text-slate-700 mb-3">{t("teamGroups")}</div>
        <ul className="space-y-1">
          {teamGroups.map((g) => (
            <li key={g}>
              <button className="w-full text-left px-3 py-2 rounded-md text-sm text-slate-600 hover:bg-slate-50 transition-colors">
                {g}
              </button>
            </li>
          ))}
        </ul>
      </aside>

      <div className="flex-1 p-6 overflow-auto">
        <div className="bg-white rounded-lg overflow-hidden" style={{ border: "1px solid #E2E8F0" }}>
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-slate-500" style={{ borderBottom: "1px solid #E2E8F0" }}>
                <th className="px-4 py-3 font-medium">{t("name")}</th>
                <th className="px-4 py-3 font-medium">{t("deviceId")}</th>
                <th className="px-4 py-3 font-medium">{t("alias")}</th>
                <th className="px-4 py-3 font-medium">{t("group")}</th>
                <th className="px-4 py-3 font-medium">{t("syncStatus")}</th>
                <th className="px-4 py-3 font-medium">{t("lastSync")}</th>
                <th className="px-4 py-3 font-medium w-20">{t("actions")}</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={7} className="px-4 py-8 text-center text-slate-400">{t("loading")}</td></tr>
              ) : entries.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-8 text-center text-slate-400">{t("noEntries")}</td></tr>
              ) : (
                entries.map((e) => (
                  <tr key={e.id} className="hover:bg-slate-50" style={{ borderBottom: "1px solid #E2E8F0" }}>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div
                          className="h-8 w-8 rounded-full flex items-center justify-center text-xs font-semibold text-white"
                          style={{ backgroundColor: "#E95B24" }}
                        >
                          {initials(e.name)}
                        </div>
                        <span className="font-medium text-slate-800">{e.name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-slate-700">{e.device_id}</td>
                    <td className="px-4 py-3 text-slate-700">{e.alias}</td>
                    <td className="px-4 py-3 text-slate-700">{e.group}</td>
                    <td className="px-4 py-3">
                      <span
                        className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                        style={{
                          backgroundColor: e.sync_status === "synced" ? "rgba(22,163,74,0.1)" : "rgba(245,158,11,0.1)",
                          color: e.sync_status === "synced" ? "#16a34a" : "#f59e0b",
                        }}
                      >
                        {e.sync_status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-500">{e.last_sync}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <button className="text-slate-400 hover:text-slate-700"><Pencil size={16} /></button>
                        <button className="text-slate-400 hover:text-red-600"><Trash2 size={16} /></button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
