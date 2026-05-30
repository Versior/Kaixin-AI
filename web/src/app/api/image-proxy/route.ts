import type { NextRequest } from "next/server";

export const runtime = "nodejs";
export const maxDuration = 300;

const MAX_IMAGE_BYTES = 25 * 1024 * 1024;

/**
 * 内部图片地址重写映射。
 * 当容器内需要访问宿主机上的 chatgpt2api 服务时，
 * 将回环地址替换为 docker0 网关地址。
 *
 * 如果你的部署架构不同，请修改此映射或通过环境变量 INTERNAL_IMAGE_HOST_REWRITES 覆盖。
 * 格式：JSON 对象，键为原始主机名，值为目标主机名。
 * 例如：{"localhost":"host.docker.internal","127.0.0.1":"host.docker.internal"}
 */
function getInternalImageHostRewrites(): Record<string, string> {
    const envOverride = process.env.INTERNAL_IMAGE_HOST_REWRITES;
    if (envOverride) {
        try {
            return JSON.parse(envOverride);
        } catch {
            // ignore invalid JSON
        }
    }
    return {
        "127.0.0.1": "host.docker.internal",
        "localhost": "host.docker.internal",
        "[::1]": "host.docker.internal",
    };
}

function rewriteInternalImageUrl(url: URL) {
    const rewrites = getInternalImageHostRewrites();
    const targetHost = rewrites[url.hostname];
    if (targetHost && url.port === "3000") {
        url.hostname = targetHost;
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
