import { apiGet } from "@/services/api/request";
import { useUserStore } from "@/stores/use-user-store";

export type ImageUserRank = {
    userId: string;
    username: string;
    tasks: number;
    images: number;
};

export type ImageStats = {
    totalImages: number;
    todayImages: number;
    successImages: number;
    failedImages: number;
    userRanks: ImageUserRank[];
};

export async function fetchImageStats() {
    return apiGet<ImageStats>("/api/v1/images/stats", undefined, useUserStore.getState().token);
}
