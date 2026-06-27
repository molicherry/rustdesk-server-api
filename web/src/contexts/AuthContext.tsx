import { createContext, useContext, useEffect, useState, type ReactNode } from "react";

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

const baseURL = import.meta.env.VITE_API_URL || "http://localhost:21114";

export function AuthProvider({ children }: { children: ReactNode }) {
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
      throw new Error(body.message || "Login failed");
    }
    const data = await res.json();
    const t = data.token || data.access_token || data.accessToken || "";
    const u: User = data.user || { username };
    setToken(t);
    setUser(u);
    localStorage.setItem("api-token", t);
    localStorage.setItem("user", JSON.stringify(u));
  };

  const logout = async () => {
    const t = localStorage.getItem("api-token");
    if (t) {
      await fetch(`${baseURL}/api/admin/logout`, {
        method: "POST",
        headers: { Authorization: `Bearer ${t}` },
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

import { Navigate } from "react-router-dom";

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) return <div className="flex h-screen items-center justify-center">Loading...</div>;
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
