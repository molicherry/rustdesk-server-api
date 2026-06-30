import { request, Page } from "@playwright/test";

// ─── Configuration ───────────────────────────────────────────────────────────

export function getApiUrl(): string {
  if (!process.env.API_URL) throw new Error("API_URL not set (global-setup must run first)");
  return process.env.API_URL;
}

export function getAdminToken(): string {
  if (!process.env.ADMIN_TOKEN) throw new Error("ADMIN_TOKEN not set (global-setup must run first)");
  return process.env.ADMIN_TOKEN;
}

export function getAdminUserId(): number {
  if (!process.env.ADMIN_USER_ID) throw new Error("ADMIN_USER_ID not set");
  return parseInt(process.env.ADMIN_USER_ID, 10);
}

export function getAdminPassword(): string {
  return process.env.ADMIN_PASSWORD ?? "admin123";
}

// ─── API Helpers ─────────────────────────────────────────────────────────────

export async function apiRequest(
  method: string,
  path: string,
  body?: Record<string, unknown>,
  token?: string
): Promise<{ status: number; data: unknown }> {
  const apiUrl = getApiUrl();
  const ctx = await request.newContext({ baseURL: apiUrl });
  try {
    const headers: Record<string, string> = {};
    if (token) {
      headers["api-token"] = token;
    }
    const options: Record<string, unknown> = { headers };
    if (body) options.data = body;

    const response = await ctx.fetch(path, { ...options, method } as any);
    const text = await response.text();
    let data: unknown = text;
    try {
      data = JSON.parse(text);
    } catch {
      /* keep as text */
    }
    return { status: response.status(), data };
  } finally {
    await ctx.dispose();
  }
}

// ─── Auth Bypass (Browser fetch sends Bearer, backend expects api-token) ─────

export async function setupAuthFix(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const origFetch = window.fetch.bind(window);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).fetch = function (
      input: RequestInfo | URL,
      init?: RequestInit
    ): Promise<Response> {
      if (typeof input === "string" && input.includes("/api/")) {
        const opts = { ...(init || {}) };
        if (opts.headers && typeof opts.headers === "object" && !(opts.headers instanceof Headers)) {
          const h = opts.headers as Record<string, string>;
          const auth = h["Authorization"];
          if (auth && auth.startsWith("Bearer ")) {
            h["api-token"] = auth.slice(7);
            delete h["Authorization"];
          }
        }
        return origFetch(input, opts);
      }
      return origFetch(input, init);
    };
  });
}

// ─── Authenticated Page Setup ─────────────────────────────────────────────────

export async function setupAuthenticatedPage(page: Page): Promise<void> {
  await setupAuthFix(page);
  await page.addInitScript((token: string) => {
    localStorage.setItem("api-token", token);
    localStorage.setItem("user", JSON.stringify({ username: "admin", role: "admin" }));
  }, getAdminToken());
}

export async function goToPage(page: Page, path: string): Promise<void> {
  const apiUrl = getApiUrl();
  await setupAuthFix(page);
  // Set token before navigating so ProtectedRoute lets us through
  await page.goto(`${apiUrl}${path}`, { waitUntil: "domcontentloaded" });
}

// ─── Page Helpers ────────────────────────────────────────────────────────────

export async function logoutViaUI(page: Page): Promise<void> {
  const logoutBtn = page.locator('button[title="Logout"]');
  if (await logoutBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
    await logoutBtn.click();
  }
  // Clear storage for clean state
  await page.evaluate(() => {
    localStorage.clear();
    sessionStorage.clear();
  });
  await page.context().clearCookies();
}

// ─── Test Data Helpers ───────────────────────────────────────────────────────

export async function createTestUserViaApi(
  username: string,
  password: string,
  isAdmin: boolean = false,
  role?: string
): Promise<{ id: number; username: string }> {
  const adminToken = getAdminToken();
  const body: Record<string, unknown> = { username, password, is_admin: isAdmin };
  if (role) body.role = role;

  const { status, data } = await apiRequest("POST", "/api/admin/user/create", body, adminToken);
  if (status !== 200) {
    throw new Error(`Failed to create test user: ${JSON.stringify(data)}`);
  }
  const user = data as { id: number; username: string };
  return { id: user.id, username: user.username };
}

export async function deleteTestUserViaApi(userId: number): Promise<void> {
  const adminToken = getAdminToken();
  await apiRequest("POST", "/api/admin/user/delete", { id: userId }, adminToken);
}

export async function createTestDeviceViaApi(
  peerId: string,
  hostname: string,
  os?: string
): Promise<void> {
  const uuid = `uuid-${peerId}`;
  // Simulate heartbeat — uses "id" and "uuid" per RustDesk client protocol
  await apiRequest("POST", "/api/heartbeat", {
    id: peerId,
    uuid,
    ver: 1,
    conns: [],
  });
  // Submit sysinfo — uses "id" and "uuid" to match the heartbeat peer
  await apiRequest("POST", "/api/sysinfo", {
    id: peerId,
    uuid,
    hostname,
    os: os ?? "Windows 11",
    username: "testuser",
    version: "1.2.3",
    cpu: "Intel Core i7",
    memory: "16GB",
  });
}

export async function auditConnEventViaApi(
  peerId: string,
  connId: number,
  connType: number = 0
): Promise<void> {
  await apiRequest("POST", "/api/audit/conn", {
    peer_id: peerId,
    conn_id: connId,
    session_id: `session-${connId}`,
    action: "new",
    type: connType,
    uuid: `uuid-audit-${connId}`,
  });
}

export async function auditFileEventViaApi(
  peerId: string,
  path: string,
  fileType: number = 0
): Promise<void> {
  await apiRequest("POST", "/api/audit/file", {
    peer_id: peerId,
    path,
    info: "{}",
    is_file: true,
    type: fileType,
    uuid: `uuid-file-${Math.random().toString(36).slice(2, 10)}`,
  });
}
