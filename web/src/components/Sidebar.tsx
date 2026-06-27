import { useState } from "react";
import {
  LayoutDashboard,
  Monitor,
  BookOpen,
  History,
  Settings,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { NavLink, useLocation } from "react-router-dom";

const navItems = [
  { to: "/dashboard", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/devices", icon: Monitor, label: "Devices" },
  { to: "/address-book", icon: BookOpen, label: "Address Book" },
  { to: "/logs", icon: History, label: "Logs" },
  { to: "/settings", icon: Settings, label: "Settings" },
];

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();

  return (
    <aside
      className="flex flex-col transition-all duration-300"
      style={{
        width: collapsed ? 64 : 260,
        backgroundColor: "#0F172A",
        fontFamily: "Inter, system-ui, sans-serif",
      }}
    >
      <div
        className="flex items-center px-4"
        style={{ height: 64, borderBottom: "1px solid rgba(255,255,255,0.08)" }}
      >
        {!collapsed && (
          <span className="text-lg font-semibold text-white tracking-tight">RustDesk</span>
        )}
        <button
          onClick={() => setCollapsed((c) => !c)}
          className="ml-auto text-slate-400 hover:text-white"
          title={collapsed ? "Expand" : "Collapse"}
        >
          {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
        </button>
      </div>

      <nav className="flex-1 py-4 space-y-1">
        {navItems.map((item) => {
          const isActive = location.pathname.startsWith(item.to);
          return (
            <NavLink
              key={item.to}
              to={item.to}
              className={
                "flex items-center gap-3 px-4 py-2.5 text-sm font-medium transition-colors " +
                (isActive
                  ? " text-white"
                  : " text-slate-400 hover:bg-slate-800 hover:text-white")
              }
              style={isActive ? { borderLeft: "2px solid #E95B24", backgroundColor: "rgba(233,91,36,0.08)" } : { borderLeft: "2px solid transparent" }}
              title={collapsed ? item.label : undefined}
            >
              <item.icon size={18} />
              {!collapsed && <span>{item.label}</span>}
            </NavLink>
          );
        })}
      </nav>
    </aside>
  );
}
