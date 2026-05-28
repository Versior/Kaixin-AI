import { AUTH_TOKEN_KEY } from "@/services/api/auth";

type LocalStorageReader = (key: string) => string | null;

export function userScopedStorageKey(name: string, getItem?: LocalStorageReader) {
    if (typeof window === "undefined" && !getItem) return name;
    const read = getItem || ((key: string) => window.localStorage.getItem(key));
    try {
        const auth = JSON.parse(read(AUTH_TOKEN_KEY) || "{}");
        const token = auth?.state?.token || "guest";
        const payload = token && token !== "guest" ? JSON.parse(decodeBase64Url(token.split(".")[1] || "")) : null;
        return `${name}:${payload?.userId || "guest"}`;
    } catch {
        return `${name}:guest`;
    }
}

function decodeBase64Url(value: string) {
    if (typeof window === "undefined") {
        return Buffer.from(value, "base64url").toString("utf8");
    }
    const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), "=");
    return window.atob(padded);
}
