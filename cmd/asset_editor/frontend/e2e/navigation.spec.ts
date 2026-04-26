import { test, expect } from "@playwright/test";

test.describe("navigation", () => {
    test("root redirects to sprites", async ({ page }) => {
        await page.goto("/");
        await expect(page).toHaveURL(/\/sprites/);
    });

    test("nav rail links load each section", async ({ page }) => {
        await page.goto("/sprites");

        const sections = [
            { title: "Sprites", url: /\/sprites/ },
            { title: "Audio", url: /\/audio/ },
            { title: "Scripts", url: /\/scripts/ },
            { title: "Skills", url: /\/skills/ },
            { title: "Primortals", url: /\/primortals/ },
            { title: "Combat", url: /\/combat/ },
            { title: "Saves", url: /\/saves/ },
            { title: "Reference", url: /\/reference/ },
            { title: "Debug", url: /\/debug/ },
        ];

        for (const section of sections) {
            await page.locator(`nav a[title="${section.title}"]`).click();
            await expect(page).toHaveURL(section.url);
        }
    });
});
