"use client";

import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { App } from "antd";

import { deleteAdminGenerationLog, fetchAdminGenerationLogs, fetchAdminUsers, type AdminGenerationLog, type AdminUser } from "@/services/api/admin";
import { useUserStore } from "@/stores/use-user-store";

const defaultPageSize = 8;

export function useAdminGenerationLogs() {
    const { message } = App.useApp();
    const queryClient = useQueryClient();
    const token = useUserStore((state) => state.token);
    const clearSession = useUserStore((state) => state.clearSession);
    const [keyword, setKeyword] = useState("");
    const [type, setType] = useState("");
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(defaultPageSize);

    const usersQuery = useQuery({
        queryKey: ["admin", "users-for-generation-logs", token],
        queryFn: () => fetchAdminUsers(token, { page: 1, pageSize: 1000 }),
        enabled: Boolean(token),
        retry: false,
    });

    const query = useQuery({
        queryKey: ["admin", "generation-logs", token, keyword, type, page, pageSize],
        queryFn: () => fetchAdminGenerationLogs(token, { keyword, type, page, pageSize }),
        enabled: Boolean(token),
        retry: false,
    });

    const usersById = new Map((usersQuery.data?.items || []).map((user: AdminUser) => [user.id, user]));
    const logs: AdminGenerationLog[] = (query.data?.items || []).map((log) => {
        const user = usersById.get(log.userId);
        return { ...log, username: user?.displayName || user?.username || log.username || "-" };
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteAdminGenerationLog(token, id),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ["admin", "generation-logs"] });
            message.success("记录已删除");
        },
        onError: (error) => message.error(error instanceof Error ? error.message : "删除失败"),
    });

    useEffect(() => {
        if (query.isError) {
            const errorMessage = query.error instanceof Error ? query.error.message : "读取生成记录失败";
            message.error(errorMessage);
            if (errorMessage.includes("未登录") || errorMessage.includes("权限不足") || errorMessage.includes("登录状态无效")) clearSession();
        }
    }, [clearSession, message, query.error, query.isError]);

    const updateFilters = (next: Partial<{ keyword: string; type: string; page: number; pageSize: number }>) => {
        const queryState = { keyword, type, page, pageSize, ...next };
        if (next.keyword !== undefined || next.type !== undefined || next.pageSize !== undefined) queryState.page = 1;
        setKeyword(queryState.keyword);
        setType(queryState.type);
        setPage(queryState.page);
        setPageSize(queryState.pageSize);
    };

    return {
        logs, keyword, kind: type, page, pageSize, total: query.data?.total || 0,
        isLoading: query.isFetching || usersQuery.isFetching || deleteMutation.isPending,
        searchLogs: (value = keyword) => updateFilters({ keyword: value }), changeKind: (value: string) => updateFilters({ type: value }), changePage: (value: number) => updateFilters({ page: value }), changePageSize: (value: number) => updateFilters({ pageSize: value }), resetFilters: () => updateFilters({ keyword: "", type: "", page: 1, pageSize: defaultPageSize }), refreshLogs: () => query.refetch(), deleteLog: (id: string) => deleteMutation.mutateAsync(id),
    };
}
