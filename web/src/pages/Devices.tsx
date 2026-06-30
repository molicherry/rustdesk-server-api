import { useEffect, useMemo, useState } from "react";
import { Search, MoreHorizontal, Filter } from "lucide-react";
import { useTranslation } from "react-i18next";
import { api } from "../api/client";

interface Peer {
  id: string;
  hostname: string;
  os: string;
  version: string;
  status: string;
  last_seen: string;
}

function osIcon(os: string) {
  if (/windows/i.test(os)) return "\u{1FA9F}";
  if (/mac|darwin/i.test(os)) return "\u{1F34E}";
  if (/linux/i.test(os)) return "\u{1F427}";
  return "\u{1F4BB}";
}

export default function DevicesPage() {
  const { t } = useTranslation("devices");
  const [peers, setPeers] = useState<Peer[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedGroup, setSelectedGroup] = useState(t("allDevices"));
  const [osFilter, setOsFilter] = useState(t("allOS"));
  const [statusFilter, setStatusFilter] = useState("allStatus");
  const [search, setSearch] = useState("");

  const groups = [t("allDevices"), t("hqOffice"), t("remoteUsers"), t("serverRoom")];

  useEffect(() => {
    let mounted = true;
    api.get<Peer[]>("/api/admin/peer/list")
      .then((data) => { if (mounted) setPeers(data); })
      .catch(() => {})
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, []);

  const filtered = useMemo(() => {
    return peers.filter((p) => {
      if (osFilter !== t("allOS") && !p.os.toLowerCase().includes(osFilter.toLowerCase())) return false;
      if (statusFilter !== "allStatus" && p.status !== statusFilter) return false;
      if (search) {
        const q = search.toLowerCase();
        return (
          p.hostname.toLowerCase().includes(q) ||
          p.id.toLowerCase().includes(q) ||
          p.os.toLowerCase().includes(q)
        );
      }
      return true;
    });
  }, [peers, osFilter, statusFilter, search, t]);

  return (
    <div className="flex h-full">
      {/* Left group tree */}
      <aside className="w-56 bg-white border-r border-slate-200 p-4">
        <div className="text-sm font-semibold text-slate-700 mb-3">{t("groups")}</div>
        <ul className="space-y-1">
          {groups.map((g) => (
            <li key={g}>
              <button
                onClick={() => setSelectedGroup(g)}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                  selectedGroup === g
                    ? "bg-slate-100 text-slate-900 font-medium"
                    : "text-slate-600 hover:bg-slate-50"
                }`}
              >
                {g}
              </button>
            </li>
          ))}
        </ul>
      </aside>

      <div className="flex-1 p-6 overflow-auto">
        {/* Top bar */}
        <div className="flex flex-wrap items-center gap-3 mb-4">
          <div className="flex items-center gap-2 bg-white border rounded-md px-3 py-2" style={{ borderColor: "#E2E8F0" }}>
            <Filter size={14} className="text-slate-400" />
            <select
              className="text-sm bg-transparent outline-none text-slate-700"
              value={osFilter}
              onChange={(e) => setOsFilter(e.target.value)}
            >
              <option>{t("allOS")}</option>
              <option>{t("windows")}</option>
              <option>{t("macos")}</option>
              <option>{t("linux")}</option>
            </select>
          </div>
          <div className="flex items-center gap-2 bg-white border rounded-md px-3 py-2" style={{ borderColor: "#E2E8F0" }}>
            <Filter size={14} className="text-slate-400" />
            <select
              className="text-sm bg-transparent outline-none text-slate-700"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <option value="allStatus">{t("allStatus")}</option>
              <option value="online">{t("online")}</option>
              <option value="offline">{t("offline")}</option>
            </select>
          </div>
          <div className="flex items-center gap-2 bg-white border rounded-md px-3 py-2 flex-1 min-w-[200px]" style={{ borderColor: "#E2E8F0" }}>
            <Search size={14} className="text-slate-400" />
            <input
              type="text"
              placeholder={t("searchDevices")}
              className="text-sm bg-transparent outline-none w-full text-slate-700"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <div className="flex gap-2">
            <button className="text-sm px-3 py-2 rounded-md border bg-white text-slate-700 hover:bg-slate-50" style={{ borderColor: "#E2E8F0" }}>
              {t("batchAction")}
            </button>
          </div>
        </div>

        {/* Table */}
        <div className="bg-white rounded-lg overflow-hidden" style={{ border: "1px solid #E2E8F0" }}>
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-slate-500" style={{ borderBottom: "1px solid #E2E8F0" }}>
                <th className="px-4 py-3 font-medium">{t("nameId")}</th>
                <th className="px-4 py-3 font-medium">{t("os")}</th>
                <th className="px-4 py-3 font-medium">{t("version")}</th>
                <th className="px-4 py-3 font-medium">{t("status")}</th>
                <th className="px-4 py-3 font-medium">{t("lastSeen")}</th>
                <th className="px-4 py-3 font-medium w-10"></th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-slate-400">{t("loading")}</td></tr>
              ) : filtered.length === 0 ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-slate-400">{t("noDevices")}</td></tr>
              ) : (
                filtered.map((p) => (
                  <tr key={p.id} className="hover:bg-slate-50" style={{ borderBottom: "1px solid #E2E8F0" }}>
                    <td className="px-4 py-3">
                      <div className="font-medium text-slate-800">{p.hostname}</div>
                      <div className="text-xs text-slate-400">{p.id}</div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="flex items-center gap-1">
                        <span>{osIcon(p.os)}</span>
                        <span className="text-slate-700">{p.os}</span>
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-700">{p.version}</td>
                    <td className="px-4 py-3">
                      <span
                        className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                        style={{
                          backgroundColor: p.status === "online" ? "rgba(22,163,74,0.1)" : "rgba(100,116,139,0.1)",
                          color: p.status === "online" ? "#16a34a" : "#64748b",
                        }}
                      >
                        {p.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-500">{p.last_seen}</td>
                    <td className="px-4 py-3">
                      <button className="text-slate-400 hover:text-slate-700">
                        <MoreHorizontal size={16} />
                      </button>
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
