import { test, expect } from "@playwright/test";

const API = "http://localhost:8090/api/v1/scripts/_validate-expr";

test.describe("expression validation API", () => {
    test("valid simple expression", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "1 + 2" } });
        expect(res.ok()).toBeTruthy();
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid env variable access", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "var.count + 1" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid function call", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "min(var.a, var.b)" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid clamp function", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "clamp(var.x, 0, 100)" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid hasKey function", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "hasKey(var.data, 'key')" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid string comparison", async ({ request }) => {
        const res = await request.post(API, { data: { expression: 'global.stage == "intro"' } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid boolean logic", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "var.a > 0 && var.b < 10" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid ternary", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "var.x > 0 ? 'yes' : 'no'" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid array index", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "var.items[var.idx]" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("valid nested function", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "min(var.count, len(var.items) - 1)" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("invalid syntax returns error", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "var.count +" } });
        const body = await res.json();
        expect(body.valid).toBe(false);
        expect(body.error).toBeTruthy();
    });

    test("unknown function returns error", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "nonexistent(1)" } });
        const body = await res.json();
        expect(body.valid).toBe(false);
        expect(body.error).toBeTruthy();
    });

    test("empty expression is valid", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("save data access", async ({ request }) => {
        const res = await request.post(API, { data: { expression: "save.animech_level" } });
        expect(await res.json()).toEqual({ valid: true });
    });

    test("self/player/source string vars", async ({ request }) => {
        const exprs = ["self", "player", "source"];
        for (const expr of exprs) {
            const res = await request.post(API, { data: { expression: expr } });
            expect(await res.json()).toEqual({ valid: true });
        }
    });
});
