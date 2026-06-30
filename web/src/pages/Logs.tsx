import { useEffect, useState } from "react";
import { Download } from "lucide-react";
import { useTranslation } from "react-i18next";
import { api } from "../api/client";

interface AuditConn {
  id: string;
  conn_id: string;
  peer_id: string;
  peer_hostname: string;
  peer_platform: string;
  peer_user: string;
  action: string;
  ip: string;
  created_at: string;
}

interface AuditFile {
  id: string;
  conn_id: string;
  peer_id: string;
  path: string;
  direction: string;
  size: number;
  created_at: string;
}

function formatBytes(b: number) {
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
  return `${(b / (1024 * 1024)).toFixed(1)} MB`;
}

export default function LogsPage() {
  const { t } = useTranslation("logs");
  const [tab, setTab] = useState<"sessions" | "files">("sessions");
  const [connLogs, setConnLogs] = useState<AuditConn[]>([]);
  const [fileLogs, setFileLogs] = useState<AuditFile[]>([]);
  const [loading, setLoading] = useState(true);
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");

  useEffect(() => {
    let mounted = true;
    async function load() {
      try {
        const [c, f] = await Promise.all([
          api.get<AuditConn[]>("/api/admin/audit_conn/list"),
          api.get<AuditFile[]>("/api/admin/audit_file/list"),
        ]);
        if (mounted) {
          setConnLogs(c);
          setFileLogs(f);
        }
      } catch {
        // ignore
      } finally {
        if (mounted) setLoading(false);
      }
    }
    load();
    return () => { mounted = false; };
  }, []);

  const exportCSV = () => {
    const data = tab === "sessions" ? connLogs : fileLogs;
    const headers = tab === "sessions"
      ? [t("connId"), t("peer"), t("platform"), t("user"), t("action"), t("ip"), t("created")]
      : [t("connId"), t("peer"), t("path"), t("direction"), t("size"), t("created")];
    const rows = data.map((d: AuditConn | AuditFile) =>
      tab === "sessions"
        ? [(d as AuditConn).conn_id, (d as AuditConn).peer_hostname, (d as AuditConn).peer_platform, (d as AuditConn).peer_user, (d as AuditConn).action, (d as AuditConn).ip, (d as AuditConn).created_at]
        : [(d as AuditFile).conn_id, (d as AuditFile).peer_id, (d as AuditFile).path, (d as AuditFile).direction, String((d as AuditFile).size), (d as AuditFile).created_at]
    );
    const csv = [headers, ...rows].map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(",")).join("\n");
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${tab}_logs.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="p-6 h-full overflow-auto">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-1 bg-white rounded-md p-1" style={{ border: "1px solid #E2E8F0" }}>
          <button
            onClick={() => setTab("sessions")}
            className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
              tab === "sessions" ? "bg-slate-100 text-slate-900 font-medium" : "text-slate-600 hover:bg-slate-50"
            }`}
          >
            {t("remoteSessions")}
          </button>
          <button
            onClick={() => setTab("files")}
            className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
              tab === "files" ? "bg-slate-100 text-slate-900 font-medium" : "text-slate-600 hover:bg-slate-50"
            }`}
          >
            {t("fileTransfers")}
          </button>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <input
              type="date"
              className="text-sm border rounded-md px-2 py-1.5"
              style={{ borderColor: "#E2E8F0" }}
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
            />
            <span className="text-sm text-slate-400">{t("to")}</span>
            <input
              type="date"
              className="text-sm border rounded-md px-2 py-1.5"
              style={{ borderColor: "#E2E8F0" }}
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
            />
          </div>
          <button
            onClick={exportCSV}
            className="flex items-center gap-2 text-sm px-3 py-1.5 rounded-md border bg-white text-slate-700 hover:bg-slate-50"
            style={{ borderColor: "#E2E8F0" }}
          >
            <Download size={14} />
            {t("exportCSV")}
          </button>
        </div>
      </div>

      <div className="bg-white rounded-lg overflow-hidden" style={{ border: "1px solid #E2E8F0" }}>
        {tab === "sessions" ? (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-slate-500" style={{ borderBottom: "1px solid #E2E8F0" }}>
                <th className="px-4 py-3 font-medium">{t("connId")}</th>
                <th className="px-4 py-3 font-medium">{t("peer")}</th>
                <th className="px-4 py-3 font-medium">{t("platform")}</th>
                <th className="px-4 py-3 font-medium">{t("user")}</th>
                <th className="px-4 py-3 font-medium">{t("action")}</th>
                <th className="px-4 py-3 font-medium">{t("ip")}</th>
                <th className="px-4 py-3 font-medium">{t("created")}</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={7} className="px-4 py-8 text-center text-slate-400">{t("loading")}</td></tr>
              ) : connLogs.length === 0 ? (
                <tr><td colSpan={7} className="px-4 py-8 text-center text-slate-400">{t("noSessionLogs")}</td></tr>
              ) : (
                connLogs.map((l) => (
                  <tr key={l.id} className="hover:bg-slate-50" style={{ borderBottom: "1px solid #E2E8F0" }}>
                    <td className="px-4 py-3 font-medium text-slate-700">{l.conn_id}</td>
                    <td className="px-4 py-3">
                      <div className="text-slate-800">{l.peer_hostname}</div>
                      <div className="text-xs text-slate-400">{l.peer_id}</div>
                    </td>
                    <td className="px-4 py-3 text-slate-700">{l.peer_platform}</td>
                    <td className="px-4 py-3 text-slate-700">{l.peer_user}</td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium" style={{ backgroundColor: "rgba(233,91,36,0.1)", color: "#E95B24" }}>
                        {l.action}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-700">{l.ip}</td>
                    <td className="px-4 py-3 text-slate-500">{l.created_at}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-slate-500" style={{ borderBottom: "1px solid #E2E8F0" }}>
                <th className="px-4 py-3 font-medium">{t("connId")}</th>
                <th className="px-4 py-3 font-medium">{t("peer")}</th>
                <th className="px-4 py-3 font-medium">{t("path")}</th>
                <th className="px-4 py-3 font-medium">{t("direction")}</th>
                <th className="px-4 py-3 font-medium">{t("size")}</th>
                <th className="px-4 py-3 font-medium">{t("created")}</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-slate-400">{t("loading")}</td></tr>
              ) : fileLogs.length === 0 ? (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-slate-400">{t("noFileTransferLogs")}</td></tr>
              ) : (
                fileLogs.map((l) => (
                  <tr key={l.id} className="hover:bg-slate-50" style={{ borderBottom: "1px solid #E2E8F0" }}>
                    <td className="px-4 py-3 font-medium text-slate-700">{l.conn_id}</td>
                    <td className="px-4 py-3 text-slate-700">{l.peer_id}</td>
                    <td className="px-4 py-3 text-slate-700">{l.path}</td>
                    <td className="px-4 py-3">
                      <span
                        className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                        style={{
                          backgroundColor: l.direction === "send" ? "rgba(22,163,74,0.1)" : "rgba(59,130,246,0.1)",
                          color: l.direction === "send" ? "#16a34a" : "#3b82f6",
                        }}
                      >
                        {l.direction}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-700">{formatBytes(l.size)}</td>
                    <td className="px-4 py-3 text-slate-500">{l.created_at}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
