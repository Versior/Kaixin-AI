import { apiDelete, apiGet, apiPost } from "@/services/api/request";
import { useUserStore } from "@/stores/use-user-store";
import type { CanvasProject } from "@/app/(user)/canvas/stores/use-canvas-store";

export type CloudCanvasProject = {
    id: string;
    userId: string;
    title: string;
    payload: string;
    createdAt: string;
    updatedAt: string;
};

export async function fetchCanvasProjects() {
    return apiGet<CloudCanvasProject[]>("/api/canvas/projects", undefined, useUserStore.getState().token);
}

export async function saveCanvasProject(project: CanvasProject) {
    return apiPost<CloudCanvasProject>(
        "/api/canvas/projects",
        {
            id: project.id,
            title: project.title,
            payload: JSON.stringify(project),
            createdAt: project.createdAt,
            updatedAt: project.updatedAt,
        },
        useUserStore.getState().token,
    );
}

export async function deleteCanvasProject(id: string) {
    return apiDelete<boolean>(`/api/canvas/projects/${encodeURIComponent(id)}`, useUserStore.getState().token);
}

export function parseCloudCanvasProject(item: CloudCanvasProject) {
    return JSON.parse(item.payload) as CanvasProject;
}
