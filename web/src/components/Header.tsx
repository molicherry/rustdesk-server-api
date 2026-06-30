import { Bell, Globe, LogOut, Shield } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useAuth } from "../contexts/AuthContext";
import { useWebSocket } from "../hooks/useWebSocket";

function Breadcrumb({ dashboardLabel }: { dashboardLabel: string }) {
  const parts = window.location.pathname.split("/").filter(Boolean);
  return (
    <nav className="text-sm text-slate-500">
      <span className="font-medium text-slate-700 capitalize">{parts[0] || dashboardLabel}</span>
    </nav>
  );
}

export default function Header() {
  const { t, i18n } = useTranslation("common");
  const { user, logout } = useAuth();
  const { status } = useWebSocket();

  const wsColor = status === "connected" ? "#16a34a" : status === "connecting" ? "#f59e0b" : "#64748b";

  const handleLanguageChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    void i18n.changeLanguage(e.target.value);
    document.documentElement.lang = e.target.value;
  };

  return (
    <header
      className="flex items-center justify-between px-6 bg-white"
      style={{ height: 64, borderBottom: "1px solid #E2E8F0" }}
    >
      <div className="flex items-center gap-4">
        <Breadcrumb dashboardLabel={t("dashboard")} />
        <span className="text-sm text-slate-400">|</span>
        <span className="text-sm font-medium text-slate-700">{t("defaultTenant")}</span>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1.5 text-xs text-slate-500">
          <Shield size={14} />
          <span>{t("twoFA")}</span>
          <span className="ml-1 h-2 w-2 rounded-full inline-block" style={{ backgroundColor: "#16a34a" }} />
        </div>

        <div className="flex items-center gap-1.5 text-xs text-slate-500">
          <span>{t("ws")}</span>
          <span className="h-2 w-2 rounded-full inline-block" style={{ backgroundColor: wsColor }} />
        </div>

        <div className="flex items-center gap-1.5">
          <Globe size={16} className="text-slate-400" />
          <select
            value={i18n.language}
            onChange={handleLanguageChange}
            className="text-sm bg-transparent outline-none text-slate-600 hover:text-slate-900 cursor-pointer"
            aria-label={t("languageSelector")}
          >
            <option value="en">{t("english")}</option>
            <option value="zh-CN">{t("chinese")}</option>
          </select>
        </div>

        <button className="relative text-slate-600 hover:text-slate-900">
          <Bell size={18} />
          <span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-red-500" />
        </button>

        <div className="flex items-center gap-2">
          <span className="text-sm text-slate-700">{user?.username || t("admin")}</span>
          <button
            onClick={() => logout()}
            className="text-slate-500 hover:text-red-600"
            title={t("logout")}
          >
            <LogOut size={18} />
          </button>
        </div>
      </div>
    </header>
  );
}
