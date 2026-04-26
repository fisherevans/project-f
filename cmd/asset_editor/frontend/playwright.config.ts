import { defineConfig } from "@playwright/test";

export default defineConfig({
    testDir: "./e2e",
    timeout: 30_000,
    expect: { timeout: 5_000 },
    fullyParallel: true,
    retries: 0,
    reporter: "list",
    use: {
        baseURL: "http://localhost:5173",
        screenshot: "only-on-failure",
        trace: "retain-on-failure",
    },
    projects: [
        { name: "chromium", use: { browserName: "chromium" } },
    ],
    webServer: [
        {
            command: "GOWORK=off go run ../../.. -dev -port 8090",
            cwd: "../",
            port: 8090,
            reuseExistingServer: true,
            timeout: 30_000,
        },
        {
            command: "npx vite --port 5173",
            port: 5173,
            reuseExistingServer: true,
            timeout: 15_000,
        },
    ],
});
