import { NextRequest } from "next/server";

export const runtime = "nodejs";
export const maxDuration = 120;

export async function GET(request: NextRequest) {
    const url = request.nextUrl.searchParams.get("url") || "";
    const thumb = request.nextUrl.searchParams.get("thumb") === "1";
    if (!/^https?:\/\//i.test(url)) {
        return Response.json({ code: 1, msg: "图片地址无效" }, { status: 400 });
    }
    const response = await fetch(url, { redirect: "follow" });
    if (!response.ok || !response.body) {
        return Response.json({ code: 1, msg: `图片下载失败：${response.status}` }, { status: 502 });
    }
    const contentType = response.headers.get("content-type") || "image/png";
    const headers = new Headers();
    headers.set("cache-control", "private, max-age=3600");
    if (thumb && contentType.startsWith("image/")) {
        try {
            const sharp = (await import("sharp")).default;
            const input = Buffer.from(await response.arrayBuffer());
            const output = await sharp(input).resize({ width: 220, height: 160, fit: "inside", withoutEnlargement: true }).webp({ quality: 72 }).toBuffer();
            headers.set("content-type", "image/webp");
            return new Response(output, { status: 200, headers });
        } catch {
            // Fall back to the original image stream when sharp is unavailable.
        }
    }
    headers.set("content-type", contentType);
    return new Response(response.body, { status: 200, headers });
}
