import { create } from "zustand";
import { persist, type PersistStorage, type StorageValue } from "zustand/middleware";

import { nanoid } from "nanoid";
import { localForageStorage } from "@/lib/localforage-storage";
import { deleteCanvasProject, fetchCanvasProjects, parseCloudCanvasProject, saveCanvasProject } from "@/services/api/canvas-projects";
import { useUserStore } from "@/stores/use-user-store";
import { userScopedStorageKey } from "@/lib/user-scoped-storage";
import type { CanvasBackgroundMode } from "@/lib/canvas-theme";
import type { CanvasAssistantSession, CanvasConnection, CanvasNodeData, ViewportTransform } from "../types";

export type CanvasProject = {
    id: string;
    title: string;
    createdAt: string;
    updatedAt: string;
    nodes: CanvasNodeData[];
    connections: CanvasConnection[];
    chatSessions: CanvasAssistantSession[];
    activeChatId: string | null;
    backgroundMode: CanvasBackgroundMode;
    showImageInfo: boolean;
    viewport: ViewportTransform;
};

type CanvasStore = {
    hydrated: boolean;
    cloudHydratedUserId: string | null;
    projects: CanvasProject[];
    createProject: (title?: string) => string;
    importProject: (project: Partial<CanvasProject>) => string;
    openProject: (id: string) => CanvasProject | null;
    renameProject: (id: string, title: string) => void;
    deleteProjects: (ids: string[]) => void;
    updateProject: (id: string, patch: Partial<Pick<CanvasProject, "nodes" | "connections" | "chatSessions" | "activeChatId" | "backgroundMode" | "showImageInfo" | "viewport">>) => void;
    hydrateCloudProjects: () => Promise<void>;
};

const initialViewport: ViewportTransform = { x: 0, y: 0, k: 1 };
const CANVAS_STORE_KEY = "infinite-canvas:canvas_store";
type PersistedCanvasState = Pick<CanvasStore, "projects">;
let saveTimer: ReturnType<typeof setTimeout> | null = null;
let queuedPersistState: PersistedCanvasState | null = null;
const cloudSaveTimers = new Map<string, ReturnType<typeof setTimeout>>();

function mergeProjects(localProjects: CanvasProject[], cloudProjects: CanvasProject[]) {
    const projects = new Map<string, CanvasProject>();
    [...cloudProjects, ...localProjects].forEach((project) => {
        const existing = projects.get(project.id);
        if (!existing || new Date(project.updatedAt).getTime() >= new Date(existing.updatedAt).getTime()) projects.set(project.id, project);
    });
    return Array.from(projects.values()).sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
}

function queueCloudSave(project: CanvasProject) {
    const { token, user } = useUserStore.getState();
    if (!token || !user || user.role === "guest") return;
    const existing = cloudSaveTimers.get(project.id);
    if (existing) clearTimeout(existing);
    cloudSaveTimers.set(
        project.id,
        setTimeout(() => {
            cloudSaveTimers.delete(project.id);
            void saveCanvasProject(project).catch((error) => console.error("Failed to sync canvas project", project.id, error));
        }, 800),
    );
}

function queueCloudDelete(ids: string[]) {
    const { token, user } = useUserStore.getState();
    if (!token || !user || user.role === "guest") return;
    ids.forEach((id) => {
        const existing = cloudSaveTimers.get(id);
        if (existing) clearTimeout(existing);
        cloudSaveTimers.delete(id);
        void deleteCanvasProject(id).catch((error) => console.error("Failed to delete cloud canvas project", id, error));
    });
}

const canvasStorage: PersistStorage<CanvasStore> = {
    getItem: async (name) => {
        const value = await localForageStorage.getItem(userScopedStorageKey(name));
        if (!value) return null;
        const parsed = JSON.parse(value) as StorageValue<CanvasStore>;
        queuedPersistState = parsed.state as PersistedCanvasState;
        return parsed;
    },
    setItem: (name, value) => {
        const nextState = value.state as PersistedCanvasState;
        if (queuedPersistState && queuedPersistState.projects === nextState.projects) return;
        queuedPersistState = nextState;
        if (saveTimer) clearTimeout(saveTimer);
        saveTimer = setTimeout(() => {
            saveTimer = null;
            void localForageStorage.setItem(userScopedStorageKey(name), JSON.stringify(value));
        }, 400);
    },
    removeItem: (name) => localForageStorage.removeItem(userScopedStorageKey(name)),
};

export const useCanvasStore = create<CanvasStore>()(
    persist(
        (set, get) => ({
            hydrated: false,
            cloudHydratedUserId: null,
            projects: [],
            createProject: (title = "未命名灵感") => {
                const now = new Date().toISOString();
                const id = nanoid();
                const project: CanvasProject = {
                    id,
                    title,
                    createdAt: now,
                    updatedAt: now,
                    nodes: [],
                    connections: [],
                    chatSessions: [],
                    activeChatId: null,
                    backgroundMode: "lines",
                    showImageInfo: false,
                    viewport: initialViewport,
                };
                queueCloudSave(project);
                set((state) => ({ projects: [project, ...state.projects] }));
                return id;
            },
            importProject: (source) => {
                const now = new Date().toISOString();
                const project: CanvasProject = {
                    id: nanoid(),
                    title: source.title || "导入灵感",
                    createdAt: source.createdAt || now,
                    updatedAt: now,
                    nodes: source.nodes || [],
                    connections: source.connections || [],
                    chatSessions: source.chatSessions || [],
                    activeChatId: source.activeChatId || null,
                    backgroundMode: source.backgroundMode || "lines",
                    showImageInfo: source.showImageInfo || false,
                    viewport: source.viewport || initialViewport,
                };
                queueCloudSave(project);
                set((state) => ({ projects: [project, ...state.projects] }));
                return project.id;
            },
            openProject: (id) => {
                return get().projects.find((item) => item.id === id) || null;
            },
            renameProject: (id, title) =>
                set((state) => {
                    let updatedProject: CanvasProject | null = null;
                    const projects = state.projects.map((project) => {
                        if (project.id !== id) return project;
                        updatedProject = { ...project, title: title.trim() || project.title, updatedAt: new Date().toISOString() };
                        return updatedProject;
                    });
                    if (updatedProject) queueCloudSave(updatedProject);
                    return { projects };
                }),
            deleteProjects: (ids) =>
                set((state) => {
                    queueCloudDelete(ids);
                    const projects = state.projects.filter((project) => !ids.includes(project.id));
                    return { projects };
                }),
            updateProject: (id, patch) =>
                set((state) => {
                    let updatedProject: CanvasProject | null = null;
                    const projects = state.projects.map((project) => {
                        if (project.id !== id) return project;
                        updatedProject = { ...project, ...patch, updatedAt: new Date().toISOString() };
                        return updatedProject;
                    });
                    if (updatedProject) queueCloudSave(updatedProject);
                    return { projects };
                }),
            hydrateCloudProjects: async () => {
                const { token, user } = useUserStore.getState();
                if (!token || !user || user.role === "guest") {
                    set({ cloudHydratedUserId: null });
                    return;
                }
                if (get().cloudHydratedUserId === user.id) return;
                const cloudProjects = (await fetchCanvasProjects()).map(parseCloudCanvasProject);
                set((state) => ({ projects: mergeProjects(state.projects, cloudProjects), cloudHydratedUserId: user.id }));
            },
        }),
        {
            name: CANVAS_STORE_KEY,
            storage: canvasStorage,
            partialize: (state) =>
                ({
                    projects: state.projects,
                }) as StorageValue<CanvasStore>["state"],
            onRehydrateStorage: () => () => {
                useCanvasStore.setState({ hydrated: true });
            },
        },
    ),
);
