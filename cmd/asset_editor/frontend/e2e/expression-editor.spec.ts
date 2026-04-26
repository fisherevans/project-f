import { test, expect } from "@playwright/test";

test.describe("expression editor in script UI", () => {
    test("add if step and interact with expression input", async ({ page }) => {
        await page.goto("/scripts/hq/main.yaml");

        await page.getByRole("button", { name: "hq.bed" }).click();
        await expect(page.getByText("on_interact_self")).toBeVisible();

        const addBtn = page.getByRole("button", { name: /add step/i }).first();
        await addBtn.click();

        // Step kind picker should appear
        const searchInput = page.getByPlaceholder(/search/i);
        await expect(searchInput).toBeVisible({ timeout: 3000 });
        await searchInput.fill("if");
        await page.getByText("if", { exact: true }).first().click();

        // CodeMirror expression input should render for the "when" param
        const cmEditor = page.locator(".cm-editor").first();
        await expect(cmEditor).toBeVisible({ timeout: 3000 });

        // Type into the editor and trigger autocomplete
        const cmContent = cmEditor.locator(".cm-content");
        await cmContent.click();
        // Type "mi" to match "min" function
        await page.keyboard.type("mi", { delay: 50 });

        // Trigger autocomplete explicitly (activateOnTyping may not fire with programmatic input)
        await page.keyboard.press("Control+Space");
        const tooltip = page.locator(".cm-tooltip-autocomplete");
        await expect(tooltip).toBeVisible({ timeout: 5000 });
    });

    test("expression input shows lint error for invalid expression", async ({ page }) => {
        await page.goto("/scripts/hq/main.yaml");

        await page.getByRole("button", { name: "hq.bed" }).click();
        await expect(page.getByText("on_interact_self")).toBeVisible();

        const addBtn = page.getByRole("button", { name: /add step/i }).first();
        await addBtn.click();

        const searchInput = page.getByPlaceholder(/search/i);
        await expect(searchInput).toBeVisible({ timeout: 3000 });
        await searchInput.fill("if");
        await page.getByText("if", { exact: true }).first().click();

        const cmEditor = page.locator(".cm-editor").first();
        await expect(cmEditor).toBeVisible({ timeout: 3000 });

        const cmContent = cmEditor.locator(".cm-content");
        await cmContent.click();
        await page.keyboard.type("invalid_func()", { delay: 30 });

        // Wait for lint to fire (500ms debounce + network)
        const lintMark = page.locator(".cm-lintRange-error");
        await expect(lintMark).toBeVisible({ timeout: 5000 });
    });
});
