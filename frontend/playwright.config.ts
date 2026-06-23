import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright E2E config. Tests target a running frontend (default
 * http://localhost:3000). Set E2E_BASE_URL to point at a different host.
 * Set E2E_API_BASE_URL to the backend's base for registration setup.
 */
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: "list",
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3000",
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
