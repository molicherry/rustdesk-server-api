import { test, expect } from "@playwright/test";
import {
  getApiUrl,
  getAdminToken,
  getAdminUserId,
  apiRequest,
  setupAuthFix,
  createTestUserViaApi,
  deleteTestUserViaApi,
} from "./helpers";

test.describe("User Management", () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthFix(page);
    const token = getAdminToken();
    await page.goto(getApiUrl() + "/settings");
    await page.evaluate((t) => {
      localStorage.setItem("api-token", t);
      localStorage.setItem("user", JSON.stringify({ username: "admin", role: "admin" }));
    }, token);
    await page.reload({ waitUntil: "networkidle" });
    await page.waitForSelector("h2", { timeout: 10_000 }).catch(() => {});
  });

  test("create user appears in list via API", async () => {
    const token = getAdminToken();
    const username = `e2e-user-${Date.now()}`;

    // Create user
    const user = await createTestUserViaApi(username, "TestPass123!", false);

    // Verify in list
    const { status, data } = await apiRequest(
      "GET",
      `/api/admin/user/list?page=1&page_size=50&search=${username}`,
      undefined,
      token
    );
    expect(status).toBe(200);
    const list = data as { total: number; data: { username: string }[] };
    expect(list.total).toBeGreaterThanOrEqual(1);
    expect(list.data.some((u) => u.username === username)).toBe(true);

    // Cleanup
    await deleteTestUserViaApi(user.id);
  });

  test("edit user role via API", async () => {
    const token = getAdminToken();
    const username = `e2e-role-${Date.now()}`;

    // Create user first
    const user = await createTestUserViaApi(username, "TestPass123!", false, "user");

    // Update role to auditor
    const { status, data } = await apiRequest(
      "POST",
      "/api/admin/user/update",
      {
        id: user.id,
        role: "auditor",
      },
      token
    );
    expect(status).toBe(200);
    const updated = data as { role: string };
    expect(updated.role).toBe("auditor");

    // Cleanup
    await deleteTestUserViaApi(user.id);
  });

  test("delete user removes from list via API", async () => {
    const token = getAdminToken();
    const username = `e2e-del-${Date.now()}`;

    // Create
    const user = await createTestUserViaApi(username, "TestPass123!", false);
    expect(user.id).toBeGreaterThan(0);

    // Delete
    await deleteTestUserViaApi(user.id);

    // Verify removed
    const { status, data } = await apiRequest(
      "GET",
      `/api/admin/user/list?page=1&page_size=50&search=${username}`,
      undefined,
      token
    );
    expect(status).toBe(200);
    const list = data as { total: number; data: { username: string }[] };
    expect(list.data.some((u) => u.username === username)).toBe(false);
  });

  test("non-admin user cannot access user management API", async () => {
    const token = getAdminToken();
    const username = `e2e-nonadmin-${Date.now()}`;

    // Create a regular user (no admin privileges)
    const user = await createTestUserViaApi(username, "TestPass123!", false, "user");

    // Login as this user to get their token
    const loginRes = await apiRequest("POST", "/api/admin/login", {
      username,
      password: "TestPass123!",
    });
    expect(loginRes.status).toBe(200);
    const loginData = loginRes.data as { token: string };
    const userToken = loginData.token;
    expect(userToken).toBeTruthy();

    // Try to access user list with non-admin token → should fail
    const { status } = await apiRequest(
      "GET",
      "/api/admin/user/list?page=1&page_size=20",
      undefined,
      userToken
    );
    // User management requires "admin" role via RequireRole middleware
    expect(status).toBe(403);

    // Cleanup
    await deleteTestUserViaApi(user.id);
  });
});
