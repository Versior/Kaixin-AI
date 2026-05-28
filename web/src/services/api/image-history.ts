import { apiGet, apiPost, type ApiParams } from "@/services/api/request";
import { useUserStore } from "@/stores/use-user-store";

export type ImageHistoryLog = {
    id: string;
    taskId?: string;
    userId: string;
    username?: string;
    kind: "image" | "chat" | string;
    model: string;
    path: string;
    prompt: string;
    images: string[];
    status: string;
    error?: string;
    createdAt: string;
};

export type ImageHistoryList = {
    items: ImageHistoryLog[];
    total: number;
};

export async function fetchImageHistory(params: ApiParams = {}) {
    return apiGet<ImageHistoryList>("/api/v1/images/history", params, useUserStore.getState().token);
}

export async function deleteImageHistory(ids: string[]) {
    return apiPost<{ success: boolean }>("/api/v1/images/history/batch-delete", { ids }, useUserStore.getState().token);
}
