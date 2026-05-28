import { apiGet } from "@/services/api/request";
import { useUserStore } from "@/stores/use-user-store";

export type ImageTaskInfo = {
    id: string;
    userId: string;
    username: string;
    model: string;
    status: "running" | "waiting";
    createdAt: string;
    startedAt?: string;
    estimatedWaitSeconds: number;
};

export type ImageTaskStatus = {
    running: ImageTaskInfo | null;
    waiting: ImageTaskInfo[];
};

export async function fetchImageTaskStatus() {
    return apiGet<ImageTaskStatus>("/api/v1/images/tasks", undefined, useUserStore.getState().token);
}
