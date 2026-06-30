import { defineConfig } from "@playwright/test";

const isCI = !!process.env.CI;

export default defineConfig({
  testDir: ".",
  timeout: 30_000,
  expect: {
    timeout: 10_000,
  },
  retries: isCI ? 1 : 0,
  workers: 1,
  reporter: [
    [isCI ? "github" : "list"],
    ["html", { open: "never" }],
  ],
  use: {
    baseURL: process.env.API_URL ?? "http://localhost:21114",
    trace: isCI ? "on-first-retry" : "off",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: {
        browserName: "chromium",
      },
    },
  ],
  globalSetup: "./global-setup.ts",
  globalTeardown: "./global-teardown.ts",
  // Let global-setup manage the server
  webServer: undefined,
});
