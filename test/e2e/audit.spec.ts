import { test, expect } from "@playwright/test";
import {
  getApiUrl,
  getAdminToken,
  apiRequest,
  setupAuthFix,
  auditConnEventViaApi,
  auditFileEventViaApi,
  createTestDeviceViaApi,
} from "./helpers";

test.describe("Audit Logs", () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthFix(page);
    const token = getAdminToken();
    await page.goto(getApiUrl() + "/logs");
    await page.evaluate((t) => {
      localStorage.setItem("api-token", t);
      localStorage.setItem("user", JSON.stringify({ username: "admin", role: "admin" }));
    }, token);
    await page.reload({ waitUntil: "networkidle" });
    // Wait for logs page to render (tabs are always present)
    await page.waitForSelector('button', { timeout: 10_000 }).catch(() => {});
  });

  test("connection logs list loads", async ({ page }) => {
    const token = getAdminToken();
    const peerId = `audit-conn-${Date.now()}`;

    // Seed device and audit events
    await createTestDeviceViaApi(peerId, "audit-conn-host");
    await auditConnEventViaApi(peerId, 1001, 1);

    // Navigate to logs page
    await page.goto(`${getApiUrl()}/logs`, { waitUntil: "networkidle" });

    // Click on Connection logs tab if present
    const connTab = page.getByText(/connection/i).first();
    if (await connTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await connTab.click();
      await page.waitForTimeout(1000);
    }

    // Verify connection log data loaded via API
    const { status, data } = await apiRequest(
      "GET",
      `/api/admin/audit_conn/list?page=1&page_size=20&peer_id=${peerId}`,
      undefined,
      token
    );
    expect(status).toBe(200);
    const list = data as { total: number };
    expect(list.total).toBeGreaterThanOrEqual(1);
  });

  test("file transfer logs list loads", async ({ page }) => {
    const token = getAdminToken();
    const peerId = `audit-file-${Date.now()}`;

    await createTestDeviceViaApi(peerId, "audit-file-host");
    await auditFileEventViaApi(peerId, "/home/test/file.txt", 1);

    await page.goto(`${getApiUrl()}/logs`, { waitUntil: "networkidle" });

    // Click on File Transfer logs tab if present
    const fileTab = page.getByText(/file/i).first();
    if (await fileTab.isVisible({ timeout: 3000 }).catch(() => false)) {
      await fileTab.click();
      await page.waitForTimeout(1000);
    }

    // Verify via API
    const { status, data } = await apiRequest(
      "GET",
      `/api/admin/audit_file/list?page=1&page_size=20&peer_id=${peerId}`,
      undefined,
      token
    );
    expect(status).toBe(200);
    const list = data as { total: number };
    expect(list.total).toBeGreaterThanOrEqual(1);
  });

  test("date filter on audit logs via API", async ({ page }) => {
    const token = getAdminToken();
    const peerId = `audit-date-${Date.now()}`;

    await createTestDeviceViaApi(peerId, "audit-date-host");
    await auditConnEventViaApi(peerId, 2001, 1);

    // Query with a wide date range that should include the event
    const now = Math.floor(Date.now() / 1000);
    const startTime = now - 86400; // 24 hours ago
    const endTime = now + 3600; // 1 hour from now

    const { status, data } = await apiRequest(
      "GET",
      `/api/admin/audit_conn/list?page=1&page_size=20&peer_id=${peerId}&start_time=${startTime}&end_time=${endTime}`,
      undefined,
      token
    );
    expect(status).toBe(200);
    const list = data as { total: number };
    expect(list.total).toBeGreaterThanOrEqual(1);

    // Query with past date range (should be empty)
    const { status: status2, data: data2 } = await apiRequest(
      "GET",
      `/api/admin/audit_conn/list?page=1&page_size=20&peer_id=${peerId}&start_time=1&end_time=2`,
      undefined,
      token
    );
    expect(status2).toBe(200);
    const emptyList = data2 as { total: number };
    expect(emptyList.total).toBe(0);
  });

  test("simulate audit events via API appear in logs", async ({ page }) => {
    const token = getAdminToken();
    const peerId = `audit-sim-${Date.now()}`;

    // Simulate multiple events via API
    await createTestDeviceViaApi(peerId, "audit-sim-host");
    await auditConnEventViaApi(peerId, 3001, 1);
    await auditConnEventViaApi(peerId, 3002, 2);
    await auditFileEventViaApi(peerId, "/tmp/sim-file1.txt", 1);
    await auditFileEventViaApi(peerId, "/tmp/sim-file2.bin", 2);

    // Verify both conn and file logs have data
    const connRes = await apiRequest("GET", `/api/admin/audit_conn/list?page=1&page_size=100&peer_id=${peerId}`, undefined, token);
    expect(connRes.status).toBe(200);
    const connList = connRes.data as { total: number; data: unknown[] };
    expect(connList.total).toBeGreaterThanOrEqual(2);

    const fileRes = await apiRequest("GET", `/api/admin/audit_file/list?page=1&page_size=100&peer_id=${peerId}`, undefined, token);
    expect(fileRes.status).toBe(200);
    const fileList = fileRes.data as { total: number; data: unknown[] };
    expect(fileList.total).toBeGreaterThanOrEqual(2);
  });
});
