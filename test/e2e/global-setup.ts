import { spawn } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { randomBytes } from "node:crypto";
import type { FullConfig } from "@playwright/test";
import http from "node:http";
const ADMIN_USERNAME = "admin";
const ADMIN_PASSWORD = "admin123";
const PROJECT_ROOT = join(__dirname, "..", "..");

function getRandomPort(): number {
  return 10000 + Math.floor(Math.random() * 55000);
}

function httpGet(url: string): Promise<string> {
  return new Promise((resolve, reject) => {
    http
      .get(url, (res) => {
        let data = "";
        res.on("data", (chunk) => (data += chunk));
        res.on("end", () => resolve(data));
      })
      .on("error", reject);
  });
}

function httpPost(url: string, body: string, headers: Record<string, string> = {}): Promise<{ status: number; data: string }> {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    const req = http.request(
      {
        hostname: parsed.hostname,
        port: parsed.port,
        path: parsed.pathname,
        method: "POST",
        headers: { "Content-Type": "application/json", ...headers },
      },
      (res) => {
        let data = "";
        res.on("data", (chunk) => (data += chunk));
        res.on("end", () => resolve({ status: res.statusCode ?? 0, data }));
      }
    );
    req.on("error", reject);
    req.write(body);
    req.end();
  });
}

async function pollHealth(apiUrl: string, maxRetries: number = 180): Promise<void> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      await httpGet(`${apiUrl}/api/health`);
      console.log(`[global-setup] Server ready at ${apiUrl}`);
      return;
    } catch {
      if (i === 0) console.log("[global-setup] Waiting for server to start (Go compile + boot)...");
      await new Promise((r) => setTimeout(r, 500));
    }
  }
  throw new Error("Server failed to start within timeout");
}

async function loginAsAdmin(apiUrl: string): Promise<{ token: string; userId: number }> {
  const { status, data } = await httpPost(
    `${apiUrl}/api/admin/login`,
    JSON.stringify({ username: ADMIN_USERNAME, password: ADMIN_PASSWORD })
  );

  if (status !== 200) {
    throw new Error(`Admin login failed with status ${status}: ${data}`);
  }

  const response = JSON.parse(data);
  if (!response.token) {
    throw new Error(`Admin login response missing token: ${data}`);
  }

  return { token: response.token, userId: response.user?.id ?? 0 };
}

async function globalSetup(_config: FullConfig): Promise<void> {
  const port = getRandomPort();
  const apiUrl = `http://127.0.0.1:${port}`;
  const tempDir = join(tmpdir(), `rustdesk-e2e-${randomBytes(4).toString("hex")}`);
  mkdirSync(tempDir, { recursive: true });

  const dbPath = join(tempDir, "api.db");
  const configPath = join(tempDir, "config.yaml");
  const pidFile = join(tempDir, "server.pid");

  // Write minimal config
  const configYaml = [
    "server:",
    `  addr: "127.0.0.1:${port}"`,
    '  mode: "debug"',
    "database:",
    '  type: "sqlite"',
    `  path: "${dbPath}"`,
    "jwt:",
    `  key: "e2e-test-${randomBytes(4).toString("hex")}"`,
    "  expire_hours: 168",
    "app:",
    '  title: "RustDesk Admin"',
    "  register: true",
    "  captcha_threshold: 9999",
    "  web_client: true",
    "log:",
    '  level: "warn"',
    `  path: "${join(tempDir, "api.log")}"`,
  ].join("\n");

  writeFileSync(configPath, configYaml, "utf-8");

  console.log(`[global-setup] Config: ${configPath}`);
  console.log(`[global-setup] DB: ${dbPath}`);
  console.log(`[global-setup] Starting server on port ${port}...`);

  const proc = spawn("go", ["run", "./cmd/", "serve", "-c", configPath], {
    cwd: PROJECT_ROOT,
    env: {
      ...process.env,
      RUSTDESK_API_ADMIN_PASSWORD: ADMIN_PASSWORD,
      RUSTDESK_API_APP_CAPTCHA_THRESHOLD: "0",
    },
    stdio: ["ignore", "pipe", "pipe"],
  });

  proc.stdout?.on("data", (d: Buffer) => process.stdout.write(`[server] ${d}`));
  proc.stderr?.on("data", (d: Buffer) => process.stderr.write(`[server:err] ${d}`));
  proc.on("exit", (code) => {
    console.log(`[global-setup] Server process exited with code ${code}`);
  });

  // Write PID for teardown
  writeFileSync(pidFile, String(proc.pid ?? 0), "utf-8");

  // Poll health (Go compile + server boot time)
  await pollHealth(apiUrl);

  // Login as admin to get token
  const { token, userId } = await loginAsAdmin(apiUrl);
  console.log(`[global-setup] Logged in as admin (id=${userId}, token=${token.substring(0, 16)}...)`);

  // Export config for tests and teardown
  process.env.API_URL = apiUrl;
  process.env.ADMIN_TOKEN = token;
  process.env.ADMIN_USER_ID = String(userId);
  process.env.ADMIN_PASSWORD = ADMIN_PASSWORD;
  process.env.E2E_TEMP_DIR = tempDir;
  process.env.E2E_PID_FILE = pidFile;
  process.env.E2E_DB_PATH = dbPath;
}

export default globalSetup;
