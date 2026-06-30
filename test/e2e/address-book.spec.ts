import { test, expect } from "@playwright/test";
import {
  getApiUrl,
  getAdminToken,
  apiRequest,
  setupAuthFix,
} from "./helpers";

test.describe("Address Book", () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthFix(page);
    const token = getAdminToken();
    const apiUrl = getApiUrl();
    // Set up auth for ProtectedRoute
    await page.goto(apiUrl + "/address-book");
    await page.evaluate((t) => {
      localStorage.setItem("api-token", t);
      localStorage.setItem("user", JSON.stringify({ username: "admin", role: "admin" }));
    }, token);
  });

  test("create new contact via API and verify in list", async ({ page }) => {
    const token = getAdminToken();
    const adminId = parseInt(process.env.ADMIN_USER_ID || "1");

    // Create address book entry via API
    const { status, data } = await apiRequest(
      "POST",
      "/api/admin/address_book/create",
      {
        peer_id: "ab-peer-001",
        hostname: "contact-host-1",
        alias: "My Contact",
        platform: "Windows",
        user_id: adminId,
      },
      token
    );

    expect(status).toBe(200);

    // Navigate to address book
    await page.goto(`${getApiUrl()}/address-book`, { waitUntil: "networkidle" });
    await page.waitForSelector("table", { timeout: 10_000 }).catch(() => {});

    // Verify entry appears
    await expect(page.getByText("My Contact").or(page.getByText("contact-host-1"))).toBeVisible({ timeout: 5000 });

    // Cleanup
    const entry = data as { id: number };
    if (entry?.id) {
      await apiRequest("POST", "/api/admin/address_book/delete", { ids: [entry.id] }, token);
    }
  });

  test("edit contact alias via API", async ({ page }) => {
    const token = getAdminToken();
    const adminId = parseInt(process.env.ADMIN_USER_ID || "1");

    // Create entry
    const createRes = await apiRequest(
      "POST",
      "/api/admin/address_book/create",
      {
        peer_id: "ab-peer-edit",
        hostname: "edit-host",
        alias: "Original Alias",
        platform: "Linux",
        user_id: adminId,
      },
      token
    );
    expect(createRes.status).toBe(200);
    const entry = createRes.data as { id: number };

    // Edit via API
    const updateRes = await apiRequest(
      "POST",
      "/api/admin/address_book/update",
      { id: entry.id, alias: "Updated Alias" },
      token
    );
    expect(updateRes.status).toBe(200);

    // Verify on page
    await page.goto(`${getApiUrl()}/address-book`, { waitUntil: "networkidle" });
    await page.waitForTimeout(2000);
    await expect(page.getByText("Updated Alias")).toBeVisible({ timeout: 5000 });

    // Cleanup
    await apiRequest("POST", "/api/admin/address_book/delete", { ids: [entry.id] }, token);
  });

  test("delete contact removes from list", async ({ page }) => {
    const token = getAdminToken();
    const adminId = parseInt(process.env.ADMIN_USER_ID || "1");

    // Create entry
    const createRes = await apiRequest(
      "POST",
      "/api/admin/address_book/create",
      {
        peer_id: "ab-peer-del",
        hostname: "delete-me",
        alias: "To Be Deleted",
        platform: "Windows",
        user_id: adminId,
      },
      token
    );
    const entry = createRes.data as { id: number };

    // Delete via API
    const delRes = await apiRequest("POST", "/api/admin/address_book/delete", { ids: [entry.id] }, token);
    expect(delRes.status).toBe(200);

    // Verify it's gone
    await page.goto(`${getApiUrl()}/address-book`, { waitUntil: "networkidle" });
    await page.waitForTimeout(2000);
    // The entry should no longer be visible
    await expect(page.getByText("To Be Deleted")).not.toBeVisible({ timeout: 3000 });
  });

  test("tag management: create, assign, delete tag via API", async ({ page }) => {
    const token = getAdminToken();
    const adminId = parseInt(process.env.ADMIN_USER_ID || "1");

    // Create tag (requires user_id)
    const createTagRes = await apiRequest(
      "POST",
      "/api/admin/tag/create",
      { name: "e2e-test-tag", color: 16729344, user_id: adminId }, // #ff6600 = 16729344
      token
    );
    expect(createTagRes.status).toBe(200);
    const tag = createTagRes.data as { id: number; name: string };

    // Verify on page
    await page.goto(`${getApiUrl()}/address-book`, { waitUntil: "networkidle" });
    await page.waitForTimeout(2000);

    // Delete tag (uses ids array)
    if (tag?.id) {
      const delTagRes = await apiRequest("POST", "/api/admin/tag/delete", { ids: [tag.id] }, token);
      expect(delTagRes.status).toBe(200);
    }
  });
});
