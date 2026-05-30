"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { ProConfigProvider } from "@ant-design/pro-components";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App, ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";

import { ClientRootInit } from "@/components/layout/client-root-init";
import { getAntThemeConfig } from "@/lib/app-theme";
import { useThemeStore } from "@/stores/use-theme-store";

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 30_000,
            retry: false,
            refetchOnWindowFocus: false,
        },
    },
});

export function AppProviders({ children }: { children: ReactNode }) {
    const [hydrated, setHydrated] = useState(false);
    const theme = useThemeStore((state) => state.theme);
    const dark = theme === "dark";

    // 等待 Zustand persist 从 localStorage 恢复主题后再同步到 DOM
    // 避免默认值 "dark" 覆盖内联脚本已设好的正确主题（系统浅色偏好用户必闪）
    useEffect(() => {
        const unsub = useThemeStore.persist.onFinishHydration(() => setHydrated(true));
        if (useThemeStore.persist.hasHydrated()) setHydrated(true);
        return unsub;
    }, []);

    useEffect(() => {
        if (!hydrated) return;
        document.documentElement.classList.toggle("dark", dark);
        document.documentElement.style.colorScheme = theme;
    }, [dark, theme, hydrated]);

    return (
        <ConfigProvider locale={zhCN} theme={getAntThemeConfig(dark)}>
            <ProConfigProvider dark={dark}>
                <App>
                    <QueryClientProvider client={queryClient}>
                        <ClientRootInit>{children}</ClientRootInit>
                    </QueryClientProvider>
                </App>
            </ProConfigProvider>
        </ConfigProvider>
    );
}
