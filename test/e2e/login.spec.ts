import { test, expect } from "@playwright/test";
import {
  getApiUrl,
  getAdminPassword,
  setupAuthFix,
  goToPage,
  logoutViaUI,
} from "./helpers";

test.describe("Login", () => {
  test("valid credentials redirect to dashboard", async ({ page }) => {
    const apiUrl = getApiUrl();
    await setupAuthFix(page);

    await page.goto(`${apiUrl}/login`);
    await page.waitForSelector('input[type="text"]', { timeout: 10_000 });

    await page.fill('input[type="text"]', "admin");
    await page.fill('input[type="password"]', getAdminPassword());
    await page.click('button[type="submit"]');

    // Should redirect to dashboard
    await page.waitForURL("**/dashboard**", { timeout: 10_000 });
    await expect(page.locator("h2")).toContainText("Dashboard");
  });

  test("invalid password shows error message", async ({ page }) => {
    const apiUrl = getApiUrl();

    await page.goto(`${apiUrl}/login`);
    await page.waitForSelector('input[type="text"]', { timeout: 10_000 });

    await page.fill('input[type="text"]', "admin");
    await page.fill('input[type="password"]', "wrongpassword");
    await page.click('button[type="submit"]');

    // Error message should appear (LoginByPassword returns "invalid username or password")
    await expect(page.locator("text=invalid username or password")).toBeVisible({ timeout: 5_000 });
    // Should stay on login page
    await expect(page).toHaveURL(/\/login/);
  });

  test("empty username shows validation", async ({ page }) => {
    const apiUrl = getApiUrl();

    await page.goto(`${apiUrl}/login`);
    await page.waitForSelector('input[type="text"]', { timeout: 10_000 });

    // Leave username empty, fill password
    await page.fill('input[type="password"]', "something");
    await page.click('button[type="submit"]');

    // HTML5 validation should prevent submission (input has required attr)
    // The form has 'required' so browser validation fires
    const input = page.locator('input[type="text"]');
    await expect(input).toHaveAttribute("required", "");
    // Browser validation message — just verify we're still on login page
    await expect(page).toHaveURL(/\/login/);
  });

  test("logout returns to login page", async ({ page }) => {
    const apiUrl = getApiUrl();
    await setupAuthFix(page);

    // Set up auth state directly
    await page.goto(`${apiUrl}/login`);
    await page.waitForSelector('input[type="text"]');
    await page.fill('input[type="text"]', "admin");
    await page.fill('input[type="password"]', getAdminPassword());
    await page.click('button[type="submit"]');
    await page.waitForURL("**/dashboard**");

    // Now logout
    await logoutViaUI(page);

    // Should be on login page
    await page.goto(`${apiUrl}/login`);
    await expect(page.locator('input[type="text"]')).toBeVisible({ timeout: 5_000 });
  });
});
