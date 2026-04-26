import { test, expect } from "@playwright/test";

test.describe("script browser", () => {
    test("lists script files with handler counts", async ({ page }) => {
        await page.goto("/scripts");
        await expect(page.getByText("main").first()).toBeVisible();
        await expect(page.getByText(/handler/).first()).toBeVisible();
    });

    test("clicking a script opens the editor", async ({ page }) => {
        await page.goto("/scripts");
        await page.getByText("main").first().click();
        await expect(page).toHaveURL(/\/scripts\//);
    });
});

test.describe("script editor", () => {
    test.beforeEach(async ({ page }) => {
        await page.goto("/scripts/hq/main.yaml");
    });

    test("shows handler list in left panel", async ({ page }) => {
        await expect(page.getByRole("button", { name: "hq.xenolog" })).toBeVisible();
        await expect(page.getByRole("button", { name: "hq.bed" })).toBeVisible();
        await expect(page.getByRole("button", { name: "hq.animech" })).toBeVisible();
    });

    test("selecting a handler shows its detail", async ({ page }) => {
        await page.getByRole("button", { name: "hq.bed" }).click();
        await expect(page.getByText("on_interact_self")).toBeVisible();
    });

    test("raw YAML tab shows editable code editor", async ({ page }) => {
        await page.getByRole("tab", { name: "Raw YAML" }).click();
        const editor = page.locator(".cm-editor");
        await expect(editor).toBeVisible();
        await expect(page.locator(".cm-content")).toContainText("handlers:");
    });

    test("switching back to structured tab preserves content", async ({ page }) => {
        await page.getByRole("tab", { name: "Raw YAML" }).click();
        await expect(page.locator(".cm-editor")).toBeVisible();
        await page.getByRole("tab", { name: "Structured" }).click();
        await expect(page.getByRole("button", { name: "hq.xenolog" })).toBeVisible();
    });
});
