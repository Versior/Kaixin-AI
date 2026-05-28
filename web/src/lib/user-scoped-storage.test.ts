import { describe, expect, test } from "bun:test";

import { userScopedStorageKey } from "./user-scoped-storage";

function tokenFor(userId: string) {
    const payload = Buffer.from(JSON.stringify({ userId })).toString("base64url");
    return `header.${payload}.signature`;
}

describe("userScopedStorageKey", () => {
    test("uses the persisted auth token key to isolate storage by user id", () => {
        const getItem = (key: string) => {
            expect(key).toBe("infinite-canvas-auth-token-v1");
            return JSON.stringify({ state: { token: tokenFor("user-1") } });
        };

        expect(userScopedStorageKey("infinite-canvas:canvas_store", getItem)).toBe("infinite-canvas:canvas_store:user-1");
    });

    test("falls back to guest when token is absent or invalid", () => {
        expect(userScopedStorageKey("infinite-canvas:canvas_store", () => null)).toBe("infinite-canvas:canvas_store:guest");
        expect(userScopedStorageKey("infinite-canvas:canvas_store", () => "bad-json")).toBe("infinite-canvas:canvas_store:guest");
    });
});
