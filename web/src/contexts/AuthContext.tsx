import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Navigate } from "react-router-dom";

interface User {
  username: string;
  name?: string;
  role?: string;
}

interface AuthContextValue {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  loading: boolean;
}

const AuthContext = createContext<AuthContextValue>({
  isAuthenticated: false,
  user: null,
  token: null,
  login: async () => {},
  logout: async () => {},
  loading: true,
});

export function useAuth() {
  return useContext(AuthContext);
}

const baseURL = import.meta.env.VITE_API_URL ?? "";

export function AuthProvider({ children }: { children: ReactNode }) {
  const { t } = useTranslation("login");
  const [token, setToken] = useState<string | null>(localStorage.getItem("api-token"));
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem("user");
    return raw ? (JSON.parse(raw) as User) : null;
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(false);
  }, []);

  const isAuthenticated = !!token;

  const login = async (username: string, password: string) => {
    const res = await fetch(`${baseURL}/api/admin/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.message || t("loginFailed"));
    }
    const data = await res.json();
    const tok = data.token || data.access_token || data.accessToken || "";
    const u: User = data.user || { username };
    setToken(tok);
    setUser(u);
    localStorage.setItem("api-token", tok);
    localStorage.setItem("user", JSON.stringify(u));
  };

  const logout = async () => {
    const t = localStorage.getItem("api-token");
    if (t) {
      await fetch(`${baseURL}/api/admin/logout`, {
        method: "POST",
        headers: { "api-token": t },
      }).catch(() => {});
    }
    setToken(null);
    setUser(null);
    localStorage.removeItem("api-token");
    localStorage.removeItem("user");
    window.location.href = "/login";
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, token, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { t } = useTranslation("common");
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <div className="flex h-screen items-center justify-center">{t("loading")}</div>;
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
