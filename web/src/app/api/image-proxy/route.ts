import type { NextRequest } from "next/server";

export const runtime = "nodejs";
export const maxDuration = 300;

const MAX_IMAGE_BYTES = 25 * 1024 * 1024;

function rewriteInternalImageUrl(url: URL) {
    if (url.hostname === "183.87.136.115" && url.port === "3000") {
        url.hostname = "172.17.0.1";
    }
    return url;
}

export async function GET(request: NextRequest) {
    const rawUrl = request.nextUrl.searchParams.get("url") || "";
    let url: URL;
    try {
        url = new URL(rawUrl);
    } catch {
        return Response.json({ code: 1, data: null, msg: "图片地址无效" }, { status: 400 });
    }
    if (url.protocol !== "http:" && url.protocol !== "https:") {
        return Response.json({ code: 1, data: null, msg: "图片地址无效" }, { status: 400 });
    }
    url = rewriteInternalImageUrl(url);

    try {
        const response = await fetch(url, { redirect: "follow" });
        if (!response.ok || !response.body) {
            return Response.json({ code: 1, data: null, msg: `图片下载失败：${response.status}` }, { status: 502 });
        }
        const contentType = response.headers.get("content-type") || "application/octet-stream";
        if (!contentType.startsWith("image/")) {
            return Response.json({ code: 1, data: null, msg: "远程内容不是图片" }, { status: 502 });
        }
        const contentLength = Number(response.headers.get("content-length") || 0);
        if (contentLength > MAX_IMAGE_BYTES) {
            return Response.json({ code: 1, data: null, msg: "图片过大" }, { status: 502 });
        }
        const headers = new Headers({ "content-type": contentType, "cache-control": "public, max-age=86400" });
        return new Response(response.body, { status: 200, headers });
    } catch (error) {
        console.error("Failed to proxy image", url.toString(), error);
        return Response.json({ code: 1, data: null, msg: "图片代理请求失败" }, { status: 502 });
    }
}
