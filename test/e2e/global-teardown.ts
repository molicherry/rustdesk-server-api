import { readFileSync, rmSync } from "node:fs";
import type { FullConfig } from "@playwright/test";

async function globalTeardown(_config: FullConfig): Promise<void> {
  const pidFile = process.env.E2E_PID_FILE;
  const tempDir = process.env.E2E_TEMP_DIR;

  if (pidFile) {
    try {
      const pid = parseInt(readFileSync(pidFile, "utf-8").trim(), 10);
      if (pid && pid > 0) {
        console.log(`[global-teardown] Killing server process ${pid}...`);
        try {
          process.kill(pid, "SIGTERM");
        } catch {
          // Process might already be dead
          console.log(`[global-teardown] Process ${pid} already gone`);
        }
      }
    } catch (err) {
      console.warn(`[global-teardown] Failed to read PID file: ${err}`);
    }
  }

  // Also try to kill any remaining go processes on the test port
  if (tempDir) {
    try {
      // Small delay to let the process shut down gracefully
      await new Promise((r) => setTimeout(r, 2000));

      // Clean up temp directory
      rmSync(tempDir, { recursive: true, force: true });
      console.log(`[global-teardown] Cleaned up ${tempDir}`);
    } catch (err) {
      console.warn(`[global-teardown] Cleanup warning: ${err}`);
    }
  }

  console.log("[global-teardown] Done.");
}

export default globalTeardown;
