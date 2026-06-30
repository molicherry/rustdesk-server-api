import { useEffect, useState } from "react";
import { Monitor, Users, Zap, ShieldAlert } from "lucide-react";
import { useTranslation } from "react-i18next";
import { api } from "../api/client";

interface StatCardProps {
  icon: React.ElementType;
  label: string;
  value: string | number;
  trend: string;
  trendUp?: boolean;
}

function StatCard({ icon: Icon, label, value, trend, trendUp }: StatCardProps) {
  return (
    <div className="rounded-lg p-5 bg-white" style={{ border: "1px solid #E2E8F0" }}>
      <div className="flex items-center justify-between mb-3">
        <div className="p-2 rounded-md" style={{ backgroundColor: "rgba(233,91,36,0.08)" }}>
          <Icon size={20} style={{ color: "#E95B24" }} />
        </div>
        <span className={`text-xs font-medium ${trendUp ? "text-green-600" : "text-slate-500"}`}>{trend}</span>
      </div>
      <div className="text-2xl font-semibold text-slate-800" style={{ fontFamily: "Inter, system-ui, sans-serif" }}>
        {value}
      </div>
      <div className="text-sm text-slate-500 mt-0.5">{label}</div>
    </div>
  );
}

export default function DashboardPage() {
  const { t } = useTranslation("dashboard");
  const [onlineDevices, setOnlineDevices] = useState(0);
  const [connections, setConnections] = useState(0);
  const [users, setUsers] = useState(0);
  const [alerts, setAlerts] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    async function load() {
      try {
        const peers: unknown[] = await api.get("/api/admin/peer/list");
        if (!mounted) return;
        setOnlineDevices((peers as unknown[]).length);
        setConnections(0);
        setUsers(0);
        setAlerts(0);
      } catch {
        // ignore
      } finally {
        if (mounted) setLoading(false);
      }
    }
    load();
    return () => { mounted = false; };
  }, []);

  if (loading) {
    return (
      <div className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="rounded-lg p-5 bg-white h-28 animate-pulse" style={{ border: "1px solid #E2E8F0" }} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h2 className="text-lg font-semibold text-slate-800 mb-4" style={{ fontFamily: "Inter, system-ui, sans-serif" }}>
        {t("title")}
      </h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard icon={Monitor} label={t("onlineDevices")} value={onlineDevices} trend="-" trendUp={false} />
        <StatCard icon={Zap} label={t("activeConnections")} value={connections} trend="-" trendUp={false} />
        <StatCard icon={Users} label={t("totalUsers")} value={users} trend="-" trendUp={false} />
        <StatCard icon={ShieldAlert} label={t("securityAlerts")} value={alerts} trend="-" trendUp={false} />
      </div>
    </div>
  );
}
