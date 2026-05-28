"use client";

import { useEffect, useState } from "react";
import { App, Button, Empty, Image, Input, Pagination, Spin, Tag, Typography } from "antd";
import { Download, Search } from "lucide-react";

import { fetchImageHistory, type ImageHistoryLog } from "@/services/api/image-history";

function previewUrl(url: string) {
    return url.startsWith("http") ? `/api/image-proxy?thumb=1&url=${encodeURIComponent(url)}` : url;
}

function downloadUrl(url: string) {
    return url.startsWith("http") ? `/api/image-proxy?url=${encodeURIComponent(url)}` : url;
}

function downloadImage(url: string, filename: string) {
    const link = document.createElement("a");
    link.href = downloadUrl(url);
    link.download = filename;
    link.rel = "noopener noreferrer";
    document.body.appendChild(link);
    link.click();
    link.remove();
}

function formatTime(value: string) {
    if (!value) return "-";
    const time = new Date(value);
    if (Number.isNaN(time.getTime())) return value;
    return time.toLocaleString();
}

export default function ImageHistoryPage() {
    const { message } = App.useApp();
    const [logs, setLogs] = useState<ImageHistoryLog[]>([]);
    const [total, setTotal] = useState(0);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(12);
    const [keyword, setKeyword] = useState("");
    const [keywordText, setKeywordText] = useState("");
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        let alive = true;
        setLoading(true);
        fetchImageHistory({ page, pageSize, keyword, type: "image" })
            .then((data) => {
                if (!alive) return;
                setLogs(data.items || []);
                setTotal(data.total || 0);
            })
            .catch((error) => {
                if (alive) message.error(error instanceof Error ? error.message : "读取生成历史失败");
            })
            .finally(() => {
                if (alive) setLoading(false);
            });
        return () => {
            alive = false;
        };
    }, [page, pageSize, keyword, message]);

    const search = () => {
        setPage(1);
        setKeyword(keywordText.trim());
    };

    return (
        <main className="h-full overflow-auto bg-stone-50 px-6 py-6 dark:bg-stone-950">
            <div className="mx-auto flex max-w-7xl flex-col gap-5">
                <section className="flex flex-col gap-3 rounded-3xl border border-stone-200 bg-white p-5 shadow-sm dark:border-stone-800 dark:bg-stone-900">
                    <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
                        <div>
                            <Typography.Title level={3} style={{ margin: 0 }}>云端生成历史</Typography.Title>
                            <Typography.Text type="secondary">按账号保存，换浏览器或重新登录后仍可查看。</Typography.Text>
                        </div>
                        <Input.Search
                            allowClear
                            value={keywordText}
                            placeholder="搜索提示词 / 模型 / 状态"
                            enterButton={<Search className="size-4" />}
                            onChange={(event) => setKeywordText(event.target.value)}
                            onSearch={search}
                            className="max-w-md"
                        />
                    </div>
                </section>

                <Spin spinning={loading}>
                    {logs.length ? (
                        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                            {logs.map((log) => (
                                <article key={log.id} className="overflow-hidden rounded-3xl border border-stone-200 bg-white shadow-sm dark:border-stone-800 dark:bg-stone-900">
                                    <div className="grid grid-cols-3 gap-1 bg-stone-100 p-1 dark:bg-stone-950">
                                        {(log.images || []).slice(0, 3).map((url, index) => (
                                            <Image key={`${log.id}-${index}`} src={previewUrl(url)} alt={log.prompt || log.id} height={120} width="100%" style={{ objectFit: "cover", borderRadius: 14 }} preview={{ src: url }} fallback="" />
                                        ))}
                                        {!log.images?.length ? <div className="col-span-3 flex h-[120px] items-center justify-center text-sm text-stone-400">无图片预览</div> : null}
                                    </div>
                                    <div className="space-y-3 p-4">
                                        <div className="flex items-center justify-between gap-3">
                                            <Tag color={log.status === "success" || log.status === "succeeded" ? "green" : log.status === "partial_success" ? "orange" : "red"}>{log.status || "unknown"}</Tag>
                                            <span className="text-xs text-stone-500">{formatTime(log.createdAt)}</span>
                                        </div>
                                        <Typography.Paragraph ellipsis={{ rows: 3 }} className="!mb-0 text-sm">{log.prompt || "无提示词"}</Typography.Paragraph>
                                        <div className="flex items-center justify-between gap-2 text-xs text-stone-500">
                                            <span>{log.model || "-"}</span>
                                            <span>{log.images?.length || 0} 张</span>
                                        </div>
                                        <div className="flex flex-wrap gap-2">
                                            {(log.images || []).map((url, index) => (
                                                <Button key={`${log.id}-download-${index}`} size="small" icon={<Download className="size-3.5" />} onClick={() => downloadImage(url, `history-${log.id}-${index + 1}.png`)}>
                                                    下载{index + 1}
                                                </Button>
                                            ))}
                                        </div>
                                    </div>
                                </article>
                            ))}
                        </div>
                    ) : (
                        <div className="rounded-3xl border border-dashed border-stone-300 bg-white py-20 dark:border-stone-800 dark:bg-stone-900">
                            <Empty description="还没有云端生成历史" />
                        </div>
                    )}
                </Spin>

                <div className="flex justify-center">
                    <Pagination current={page} pageSize={pageSize} total={total} showSizeChanger onChange={(nextPage, nextPageSize) => { setPage(nextPage); setPageSize(nextPageSize); }} />
                </div>
            </div>
        </main>
    );
}
