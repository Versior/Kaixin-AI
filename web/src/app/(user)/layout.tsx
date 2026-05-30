"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { AppTopNav } from "@/components/layout/app-top-nav";
import { SiteAnnouncementModal } from "@/components/layout/site-announcement-modal";
import { useUserStore } from "@/stores/use-user-store";

/** 公开页面路径列表（无需登录即可访问） */
const PUBLIC_PATHS = new Set(["/", "/login"]);

function isPublicPath(pathname: string) {
    for (const p of PUBLIC_PATHS) {
        if (pathname === p || pathname.startsWith(p + "?") || pathname.startsWith(p + "#")) return true;
    }
    return false;
}

function LoadingSkeleton() {
    return (
        <div className="flex h-dvh flex-col bg-background text-foreground">
            {/* 顶栏 skeleton */}
            <header className="sticky top-0 z-20 h-16 shrink-0 border-b border-stone-200 bg-background/90 backdrop-blur-xl dark:border-stone-800">
                <div className="mx-auto flex h-full max-w-7xl items-center justify-between px-6">
                    <div className="h-5 w-32 animate-pulse rounded bg-stone-200 dark:bg-stone-800" />
                    <div className="flex items-center gap-2">
                        <div className="size-8 animate-pulse rounded bg-stone-200 dark:bg-stone-800" />
                        <div className="size-8 animate-pulse rounded bg-stone-200 dark:bg-stone-800" />
                        <div className="h-5 w-10 animate-pulse rounded bg-stone-200 dark:bg-stone-800" />
                    </div>
                </div>
            </header>
            {/* 主要区域 skeleton */}
            <div className="flex min-h-0 flex-1 items-center justify-center">
                <div className="flex flex-col items-center gap-4">
                    <div className="size-10 animate-pulse rounded-full bg-stone-200 dark:bg-stone-800" />
                    <div className="h-4 w-48 animate-pulse rounded bg-stone-200 dark:bg-stone-800" />
                    <div className="h-3 w-32 animate-pulse rounded bg-stone-200 dark:bg-stone-800" />
                </div>
            </div>
        </div>
    );
}

export default function UserLayout({ children }: { children: ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();
    const isReady = useUserStore((state) => state.isReady);
    const user = useUserStore((state) => state.user);

    useEffect(() => {
        if (!isReady) return;
        if (user) return;
        if (isPublicPath(pathname)) return;
        const redirect = encodeURIComponent(pathname);
        router.replace(`/login?redirect=${redirect}`);
    }, [isReady, user, pathname, router]);

    // 公开路径：直接渲染（无需登录）
    if (isPublicPath(pathname)) {
        return (
            <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground">
                <AppTopNav />
                <SiteAnnouncementModal />
                <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
            </div>
        );
    }

    // 受保护路径：等待 auth hydrated 后才显示内容
    if (!isReady) {
        return <LoadingSkeleton />;
    }

    return (
        <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground">
            <AppTopNav />
            <SiteAnnouncementModal />
            <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
        </div>
    );
}
