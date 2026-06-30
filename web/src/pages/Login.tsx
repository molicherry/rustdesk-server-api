import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { useTranslation } from "react-i18next";

export default function LoginPage() {
  const { t } = useTranslation("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/dashboard", { replace: true });
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login(username, password);
    } catch (err: unknown) {
      setError((err as { message?: string }).message || t("loginFailed"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen" style={{ backgroundColor: "#F7F9FB" }}>
      <div className="w-full max-w-sm rounded-lg shadow-sm overflow-hidden" style={{ border: "1px solid #E2E8F0", backgroundColor: "#fff" }}>
        <div className="px-6 py-5" style={{ backgroundColor: "#0F172A" }}>
          <h1 className="text-lg font-semibold text-white" style={{ fontFamily: "Inter, system-ui, sans-serif" }}>
            {t("title")}
          </h1>
        </div>
        <form onSubmit={handleSubmit} className="px-6 py-6 space-y-4">
          {error && (
            <div className="text-sm rounded-md px-3 py-2" style={{ backgroundColor: "#fef2f2", color: "#991b1b", border: "1px solid #fecaca" }}>
              {error}
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">{t("username")}</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full rounded-md px-3 py-2 text-sm border focus:outline-none focus:ring-2 focus:ring-orange-500"
              style={{ borderColor: "#E2E8F0" }}
              required
              autoFocus
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">{t("password")}</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-md px-3 py-2 text-sm border focus:outline-none focus:ring-2 focus:ring-orange-500"
              style={{ borderColor: "#E2E8F0" }}
              required
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-md px-4 py-2 text-sm font-medium text-white transition-colors disabled:opacity-60"
            style={{ backgroundColor: "#E95B24" }}
          >
            {loading ? t("signingIn") : t("signIn")}
          </button>
        </form>
      </div>
    </div>
  );
}
