import { NextRequest } from "next/server";

export const runtime = "nodejs";
export const maxDuration = 120;

export async function GET(request: NextRequest) {
    const url = request.nextUrl.searchParams.get("url") || "";
    if (!/^https?:\/\//i.test(url)) {
        return Response.json({ code: 1, msg: "图片地址无效" }, { status: 400 });
    }
    const response = await fetch(url, { redirect: "follow" });
    if (!response.ok || !response.body) {
        return Response.json({ code: 1, msg: `图片下载失败：${response.status}` }, { status: 502 });
    }
    const headers = new Headers();
    headers.set("content-type", response.headers.get("content-type") || "image/png");
    headers.set("cache-control", "private, max-age=3600");
    return new Response(response.body, { status: 200, headers });
}
