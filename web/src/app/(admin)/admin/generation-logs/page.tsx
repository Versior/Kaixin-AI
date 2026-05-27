"use client";

import { CopyOutlined, DeleteOutlined, EyeOutlined, ReloadOutlined, SearchOutlined } from "@ant-design/icons";
import { ProTable, type ProColumns } from "@ant-design/pro-components";
import { Button, Card, Col, Form, Image, Input, Modal, Popconfirm, Row, Select, Space, Tag, Tooltip, Typography } from "antd";
import dayjs from "dayjs";
import { useEffect, useState } from "react";

import { useCopyText } from "@/hooks/use-copy-text";
import type { AdminGenerationLog } from "@/services/api/admin";
import { useAdminGenerationLogs } from "./use-admin-generation-logs";

const typeOptions = [{ label: "全部类型", value: "" }, { label: "图片", value: "image" }, { label: "对话", value: "chat" }];

function previewUrl(url: string) {
    return url.startsWith("http") ? `/api/image-proxy?thumb=1&url=${encodeURIComponent(url)}` : url;
}

export default function AdminGenerationLogsPage() {
    const { logs, keyword, kind, page, pageSize, total, isLoading, searchLogs, changeKind, changePage, changePageSize, resetFilters, refreshLogs, deleteLog, deletePageLogs } = useAdminGenerationLogs();
    const [keywordText, setKeywordText] = useState(keyword);
    const [detailLog, setDetailLog] = useState<AdminGenerationLog | null>(null);
    const [deletingLog, setDeletingLog] = useState<AdminGenerationLog | null>(null);
    const copyText = useCopyText();
    useEffect(() => setKeywordText(keyword), [keyword]);

    const columns: ProColumns<AdminGenerationLog>[] = [
        { title: "预览", dataIndex: "images", width: 92, render: (_, item) => item.images?.[0] ? <Image src={previewUrl(item.images[0])} alt={item.prompt || item.id} width={44} height={34} style={{ objectFit: "cover", borderRadius: 6 }} preview={{ mask: "放大" }} fallback="" /> : <Typography.Text type="secondary">无预览</Typography.Text> },
        { title: "用户名", dataIndex: "username", width: 180, render: (_, item) => <Typography.Text copyable ellipsis>{item.username || "-"}</Typography.Text> },
        { title: "类型", dataIndex: "kind", width: 90, render: (_, item) => <Tag color={item.kind === "image" ? "purple" : "blue"}>{item.kind === "image" ? "图片" : "对话"}</Tag> },
        { title: "模型", dataIndex: "model", width: 180, render: (_, item) => <Typography.Text ellipsis>{item.model}</Typography.Text> },
        { title: "提示词 / 对话", dataIndex: "prompt", ellipsis: true, render: (_, item) => <Typography.Text type="secondary" ellipsis>{item.prompt || "-"}</Typography.Text> },
        { title: "状态", dataIndex: "status", width: 100, render: (_, item) => <Tag color={item.status === "success" ? "green" : "red"}>{item.status}</Tag> },
        { title: "时间", dataIndex: "createdAt", width: 180, render: (_, item) => <Typography.Text type="secondary">{item.createdAt ? dayjs(item.createdAt).format("YYYY-MM-DD HH:mm:ss") : "-"}</Typography.Text> },
        { title: "操作", key: "actions", width: 96, align: "right", render: (_, item) => <Space size={4}><Tooltip title="详情"><Button type="text" size="small" icon={<EyeOutlined />} onClick={() => setDetailLog(item)} /></Tooltip><Tooltip title="删除"><Button danger type="text" size="small" icon={<DeleteOutlined />} onClick={() => setDeletingLog(item)} /></Tooltip></Space> },
    ];

    return <main style={{ padding: 24 }}><Space direction="vertical" size={16} style={{ width: "100%" }}><Card variant="borderless"><Form layout="vertical"><Row gutter={16} align="bottom"><Col flex="360px"><Form.Item label="关键词"><Input.Search value={keywordText} placeholder="搜索用户名、模型、提示词、状态" allowClear enterButton={<SearchOutlined />} onSearch={() => searchLogs(keywordText)} onChange={(event) => setKeywordText(event.target.value)} /></Form.Item></Col><Col flex="180px"><Form.Item label="类型"><Select value={kind} onChange={changeKind} options={typeOptions} /></Form.Item></Col><Col flex="none"><Form.Item><Space><Button onClick={() => { setKeywordText(""); resetFilters(); }}>重置</Button><Button type="primary" icon={<ReloadOutlined />} onClick={() => searchLogs(keywordText)}>查询</Button></Space></Form.Item></Col></Row></Form></Card><ProTable<AdminGenerationLog> rowKey="id" columns={columns} dataSource={logs} loading={isLoading} search={false} defaultSize="middle" tableLayout="fixed" cardProps={{ variant: "borderless" }} headerTitle={<Space><Typography.Text strong>生成日志与图片管理</Typography.Text><Tag>{total} 条</Tag></Space>} options={{ density: true, setting: true, reload: () => void refreshLogs() }} toolBarRender={() => [<Popconfirm key="delete-page" title="删除本页日志" description={`确定删除本页 ${logs.length} 条生成日志？`} okText="删除" cancelText="取消" okButtonProps={{ danger: true }} onConfirm={() => void deletePageLogs(logs.map((item) => item.id))}><Button danger disabled={!logs.length} icon={<DeleteOutlined />}>删除本页内容</Button></Popconfirm>]} pagination={{ current: page, pageSize, total, showSizeChanger: true, pageSizeOptions: [8, 10, 20], showTotal: (value) => `共 ${value} 条`, onChange: (nextPage, nextPageSize) => (nextPageSize !== pageSize ? changePageSize(nextPageSize) : changePage(nextPage)) }} /></Space><Modal title="生成详情" open={Boolean(detailLog)} width={920} onCancel={() => setDetailLog(null)} footer={<Button onClick={() => setDetailLog(null)}>关闭</Button>}>{detailLog ? <Space direction="vertical" size={14} style={{ width: "100%" }}><Space wrap><Tag>{detailLog.kind === "image" ? "图片" : "对话"}</Tag><Tag>{detailLog.status}</Tag><Tag>{detailLog.model}</Tag><Typography.Text type="secondary">{detailLog.createdAt ? dayjs(detailLog.createdAt).format("YYYY-MM-DD HH:mm:ss") : ""}</Typography.Text></Space>{detailLog.images?.length ? <Image.PreviewGroup><Space wrap>{detailLog.images.map((url, index) => <Image key={`${url}-${index}`} src={previewUrl(url)} alt={`image-${index + 1}`} width={96} height={72} style={{ objectFit: "cover", borderRadius: 8 }} fallback="" />)}</Space></Image.PreviewGroup> : null}<Typography.Text strong>提示词 / 对话</Typography.Text><Input.TextArea value={detailLog.prompt} rows={5} readOnly />{detailLog.error ? <><Typography.Text strong type="danger">错误</Typography.Text><Input.TextArea value={detailLog.error} rows={3} readOnly /></> : null}<Typography.Text strong>请求</Typography.Text><Input.TextArea value={detailLog.request} rows={6} readOnly /><Typography.Text strong>响应</Typography.Text><Input.TextArea value={detailLog.response} rows={6} readOnly /><Button icon={<CopyOutlined />} onClick={() => copyText(JSON.stringify(detailLog, null, 2))}>复制完整记录</Button></Space> : null}</Modal><Modal title="删除记录" open={Boolean(deletingLog)} onCancel={() => setDeletingLog(null)} onOk={async () => { if (!deletingLog) return; await deleteLog(deletingLog.id); setDeletingLog(null); }} okText="删除" okButtonProps={{ danger: true }} cancelText="取消"><Typography.Text>确认删除这条生成记录？不会删除用户已保存的本地灵感内容。</Typography.Text></Modal></main>;
}
