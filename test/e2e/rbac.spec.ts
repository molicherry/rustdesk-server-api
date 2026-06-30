import { test, expect } from "@playwright/test";
import {
  getAdminToken,
  apiRequest,
  createTestUserViaApi,
  deleteTestUserViaApi,
} from "./helpers";

test.describe("RBAC (Role-Based Access Control)", () => {
  let orgId: number;

  test.beforeAll(async () => {
    const token = getAdminToken();
    // Create an organization for testing
    const { status, data } = await apiRequest(
      "POST",
      "/api/admin/organizations/create",
      { name: "E2E Test Org", description: "Organization for RBAC E2E tests" },
      token
    );
    expect(status).toBe(200);
    const org = data as { id: number };
    orgId = org.id;
    expect(orgId).toBeGreaterThan(0);
  });

  test.afterAll(async () => {
    const token = getAdminToken();
    if (orgId) {
      await apiRequest(
        "POST",
        "/api/admin/organizations/delete",
        { id: orgId },
        token
      ).catch(() => {});
    }
  });

  test("create org appears in list", async () => {
    const token = getAdminToken();
    const { status, data } = await apiRequest(
      "GET",
      "/api/admin/organizations/list",
      undefined,
      token
    );
    expect(status).toBe(200);
    const result = data as { data: { id: number; name: string }[] };
    expect(result.data.some((o) => o.name === "E2E Test Org")).toBe(true);
  });

  test("add user to org as org_admin", async () => {
    const token = getAdminToken();
    const username = `e2e-orgadmin-${Date.now()}`;

    // Create a user
    const user = await createTestUserViaApi(username, "TestPass123!", false, "user");

    // Add user to org as org_admin
    const { status, data } = await apiRequest(
      "POST",
      `/api/admin/organizations/${orgId}/users/add`,
      {
        user_id: user.id,
        role: "org_admin",
      },
      token
    );
    expect(status).toBe(200);
    const membership = data as { id: number; role: string };
    expect(membership.role).toBe("org_admin");

    // Cleanup: remove user from org first, then delete user
    await apiRequest(
      "POST",
      `/api/admin/organizations/${orgId}/users/remove`,
      { user_id: user.id },
      token
    ).catch(() => {});
    await deleteTestUserViaApi(user.id);
  });

  test("org_admin can see org devices via API", async () => {
    const token = getAdminToken();
    const username = `e2e-orgview-${Date.now()}`;

    // Create user and add to org
    const user = await createTestUserViaApi(username, "TestPass123!", false, "user");
    await apiRequest(
      "POST",
      `/api/admin/organizations/${orgId}/users/add`,
      { user_id: user.id, role: "org_admin" },
      token
    );

    // Login as this user to get their token
    const loginRes = await apiRequest("POST", "/api/admin/login", {
      username,
      password: "TestPass123!",
      captcha_token: "bypass",
    });
    expect(loginRes.status).toBe(200);
    const userToken = (loginRes.data as { token: string }).token;

    // This user should be able to see org devices (peer list)
    // The org context is usually set via header or query param
    const { status } = await apiRequest(
      "GET",
      `/api/admin/peer/list?page=1&page_size=20&org_id=${orgId}`,
      undefined,
      userToken
    );
    // Should succeed since user is org_admin
    expect(status).toBe(200);

    // Cleanup
    await apiRequest("POST", `/api/admin/organizations/${orgId}/users/remove`, { user_id: user.id }, token).catch(() => {});
    await deleteTestUserViaApi(user.id);
  });

  test("non-member user gets 403 for org resources", async () => {
    const token = getAdminToken();
    const username = `e2e-outsider-${Date.now()}`;

    // Create user but DON'T add to org
    const user = await createTestUserViaApi(username, "TestPass123!", false, "user");

    // Login as this user
    const loginRes = await apiRequest("POST", "/api/admin/login", {
      username,
      password: "TestPass123!",
      captcha_token: "bypass",
    });
    const userToken = (loginRes.data as { token: string }).token;

    // Try to access org-scoped endpoint without being a member
    const { status } = await apiRequest(
      "GET",
      `/api/admin/peer/list?page=1&page_size=20&org_id=${orgId}`,
      undefined,
      userToken
    );
    // Should get 403 since user is not in the org
    expect(status).toBe(403);

    // Cleanup
    await deleteTestUserViaApi(user.id);
  });
});
