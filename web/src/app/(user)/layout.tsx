"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { AppTopNav } from "@/components/layout/app-top-nav";
import { SiteAnnouncementModal } from "@/components/layout/site-announcement-modal";
import { useUserStore } from "@/stores/use-user-store";

export default function UserLayout({ children }: { children: ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();
    const isReady = useUserStore((state) => state.isReady);
    const user = useUserStore((state) => state.user);

    useEffect(() => {
        if (!isReady) return;
        if (user) return;
        if (pathname.startsWith("/login")) return;
        const redirect = encodeURIComponent(pathname);
        router.replace(`/login?redirect=${redirect}`);
    }, [isReady, user, pathname, router]);

    return (
        <div className="flex h-dvh flex-col overflow-hidden bg-background text-foreground">
            <AppTopNav />
            <SiteAnnouncementModal />
            <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
        </div>
    );
}
