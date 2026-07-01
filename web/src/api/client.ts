import i18n from '../i18n'

const baseURL = import.meta.env.VITE_API_URL ?? "";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

function translateError(errorCode: string, fallbackMessage: string): string {
  const key = `errors:${errorCode}`;
  const translated = i18n.t(key);
  return translated !== key ? translated : i18n.t('errors:unknown', fallbackMessage);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem("api-token");
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };
  if (token) {
    headers["api-token"] = token;
  }

  const res = await fetch(`${baseURL}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 401) {
    localStorage.removeItem("api-token");
    localStorage.removeItem("user");
    window.location.href = "/login";
    throw new ApiError(translateError("unauthorized", "Unauthorized"), 401);
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const errorCode = body.error || "";
    const fallback = body.message || `${res.status} ${res.statusText}`;
    throw new ApiError(translateError(errorCode, fallback), res.status);
  }

  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, {
      method: "POST",
      body: JSON.stringify(body),
    }),
};
